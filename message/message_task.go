package message

import "context"

// Task 包含上下文的消息传输任务，用于应对消息取消发送的场景
type Task struct {
	Ctx     context.Context
	Message Message
}
