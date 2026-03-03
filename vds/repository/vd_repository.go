package repository

import (
	"virturalDevice/vds/address"
	"virturalDevice/vds/virtual_device"
)

// VDRepository 虚拟设备相关数据仓库接口
type VDRepository interface {
	GetVDStateById(id string) (virtual_device.Flag, error)   // 根据 ID 查找虚拟设备状态信息
	GetAllVDStates() (map[string]virtual_device.Flag, error) // 找到所有在线设备状态信息
	GetVDAddrById(id string) (address.Address, error)        // 根据 id 查找虚拟设备的地址
	SetVDAddrById(id string, address address.Address) error
}
