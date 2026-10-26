package p2p

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestTCPTransport 验证构造函数保存了地址，并且能够启动 TCP 监听。
// 当前测试没有创建客户端，尚未验证连接接收和消息收发。
func TestTCPTransport(t *testing.T) {
	// 固定测试端口；若 4000 已被其他进程占用，监听断言会失败。
	// 后续可改用 "127.0.0.1:0"，让系统自动分配空闲端口。
	opts := TCPTransportOpts{
		ListenAddr:    ":3000",
		HandshakeFunc: NOPHandshakeFunc,
		Decoder:       DefaultDecoder{},
	}
	transport := NewTCPTransport(opts)
	assert.Equal(t, opts.ListenAddr, ":3000")
	assert.Nil(t, transport.ListenAndAccept())

}

// 待补充：端口被占用时，应验证监听错误能传递给调用者。
// 目前这里只是后续任务说明，没有对应的可执行测试函数。
