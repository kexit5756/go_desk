package p2p

// HandshakeFunc 在开始读取业务消息前检查 Peer，返回错误时结束连接处理。
// 不同连接可能并发调用该函数，共享状态需要自行同步。
type HandshakeFunc func(Peer) error

// NOPHandshakeFunc 是空握手，直接返回成功，不执行认证或协议协商。
func NOPHandshakeFunc(Peer) error {
	return nil
}
