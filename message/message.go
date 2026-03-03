package message

import "bytes"

type Message struct {
	SrcID string
	DstID string
	Body  []byte
}

// Byte 将消息转化为byte
func (m Message) Byte() []byte {
	buf := new(bytes.Buffer)
	buf.Write([]byte(m.SrcID))
	buf.Write([]byte(m.DstID))
	buf.Write(m.Body)
	return buf.Bytes()
}
