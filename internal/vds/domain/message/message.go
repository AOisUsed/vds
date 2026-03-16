// Package message 包括了对消息的定义
package message

// Message 传输的消息
type Message struct {
	SrcID   string
	DstID   string
	Payload []byte
}
