package mock

import (
	"context"
	"virturalDevice/vds/connection"
	"virturalDevice/vds/repository"
	"virturalDevice/vds/virtual_device"
)

// 测试用模拟vdRepo
type mockVDRepo struct {
	addressById map[string]connection.Conn
}

func (m mockVDRepo) GetVDStateById(ctx context.Context, id string) (virtual_device.Flag, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockVDRepo) GetAllVDStates(ctx context.Context) (map[string]virtual_device.Flag, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockVDRepo) GetVDConnById(ctx context.Context, id string) (connection.Conn, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockVDRepo) SetVDConnById(ctx context.Context, id string, address connection.Conn) error {
	//TODO implement me
	panic("implement me")
}

func NewMockVDRepository() repository.VDRepository {
	return &mockVDRepo{}
}
