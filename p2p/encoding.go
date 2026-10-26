package p2p

import (
	"encoding/gob"
	"io"
)

// Decoder 从字节流读取数据，并把解码结果写入传入的 RPC。
// 接口允许替换读取策略，但发送方的数据格式仍必须与解码方式一致。
type Decoder interface {
	Decoder(io.Reader, *RPC) error
}

// GOBDecoder 用于读取 Go Gob 格式的数据，不能直接解析 nc 输入的普通文本。
type GOBDecoder struct{}

func (dec GOBDecoder) Decoder(r io.Reader, msg *RPC) error {
	// TODO: 当前每次新建 Decoder，会丢失连续 Gob 消息所需的类型和缓冲状态。
	// 后续应改为每条连接独立持有并复用一个 gob.Decoder。
	return gob.NewDecoder(r).Decode(msg)
}

// DefaultDecoder 直接读取原始字节，适合当前使用 nc 的基础收发练习。
// 它尚未定义消息边界：一次 TCP Read 不保证对应发送方的一整条消息。
type DefaultDecoder struct{}

func (dec DefaultDecoder) Decoder(r io.Reader, msg *RPC) error {
	// 本次最多读取 1028 字节，只保留实际读取的 n 字节，避免把尾部零值当作内容。
	buf := make([]byte, 1028)
	n, err := r.Read(buf)
	// TODO: io.Reader 允许同时返回 n > 0 和 err；当前会忽略这部分有效数据。
	if err != nil {
		return err
	}
	msg.PayLoad = buf[:n]
	return nil
}
