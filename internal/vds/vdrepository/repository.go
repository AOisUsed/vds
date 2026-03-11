package vdrepository

import (
	"context"
	"virturalDevice/internal/connection"
	"virturalDevice/internal/vds/types"
)

// VDRepository 虚拟设备相关数据仓库接口
type VDRepository interface {
	SetVDParamsById(ctx context.Context, id string, params types.VDParams) error
	GetVDParamsById(ctx context.Context, id string) (types.VDParams, error) // 根据 ID 查找虚拟设备状态信息
	GetAllVDParams(ctx context.Context) (map[string]types.VDParams, error)  // 找到所有在线设备状态信息

	GetVDConnById(ctx context.Context, id string) (connection.Connection, error)    // 根据 id 查找虚拟设备的连接信息
	SetVDConnById(ctx context.Context, id string, conn connection.Connection) error // 根据 id 设置虚拟设备的连接信息
	RemoveVDConnById(ctx context.Context, id string) error                          // 根据 id 删除虚拟设备的连接信息
}
