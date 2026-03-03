package repository

import (
	"errors"
	"virturalDevice/vds/address"
	"virturalDevice/vds/virtual_device"
)

// 测试用模拟vdRepo
type mockVDRepo struct {
	addressById map[string]address.Address
}

func (m mockVDRepo) SetVDAddrById(id string, address address.Address) error {
	m.addressById[id] = address
	return nil
}

func (m mockVDRepo) GetVDAddrById(id string) (address.Address, error) {
	if val, ok := m.addressById[id]; ok {
		return val, nil
	}
	return nil, errors.New("VD not found")
}

func (m mockVDRepo) GetVDStateById(id string) (virtual_device.Flag, error) {
	//todo implement me
	panic("implement me")
}

func (m mockVDRepo) GetAllVDStates() (map[string]virtual_device.Flag, error) {
	//TODO implement me
	panic("implement me")
}

func NewMockVDRepository() VDRepository {
	return &mockVDRepo{}
}
