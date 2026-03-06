package mock

import (
	"context"
	"virturalDevice/connection"
	"virturalDevice/vds"
	"virturalDevice/vds/virtual_device"
)

// 测试用模拟vdRepo
type VDRepo struct {
	addressById map[string]connection.Connection
}

func (m VDRepo) GetVDStateById(ctx context.Context, id string) (virtual_device.Flag, error) {
	//TODO implement me
	panic("implement me")
}

func (m VDRepo) GetAllVDStates(ctx context.Context) (map[string]virtual_device.Flag, error) {
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

func NewMockVDRepository() vds.VDRepository {
	return &VDRepo{}
}
