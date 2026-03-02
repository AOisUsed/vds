package registry_syncer

import (
	"virturalDevice/message"
	"virturalDevice/registry"
)

type RegistrySyncer struct {
	vdRegistry *registry.VDRegistry
}

func NewRegistrySyncer(vdRegistry *registry.VDRegistry) *RegistrySyncer {
	return &RegistrySyncer{vdRegistry: vdRegistry}
}

func (r *RegistrySyncer) RegisterVD(id string, inputCh chan message.Message) {
	r.vdRegistry.RegisterVD(id, inputCh)
}
