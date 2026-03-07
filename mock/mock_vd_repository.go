package mock

import (
	"context"
	"virturalDevice/connection"
	"virturalDevice/vds/vdrepository"
	"virturalDevice/vds/virtualdevice"
)

// 测试用模拟vdRepo
type VDRepo struct {
	addressById map[string]connection.Connection
}

func (m VDRepo) GetVDStateById(ctx context.Context, id string) (virtualdevice.Flag, error) {
	//TODO implement me
	panic("implement me")
}

func (m VDRepo) GetAllVDStates(ctx context.Context) (map[string]virtualdevice.Flag, error) {
	//TODO implement me
	panic("implement me")
}

func (m VDRepo) GetVDConnById(ctx context.Context, id string) (connection.Connection, error) {
	//TODO implement me
	panic("implement me")
}

func (m VDRepo) SetVDConnById(ctx context.Context, id string, address connection.Connection) error {
	//TODO implement me
	panic("implement me")
}

func NewMockVDRepository() vdrepository.VDRepository {
	return &VDRepo{}
}
