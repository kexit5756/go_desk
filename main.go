package main

import (
	"fmt"
	"log"

	"github.com/kexit5756/go_desk/p2p"
)

// Onpeer 是握手成功后调用的示例回调。
// 注意：当前实现会立即关闭连接，随后解码会失败，无法继续接收客户端消息。
// TODO: 练习消息接收时，应让回调保留连接；需要拒绝节点时再关闭，并处理 Close 的错误。
func Onpeer(peer p2p.Peer) error {
	peer.Close()
	return nil
}

func main() {
	// 创建配置不会立即监听；ListenAndAccept 才会调用 net.Listen。
	// ":3000" 省略主机地址，表示监听可用本地网络接口的 3000 端口。
	tcpOpts := p2p.TCPTransportOpts{
		ListenAddr:    ":3000",
		HandshakeFunc: p2p.NOPHandshakeFunc, // 空握手：直接返回成功，不执行身份校验。
		Decoder:       p2p.DefaultDecoder{}, // 默认读取原始字节，可与 nc 发送的文本配合。
		OnPeer:        Onpeer,               // 当前回调会关闭新连接，接收演示前需先调整。
	}
	tr := p2p.NewTCPTransport(tcpOpts)

	// 业务层独立消费消息，传输层通过 channel 交付解码后的 RPC。
	// 当前 channel 无缓冲，发送方需要等这里接收后，才能继续下一次读取。
	go func() {
		for {
			msg := <-tr.Consume()
			fmt.Printf("Received message: %+v\n", msg)
		}
		// TODO: 后续若实现关闭 channel，需要检查接收的 ok 值或使用 range，避免空转。
	}()
	if err := tr.ListenAndAccept(); err != nil {
		log.Fatalf("Failed to start TCP transport: %v", err)
	}

	// 阻塞主 goroutine，避免 main 返回导致整个程序退出；这里不会持续忙等。
	// 当前可用 Control+C 停止；系统信号处理和正常退出仍待实现。
	select {}
}
