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
	listenAddr := ":4000"
	tr := NewTCPTransport(listenAddr)

	// 检查保存的配置。Equal 的惯例是期望值在前、实际值在后；
	// 当前顺序不影响相等判断，但失败报告中的 expected/actual 会颠倒。
	assert.Equal(t, tr.ListenAddress, listenAddr)

	// 返回 nil 表示启动没有错误。检查 error 时也可以使用 assert.NoError。
	assert.Nil(t, tr.ListenAndAccept())

	// TODO: 后续加入 t.Cleanup，在成功启动后关闭监听器，回收测试资源。
}

// 待补充：端口被占用时，应验证监听错误能传递给调用者。
// 目前这里只是后续任务说明，没有对应的可执行测试函数。
