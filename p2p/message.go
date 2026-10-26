package p2p

import (
	"net"
)

// RPC 是当前传输层使用的消息信封，保存来源与原始字节内容。
// 当前尚无方法名、请求编号或响应关联，因此还不是完整的远程方法调用协议。
type RPC struct {
	From    net.Addr // 接收端通过 conn.RemoteAddr() 设置的发送方地址。
	PayLoad []byte   // 实际读取的字节；打印文字可以使用 %s 或 string(PayLoad)。
}
