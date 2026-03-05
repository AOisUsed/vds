// Package repository 与数据库交互层
package repository

import (
	"context"
	"virturalDevice/vds/connection"
	"virturalDevice/vds/virtual_device"
)

// VDRepository 虚拟设备相关数据仓库接口
type VDRepository interface {
	GetVDStateById(ctx context.Context, id string) (virtual_device.Flag, error) // 根据 ID 查找虚拟设备状态信息
	GetAllVDStates(ctx context.Context) (map[string]virtual_device.Flag, error) // 找到所有在线设备状态信息
	GetVDConnById(ctx context.Context, id string) (connection.Conn, error)      // 根据 id 查找虚拟设备的地址
	SetVDConnById(ctx context.Context, id string, address connection.Conn) error
}
