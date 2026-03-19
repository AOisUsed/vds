package repository

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
	"virturalDevice/internal/vds/domain/connection"
	"virturalDevice/internal/vds/domain/repository"
	"virturalDevice/internal/vds/domain/virtualdevice/params"
)

// Repository 测试用模拟vdRepo
type Repository struct {
	connByID     map[string]connection.Connection
	vdParamsByID map[string]params.Params

	rwMu             sync.RWMutex
	simulatedLatency time.Duration //模拟数据库操作的延迟，便于测试context取消功能
}

// NewVDRepository 创建mock repo, simulatedLatency 用于模拟数据库操作时间
func NewVDRepository(simulatedLatency time.Duration) repository.VDRepository {
	return &Repository{
		connByID:         make(map[string]connection.Connection),
		vdParamsByID:     make(map[string]params.Params),
		simulatedLatency: simulatedLatency,
	}
}

func (repo *Repository) SetParams(ctx context.Context, id string, params params.Params) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(repo.simulatedLatency):
		repo.rwMu.Lock()
		defer repo.rwMu.Unlock()
		repo.vdParamsByID[id] = params
		return nil
	}
}

// RemoveParams 不存在params不会返回error
func (repo *Repository) RemoveParams(ctx context.Context, id string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(repo.simulatedLatency):
		repo.rwMu.Lock()
		defer repo.rwMu.Unlock()
		delete(repo.vdParamsByID, id)
	}
	return nil
}

func (repo *Repository) Params(ctx context.Context, id string) (params.Params, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(repo.simulatedLatency):
		repo.rwMu.RLock()
		defer repo.rwMu.RUnlock()
		if val, ok := repo.vdParamsByID[id]; ok {
			return val, nil
		}
		return nil, errors.New(fmt.Sprintf("不存在设备%v的参数", id))
	}
}

func (repo *Repository) AllParams(ctx context.Context) (map[string]params.Params, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(repo.simulatedLatency):
		repo.rwMu.RLock()
		defer repo.rwMu.RUnlock()
		if len(repo.vdParamsByID) <= 0 {
			return nil, errors.New("数据库为空")
		}
		return repo.vdParamsByID, nil
	}
}

func (repo *Repository) Connection(ctx context.Context, id string) (connection.Connection, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(repo.simulatedLatency):
		repo.rwMu.RLock()
		defer repo.rwMu.RUnlock()
		if val, ok := repo.connByID[id]; ok {
			return val, nil
		}
		return nil, errors.New(fmt.Sprintf("不存在设备%v的连接", id))
	}
}

func (repo *Repository) SetConnection(ctx context.Context, id string, conn connection.Connection) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(repo.simulatedLatency):
		repo.rwMu.Lock()
		defer repo.rwMu.Unlock()
		repo.connByID[id] = conn
		return nil
	}
}

// RemoveConnection 注意：移除不存在的VDConn不会返回error
func (repo *Repository) RemoveConnection(ctx context.Context, id string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(repo.simulatedLatency):
		repo.rwMu.Lock()
		defer repo.rwMu.Unlock()
		delete(repo.connByID, id)
		return nil
	}
}
