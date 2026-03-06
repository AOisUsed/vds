package message

import "context"

// Task 包含上下文的消息传输任务，用于应对任务取消的场景
type Task struct {
	Ctx     context.Context
	Message Message
}
