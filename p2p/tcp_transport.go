package p2p

import (
	"fmt"
	"net"
	"sync"
)

// TCPPeer 封装一条已经建立的 TCP 连接，以及连接的发起方向。
// 它表示远程节点的一条连接，TcpTransport 则负责监听和消息交付。
type TCPPeer struct {
	// 命名字段，不是嵌入 net.Conn；调用连接方法需要使用 p.conn。
	conn net.Conn
	// outbound 为 true 表示主动 Dial，为 false 表示通过 Accept 接入。
	// 这是本节点记录的方向标记，TCP 连接本身仍然可以双向读写。
	outbound bool
}

// TCPTransportOpts 将可替换的握手、解码和节点回调与监听地址放在一起。
// 当前构造函数不校验必填项，启动收发前应提供非 nil 的 HandshakeFunc 和 Decoder。
type TCPTransportOpts struct {
	ListenAddr    string
	HandshakeFunc HandshakeFunc
	Decoder       Decoder
	// OnPeer 在握手成功后调用；为 nil 时跳过，返回错误时结束当前连接处理。
	OnPeer func(Peer) error
}

// Close 使 *TCPPeer 满足 Peer 接口，并把关闭操作委托给底层连接。
func (p *TCPPeer) Close() error {
	return p.conn.Close()
}

// TcpTransport 保存配置、监听器以及交付给业务层的消息 channel。
type TcpTransport struct {
	TCPTransportOpts
	Listener net.Listener
	// rpcch 在构造时初始化；当前无缓冲，发送会等待消费者接收。
	rpcch chan *RPC
	// mu 预留用于保护 peers；声明锁不会自动保护 map，当前尚无节点登记逻辑。
	mu sync.RWMutex
	// peers 当前为 nil map；写入前需要 make，并约定加锁规则。
	peers map[net.Addr]Peer
}

// NewTCCPer 创建 TCPPeer，保存连接和方向标记。
// TODO: 后续统一重命名为 NewTCPPeer，并更新调用处。
func NewTCCPer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
	}
}

// NewTCPTransport 保存配置并创建消息 channel，此时尚未监听端口。
func NewTCPTransport(opts TCPTransportOpts) *TcpTransport {
	return &TcpTransport{
		TCPTransportOpts: opts,
		// 没有第二个容量参数，因此不是缓冲 channel。
		// 若不 make，channel 的零值为 nil，发送和接收都会永久阻塞。
		rpcch: make(chan *RPC),
	}
}

// ListenAndAccept 创建监听器，并异步启动接收循环。
// 返回 nil 不代表已有客户端连接，也不代表消息已经成功交付。
func (t *TcpTransport) ListenAndAccept() error {
	var err error
	// Listener 是已有字段，err 也已声明，因此这里使用 = 赋值。
	t.Listener, err = net.Listen("tcp", t.TCPTransportOpts.ListenAddr)
	if err != nil {
		return err
	}
	go t.acceptLoop()
	return nil
}

// Consume 把内部 channel 以只读形式交给业务层，不会创建或复制 channel。
func (t *TcpTransport) Consume() <-chan *RPC {
	return t.rpcch
}

// acceptLoop 阻塞等待新连接，每条成功接入的连接交给独立 goroutine。
func (t *TcpTransport) acceptLoop() {
	for {
		conn, err := t.Listener.Accept()
		if err != nil {
			// TODO: 此处当前只有日志，错误后仍会处理 conn，可能传入 nil。
			// 应在监听器关闭时退出，对其他错误分类处理，成功后才启动 handleConn。
			fmt.Printf("Tcp accept error :%s\n", err)
		}
		// TODO: 用 conn.RemoteAddr().String() 打印地址；%s 不适合直接打印连接对象。
		fmt.Printf("Tcp accept new connection from %s\n", conn)
		go t.handleConn(conn)
	}
}

// handleConn 依次执行握手、节点回调和消息解码，最后通过 channel 交付 RPC。
// 函数退出时关闭连接；若卡在 channel 发送处，则还不会执行 defer。
func (t *TcpTransport) handleConn(conn net.Conn) {
	var err error
	// defer 在函数返回或栈展开时执行，用于关闭底层连接。
	// TODO: 下面循环里的 err := 会遮蔽外层 err，因此这里可能打印不到真实读取错误。
	defer func() {
		fmt.Printf("dropping peer conenction:%s", err)
		conn.Close()
	}()
	// TODO: 当前只从 Accept 接入，outbound 应为 false；这里仍保留现有 true。
	peer := NewTCCPer(conn, true)
	if err = t.TCPTransportOpts.HandshakeFunc(peer); err != nil {
		return
	}
	// 回调可以登记或拒绝节点；当前 main 的 Onpeer 会立即关闭此连接。
	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil {
			return
		}
	}
	// TODO: 应移到循环内，让每条消息独立分配；复用指针会覆盖消费者持有的数据。
	// channel 传递的是指针，发送结束不代表消费者已经完成对对象的读取。
	rpc := &RPC{}
	for {
		err := t.Decoder.Decoder(conn, rpc)
		// TODO: net.Conn 的错误可能被包装；判断关闭原因时应考虑 errors.Is。
		if err == net.ErrClosed {
			return
		}
		if err != nil {
			fmt.Printf("Tcp decode error:%s\n", err)
			return
		}
		rpc.From = conn.RemoteAddr()
		// 无消费者时这里会阻塞；后续需要设计取消与退出机制。
		t.rpcch <- rpc
	}
}
