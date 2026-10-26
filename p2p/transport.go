package p2p

// Peer 定义远程节点当前需要提供的能力。
// 它已不是空接口：实现类型必须有 Close() error 方法；*TCPPeer 满足该要求。
type Peer interface {
	// Close 关闭节点底层连接，用于拒绝节点或回收资源。
	Close() error
}

// Transport 抽象传输层的启动与消息消费能力。
// Go 根据方法集合隐式判断实现关系，*TcpTransport 实现了这里的两个方法。
type Transport interface {
	// ListenAndAccept 启动监听及后台接收循环；返回 nil 只代表监听启动成功。
	ListenAndAccept() error
	// Consume 返回只读 channel，业务层可接收消息，不能通过该返回值发送消息。
	Consume() <-chan *RPC
}
