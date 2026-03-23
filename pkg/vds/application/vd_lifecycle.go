package application

import "context"

type VDLifeCycleService interface {
	StartDevice(ctx context.Context, in *StartVDRequest) error
	StopDevice(ctx context.Context, in *StopVDRequest) error
}
