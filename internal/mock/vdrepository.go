package mock

import (
	"context"
	"virturalDevice/internal/connection"
	"virturalDevice/internal/vds/types"
	"virturalDevice/internal/vds/vdrepository"
)

// 测试用模拟vdRepo
type Repository struct {
	addressById map[string]connection.Connection
}

func (repo *Repository) RemoveVDConnById(ctx context.context.Context, id string) error {
	//TODO implement me
	panic("implement me")
}

func NewVDRepository() vdrepository.VDRepository {
	return &Repository{}
}

func (repo *Repository) GetVDStateById(ctx context.Context, id string) (types.VDParams, error) {
	//TODO implement me
	panic("implement me")
}

func (repo *Repository) GetAllVDStates(ctx context.Context) (map[string]types.VDParams, error) {
	//TODO implement me
	panic("implement me")
}

func (repo *Repository) GetVDConnById(ctx context.Context, id string) (connection.Connection, error) {
	//TODO implement me
	panic("implement me")
}

func (repo *Repository) SetVDConnById(ctx context.Context, id string, address connection.Connection) error {
	//TODO implement me
	panic("implement me")
}
