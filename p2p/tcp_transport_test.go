package p2p

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestTCPTransport 创建配置并尝试启动监听，不会建立客户端连接。
// 当前不覆盖握手、OnPeer、Consume、数据内容和连接退出行为。
func TestTCPTransport(t *testing.T) {
	// 固定使用 3000 端口：若 make run 启动的服务仍在运行，测试会监听失败。
	// TODO: 改为 "127.0.0.1:0"，由系统分配空闲端口。
	opts := TCPTransportOpts{
		ListenAddr:    ":3000",
		HandshakeFunc: NOPHandshakeFunc,
		Decoder:       DefaultDecoder{},
	}
	transport := NewTCPTransport(opts)
	// TODO: 当前只检查 opts 中写入的常量，没有验证 transport 是否保存配置。
	// 后续应检查 transport.ListenAddr，并保持期望值在前、实际值在后。
	assert.Equal(t, opts.ListenAddr, ":3000")
	assert.Nil(t, transport.ListenAndAccept())
	// TODO: 先完善 acceptLoop 的关闭处理，再通过 t.Cleanup 释放监听器。
}

// 待补充：端口被占用时，应验证监听错误能传递给调用者。
// 目前这里只是后续任务说明，没有对应的可执行测试函数。
