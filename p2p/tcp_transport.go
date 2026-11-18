package p2p

import (
	"errors"
	"io"
	"log"
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

	log.Printf("TCP transport listening on %s", t.ListenAddr)
	return nil

}

// Close 关闭监听器，停止接收新连接；已接入的连接仍然可以读写。
func (t *TcpTransport) Close() error {
	if t.Listener == nil {
		return nil
	}
	return t.Listener.Close()
}

// Consume 把内部 channel 以只读形式交给业务层，不会创建或复制 channel。
func (t *TcpTransport) Consume() <-chan *RPC {
	return t.rpcch
}

// acceptLoop 阻塞等待新连接，每条成功接入的连接交给独立 goroutine。
func (t *TcpTransport) acceptLoop() {
	for {
		conn, err := t.Listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			log.Println("TCP Transport listener closed, exiting accept loop")
			return
		}
		if err != nil {
			log.Printf("TCP accept error: %v", err)
			continue
		}
		log.Printf("TCP accepted connection from %s", conn.RemoteAddr())
		go t.handleConn(conn)
	}
}

// handleConn 依次执行握手、节点回调和消息解码，最后通过 channel 交付 RPC。
// 函数退出时关闭连接；若卡在 channel 发送处，则还不会执行 defer。
func (t *TcpTransport) handleConn(conn net.Conn) {
	var err error
	defer func() {
		if err != nil && !errors.Is(err, net.ErrClosed) && !errors.Is(err, io.EOF) {
			log.Printf("dropping peer connection after error: %v", err)
		} else {
			log.Printf("closing peer connection from %s", conn.RemoteAddr())
		}
		_ = conn.Close()
	}()
	// 该连接由 Accept 接入，因此对于本节点来说是入站连接。
	peer := NewTCCPer(conn, false)
	if err = t.TCPTransportOpts.HandshakeFunc(peer); err != nil {
		return
	}
	// 回调可以登记或拒绝节点；当前 main 的 Onpeer 会立即关闭此连接。
	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil {
			return
		}
	}
	for {
		// 每条消息创建独立的 RPC，避免后续解码覆盖消费者已从 channel 收到的数据。
		rpc := &RPC{}
		err = t.Decoder.Decoder(conn, rpc)
		if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) {
			return
		}
		if err != nil {
			log.Printf("TCP decode error: %v", err)
			return
		}
		rpc.From = conn.RemoteAddr()
		// 无消费者时这里会阻塞；后续需要设计取消与退出机制。
		t.rpcch <- rpc
	}
}
