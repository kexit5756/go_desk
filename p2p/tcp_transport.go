package p2p

import (
	"errors"
	"fmt"
	"net"
	"sync"
)

// TCPPeer 封装一条已经建立的 TCP 连接，以及连接的发起方向。
// 它表示一个远程连接；TcpTransport 则负责监听和管理通信过程。
type TCPPeer struct {
	// conn 可用于读、写、关闭连接和获取远端地址。
	// 这里是命名字段，需要通过 peer.conn 访问；并不是嵌入 net.Conn。
	conn net.Conn

	// outbound 为 true 表示本节点主动 Dial，为 false 表示通过 Accept 接入。
	// 这个标记由程序设置，不限制收发方向；TCP 连接两端都可以读写。
	outbound bool
}

// TcpTransport 是基于 TCP 的传输层实现。
type TcpTransport struct {
	ListenAddress string       // 监听配置，例如 ":3000"。
	Listener      net.Listener // net.Listen 成功后得到的监听器，负责接收连接。

	// 预留用于保护 peers 的读写锁；当前还没有保存和访问节点的逻辑。
	// 声明锁不会自动保护 map，实际访问时需要主动调用 Lock/RLock 等方法。
	mu sync.RWMutex
	// 预留的远程节点表，目前是 nil map；写入前需要用 make 初始化。
	peers map[net.Addr]Peer
}

// NewTCCPer 创建 TCPPeer，并保存连接及方向标记。
// TODO: 当前函数名有拼写问题，可在后续统一重命名为 NewTCPPeer 并更新调用处。
func NewTCCPer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
	}
}

// NewTCPTransport 创建对象并保存监听配置，不在构造时建立网络监听。
func NewTCPTransport(listenAddr string) *TcpTransport {
	return &TcpTransport{
		ListenAddress: listenAddr,
	}
}

// ListenAndAccept 创建 TCP 监听器，并异步启动连接接收循环。
// 返回 nil 代表监听启动成功，不代表已经有客户端连接或消息传输成功。
func (t *TcpTransport) ListenAndAccept() error {
	var err error
	// Listener 是已有结构体字段，err 也已声明，因此这里使用 = 赋值。
	t.Listener, err = net.Listen("tcp", t.ListenAddress)
	if err != nil {
		return err
	}
	go t.acceptLoop()
	return nil
}

// acceptLoop 持续等待客户端连接；没有新连接时 Accept 会阻塞等待。
func (t *TcpTransport) acceptLoop() {
	for {
		conn, err := t.Listener.Accept()
		if err != nil {
			// 主动关闭监听器也会让 Accept 返回错误，此时退出循环即可。
			if errors.Is(err, net.ErrClosed) {
				return
			}
			fmt.Printf("Tcp accept error :%s\n", err)
			// 接收失败时不处理 conn，继续尝试接收下一条连接。
			continue
		}

		// 每条连接交给独立 goroutine 处理，当前循环继续接收其他连接。
		go t.handleConn(conn)
	}
}

// handleConn 当前只封装连接并打印日志，尚未实现读取、响应和节点登记。
func (t *TcpTransport) handleConn(conn net.Conn) {
	// TODO: 此函数目前仅从 Accept 路径调用，属于入站连接，方向应为 false。
	// 本次只添加说明，保留当前 true 的执行逻辑，便于后续单独修复和验证。
	peer := NewTCCPer(conn, true)
	fmt.Printf("new incoming connetion %+v\n", peer)

	// TODO: 后续实现连接处理时明确由谁关闭 conn；当前处理结束后没有关闭它，
	// 也没有把 peer 存入节点表，连接的完整生命周期尚未实现。
}
