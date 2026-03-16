// Package message 包括了对消息的定义
package message

// Message 传输的消息
type Message struct {
	SrcID   string // 消息来源 id
	DstID   string // 消息目标 id, 为空则认定为广播
	Payload []byte // 消息内容
}
