package vds

import (
	"virturalDevice/connection"
	"virturalDevice/message"
)

// Sender vds内统一消息发送器
type Sender interface {
	Send(dst connection.Connection, message message.Message) error
}
