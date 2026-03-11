package mock

import (
	"context"
	"errors"
	"fmt"
	"time"
	"virturalDevice/internal/connection"
	"virturalDevice/internal/vds/types"
	"virturalDevice/internal/vds/vdrepository"
)

// Repository 测试用模拟vdRepo
type Repository struct {
	connById     map[string]connection.Connection
	vdParamsByID map[string]types.VDParams

	simulatedLatency time.Duration //模拟数据库操作的延迟，便于测试context取消功能
}

// NewVDRepository 创建mock repo, simulatedLatency 用于模拟数据库操作时间
func NewVDRepository(simulatedLatency time.Duration) vdrepository.VDRepository {
	return &Repository{
		connById:         make(map[string]connection.Connection),
		vdParamsByID:     make(map[string]types.VDParams),
		simulatedLatency: simulatedLatency,
	}
}

func (repo *Repository) SetVDParamsById(ctx context.Context, id string, params types.VDParams) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(repo.simulatedLatency):
		repo.vdParamsByID[id] = params
		return nil
	}
}

func (repo *Repository) GetVDParamsById(ctx context.Context, id string) (types.VDParams, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(repo.simulatedLatency):
		if val, ok := repo.vdParamsByID[id]; ok {
			return val, nil
		}
		return nil, errors.New(fmt.Sprintf("不存在设备%v的参数", id))
	}
}

func (repo *Repository) GetAllVDParams(ctx context.Context) (map[string]types.VDParams, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(repo.simulatedLatency):
		return repo.vdParamsByID, nil
	}
}

func (repo *Repository) GetVDConnById(ctx context.Context, id string) (connection.Connection, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(repo.simulatedLatency):
		return repo.connById[id], nil
	}
}

func (repo *Repository) SetVDConnById(ctx context.Context, id string, conn connection.Connection) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(repo.simulatedLatency):
		repo.connById[id] = conn
		return nil
	}
}

// RemoveVDConnById 注意：移除不存在的VDConn不会返回error
func (repo *Repository) RemoveVDConnById(ctx context.Context, id string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(repo.simulatedLatency):
		delete(repo.connById, id)
		return nil
	}
}
