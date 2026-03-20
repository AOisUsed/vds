package application

import "context"

type VDLifeCycleService interface {
	startDevice(ctx context.Context, in *StartVDRequest) error
	stopDevice(ctx context.Context, in *StopVDRequest) error
}
