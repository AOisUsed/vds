package application

import "context"

type VDMessageService interface {
	SendMessage(ctx context.Context, in *SendMessageRequest) error
	CancelMessage(ctx context.Context, in *CancelMessageRequest) error
	SubscribeMessage(ctx context.Context, in *SubscribeDeviceMessageRequest) error
}
