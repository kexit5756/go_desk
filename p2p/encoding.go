package p2p

import (
	"encoding/gob"
	"io"
)

type Decoder interface {
	Decoder(io.Reader, *Message) error
}

type GOBDecoder struct{}

func (dec GOBDecoder) Decoder(r io.Reader, msg *Message) error {
	return gob.NewDecoder(r).Decode(msg)
}

type DefaultDecoder struct{}

func (dec DefaultDecoder) Decoder(r io.Reader, msg *Message) error {
	buf := make([]byte, 1028)
	n, err := r.Read(buf)
	if err != nil {
		return err
	}
	msg.PayLoad = buf[:n]
	return nil
}
