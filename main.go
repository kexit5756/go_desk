package main

import (
	"log"

	"github.com/kexit5756/go_desk/p2p"
)

func main() {
	// 创建 TCP 传输对象，只保存配置；此时还没有监听端口。
	// ":3000" 省略主机部分，表示监听本机可用网络接口的 3000 端口。
	tr := p2p.NewTCPTransport(":3000")

	// 尝试创建监听器，并在后台启动接收连接的循环。
	// 端口占用、地址格式错误等情况会通过 error 返回。
	if err := tr.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}

	// 空 select 永久阻塞当前 goroutine，让后台监听循环能够持续运行。
	// 当前用它保持进程存活；后续可以学习通过系统信号实现正常退出。
	select {}
}
