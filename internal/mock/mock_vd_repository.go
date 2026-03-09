package mock

import (
	"context"
	"virturalDevice/pkg/connection"
	"virturalDevice/pkg/vds/attribute"
	"virturalDevice/pkg/vds/vdrepository"
)

// 测试用模拟vdRepo
type Repo struct {
	addressById map[string]connection.Connection
}

func (m Repo) GetVDStateById(ctx context.Context, id string) (attribute.Flag, error) {
	//TODO implement me
	panic("implement me")
}

func (m Repo) GetAllVDStates(ctx context.Context) (map[string]attribute.Flag, error) {
	//TODO implement me
	panic("implement me")
}

func (m Repo) GetVDConnById(ctx context.Context, id string) (connection.Connection, error) {
	//TODO implement me
	panic("implement me")
}

func (m Repo) SetVDConnById(ctx context.Context, id string, address connection.Connection) error {
	//TODO implement me
	panic("implement me")
}

func NewMockVDRepository() vdrepository.VDRepository {
	return &Repo{}
}
