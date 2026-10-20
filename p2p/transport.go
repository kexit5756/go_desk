package p2p

// Peer 表示远程节点的抽象，目前还没有规定节点必须实现哪些方法。
// 这是空接口：任何类型都能满足它，因此暂时不能通过 Peer 调用收发方法。
type Peer interface{}

// Transport 描述传输层需要具备的能力，业务层可以面向这个接口编写代码。
// Go 按方法集合隐式判断实现关系，类型不需要声明 implements。
// 当前 *TcpTransport 已具备这里要求的 ListenAndAccept 方法。
type Transport interface {
	// ListenAndAccept 启动监听与连接接收；启动失败时返回错误。
	ListenAndAccept() error
}
