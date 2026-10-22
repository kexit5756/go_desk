package main

import (
	"log"

	"github.com/kexit5756/go_desk/p2p"
)

func main() {
	tcpOpts := p2p.TCPTransportOpts{
		ListenAddr:    ":3000",
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{}, // 使用默认解码器。
	}
	tr := p2p.NewTCPTransport(tcpOpts)
	if err := tr.ListenAndAccept(); err != nil {
		log.Fatalf("Failed to start TCP transport: %v", err)
	}

	select {}
}

// 创建 TCP 传输对象，只保存配置；此时还没有监听端口。
// ":3000" 省略主机部分，表示监听本机可用网络接口的 3000 端口。
// 阻塞主 goroutine，让后台监听循环持续运行。

// 空 select 永久阻塞当前 goroutine，让后台监听循环能够持续运行。
// 当前用它保持进程存活；后续可以学习通过系统信号实现正常退出。
