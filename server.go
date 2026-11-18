package main

import (
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/kexit5756/go_desk/p2p"
)

type FileServeropts struct {
	// ListenAddr 和 TcpTransportOpts 暂留作后续传输层配置；当前启动的是下面
	// 由调用方传入的 Transport 实例。
	ListenAddr        string
	StorageRoot       string
	PathTransformFunc PathTransformFunc
	Transport         p2p.Transport
	TcpTransportOpts  p2p.TCPTransportOpts
}

type FileServer struct {
	FileServeropts
	store    *Store
	quitch   chan struct{}
	stopOnce sync.Once
}

// NewFileServer 创建文件服务器及其本地存储，但不会立即启动网络监听。
func NewFileServer(opts FileServeropts) *FileServer {
	StoreOpts := StoreOpts{
		Root:              opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	}
	return &FileServer{
		store:          NewStore(StoreOpts),
		FileServeropts: opts,
		quitch:         make(chan struct{}),
	}
}

// Stop 通知事件循环退出；sync.Once 保证重复调用不会重复关闭 channel。
func (s *FileServer) Stop() {
	s.stopOnce.Do(func() { close(s.quitch) })
}

// loop 从传输层接收已解码的消息，并交给当前的消息处理逻辑。
func (s *FileServer) loop() {
	defer func() {
		log.Println("FileServer stopped user quit action")
		s.Transport.Close()
	}()
	for {
		select {
		case msg := <-s.Transport.Consume():
			fmt.Println(msg)
		case <-s.quitch:
			return

		}
	}
}

// Start 启动监听并处理消息，直到调用 Stop。
func (s *FileServer) Start() error {
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}
	s.loop()
	return nil

}

// Store 将 reader 中的数据保存到 id 和 key 指定的位置。
func (s *FileServer) Store(id string, key string, r io.Reader) error {
	return s.store.Write(id, key, r)
}
