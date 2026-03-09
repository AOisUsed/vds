package vdrepository

import (
	"context"
	"virturalDevice/pkg/connection"
	"virturalDevice/pkg/vds/radiomodel"
)

// VDRepository 虚拟设备相关数据仓库接口
type VDRepository interface {
	GetVDStateById(ctx context.Context, id string) (radiomodel.Params, error)    // 根据 ID 查找虚拟设备状态信息
	GetAllVDStates(ctx context.Context) (map[string]radiomodel.Params, error)    // 找到所有在线设备状态信息
	GetVDConnById(ctx context.Context, id string) (connection.Connection, error) // 根据 id 查找虚拟设备的地址
	SetVDConnById(ctx context.Context, id string, address connection.Connection) error
}
