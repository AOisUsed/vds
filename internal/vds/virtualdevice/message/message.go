// Package message 包括了对消息的定义
package message

import (
	"bytes"
)

// Message 传输的消息
type Message struct {
	SrcID   string
	DstID   string
	Payload []byte
}

// Bytes 将消息转化为byte
func (m Message) Bytes() []byte {
	buf := new(bytes.Buffer)
	buf.Write([]byte(m.SrcID))
	buf.Write([]byte(m.DstID))
	buf.Write(m.Payload)
	return buf.Bytes()
}
