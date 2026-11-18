package main

import (
	"log"
	"time"

	"github.com/kexit5756/go_desk/p2p"
)

func main() {
	// TCP 传输层负责监听端口，并通过 Consume channel 将解码后的消息交给文件服务器。
	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddr:    ":3000",
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
		// OnPeer 回调可选；不设置时，接入的节点会继续进入消息解码流程。
	}

	tcpTransport := p2p.NewTCPTransport(tcpTransportOpts)

	fileServerOpts := FileServeropts{
		StorageRoot:       "3000_gnetwork",
		PathTransformFunc: CASPathTransformFunc,
		Transport:         tcpTransport,
	}
	s := NewFileServer(fileServerOpts)
	// 当前示例会在 4 秒后自动停止；长时间运行时可移除这个定时器。
	go func() {
		time.Sleep(4 * time.Second)
		s.Stop()
	}()

	// Start 会阻塞在服务循环中；Stop 发出退出信号后，Start 才会返回。
	if err := s.Start(); err != nil {
		log.Fatalf("Failed to start file server: %v", err)
	}
}
