package mock

import (
	"context"
	"virturalDevice/pkg/connection"
	"virturalDevice/pkg/vds/radiomodel"
	"virturalDevice/pkg/vds/vdrepository"
)

// 测试用模拟vdRepo
type Repository struct {
	addressById map[string]connection.Connection
}

func (m Repository) GetVDStateById(ctx context.Context, id string) (radiomodel.Params, error) {
	//TODO implement me
	panic("implement me")
}

func (m Repository) GetAllVDStates(ctx context.Context) (map[string]radiomodel.Params, error) {
	//TODO implement me
	panic("implement me")
}

func (m Repository) GetVDConnById(ctx context.Context, id string) (connection.Connection, error) {
	//TODO implement me
	panic("implement me")
}

func (m Repository) SetVDConnById(ctx context.Context, id string, address connection.Connection) error {
	//TODO implement me
	panic("implement me")
}

func NewMockVDRepository() vdrepository.VDRepository {
	return &Repository{}
}
