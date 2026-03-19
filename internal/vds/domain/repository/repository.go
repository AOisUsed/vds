package repository

import (
	"context"
	"virturalDevice/internal/vds/domain/connection"
	"virturalDevice/internal/vds/domain/virtualdevice/params"
)

// VDRepository 虚拟设备相关数据仓库接口
type VDRepository interface {
	Params(ctx context.Context, id string) (params.Params, error)         // 根据 id 查找虚拟设备状态参数
	SetParams(ctx context.Context, id string, params params.Params) error // 根据 id 设置虚拟设备状态参数
	RemoveParams(ctx context.Context, id string) error                    //根据 id 删除拟设备参数
	AllParams(ctx context.Context) (map[string]params.Params, error)      // 找到所有设备的状态参数

	Connection(ctx context.Context, id string) (connection.Connection, error)       // 根据 id 查找虚拟设备的连接信息
	SetConnection(ctx context.Context, id string, conn connection.Connection) error // 根据 id 设置虚拟设备的连接信息
	RemoveConnection(ctx context.Context, id string) error                          // 根据 id 删除虚拟设备的连接信息
}
