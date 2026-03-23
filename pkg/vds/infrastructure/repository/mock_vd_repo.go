package repository

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
	"virturalDevice/pkg/vds/domain/connection"
	"virturalDevice/pkg/vds/domain/repository"
	"virturalDevice/pkg/vds/domain/virtualdevice/params"
)

// MockVDRepository 测试用模拟vdRepo
//
// 注意：由于进行数据库操作时，结构都是
//
// select {
// case <-ctx.Done():
//
//	// 一些操作...
//
// case <-time.After(repo.simulatedLatency):
//
//	// 一些操作...
//	}
//
// 如果测试时，主函数中，调用的数据库操作没有使用ctx.Done() 取消，
// 也没有等待数据库的模拟延迟计时(simulatedLatency)到达就结束，会因为goroutine还在等待两种情况中的一种而泄露。
//
// 解决办法：可以在主函数退出前使用
// time.Sleep(simulatedLatency) 保证select语句至少有一个出口可以结束，就可以防止此goroutine泄露
//
// 使用真实数据库时，由于一般会有连接关闭的步骤，所以不会有这样的问题
type MockVDRepository struct {
	connByID     map[string]connection.Connection
	vdParamsByID map[string]params.Params

	rwMu             sync.RWMutex
	simulatedLatency time.Duration //模拟数据库操作的延迟，便于测试context取消功能
}

// NewVDRepository 创建mock repo, simulatedLatency 用于模拟数据库操作时间
func NewVDRepository(simulatedLatency time.Duration) repository.VDRepository {
	return &MockVDRepository{
		connByID:         make(map[string]connection.Connection),
		vdParamsByID:     make(map[string]params.Params),
		simulatedLatency: simulatedLatency,
	}
}

func (repo *MockVDRepository) SetParams(ctx context.Context, id string, params params.Params) error {
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
func (repo *MockVDRepository) RemoveParams(ctx context.Context, id string) error {
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

func (repo *MockVDRepository) Params(ctx context.Context, id string) (params.Params, error) {
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

func (repo *MockVDRepository) AllParams(ctx context.Context) (map[string]params.Params, error) {
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

func (repo *MockVDRepository) Connection(ctx context.Context, id string) (connection.Connection, error) {
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

func (repo *MockVDRepository) SetConnection(ctx context.Context, id string, conn connection.Connection) error {
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
func (repo *MockVDRepository) RemoveConnection(ctx context.Context, id string) error {
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
