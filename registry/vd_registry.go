// Package registry
// 指明了所有虚拟设备对应的vds的消息入口
package registry

import (
	"virturalDevice/message"
)

type VDRegistry struct {
	VdsByVdID map[string]chan<- message.Message // vd 到 vds 消息入口的映射
}

func NewRegistry() *VDRegistry {
	return &VDRegistry{
		VdsByVdID: make(map[string]chan<- message.Message),
	}
}

// RegisterVD 注册虚拟设备到注册中心中
func (r *VDRegistry) RegisterVD(vdId string, sendChan chan<- message.Message) {
	r.VdsByVdID[vdId] = sendChan
}

// DeRegisterVD 删除虚拟设备
func (r *VDRegistry) DeRegisterVD(vdId string) {
	delete(r.VdsByVdID, vdId)
}

// GetSendChan 获取vd所在vds的消息发送通道
func (r *VDRegistry) GetSendChan(vdId string) chan<- message.Message {
	return r.VdsByVdID[vdId]
}
