package application

import "context"

type VDSLifeCycleService interface {
	StartVDS(ctx context.Context, in *StartVDSRequest) (*StartVDSResponse, error)
	StopVDS(ctx context.Context, in *StopVDSRequest) (*StopVDSResponse, error)
}

type VDMessageService interface {
	SendMessage(ctx context.Context, in *SendMessageRequest) (*SendMessageResponse, error)
	CancelMessage(ctx context.Context, in *CancelMessageRequest) (*CancelMessageResponse, error)
}

type VDLifeCycleService interface {
}
