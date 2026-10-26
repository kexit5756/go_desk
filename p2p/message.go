package p2p

import (
	"net"
)

// 支持传输任意类型的数据
// 两个节点的传输都可以
type RPC struct {
	From    net.Addr
	PayLoad []byte
}
