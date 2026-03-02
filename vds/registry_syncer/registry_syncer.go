package registry_syncer

import (
	"virturalDevice/message"
	"virturalDevice/registry"
)

// RegistrySyncer 注册同步器，负责和注册中心沟通
type RegistrySyncer struct {
	vdRegistry *registry.VDRegistry
}

func NewRegistrySyncer(vdRegistry *registry.VDRegistry) *RegistrySyncer {
	return &RegistrySyncer{vdRegistry: vdRegistry}
}

// RegisterVD 将vd注册到注册中心中
func (r *RegistrySyncer) RegisterVD(id string, inputCh chan message.Message) {
	r.vdRegistry.RegisterVD(id, inputCh)
}

// GetSendChan 获取vd所在vds的消息发送通道
func (r *RegistrySyncer) GetSendChan(vdId string) chan<- message.Message {
	return r.vdRegistry.VdsByVdID[vdId]
}
