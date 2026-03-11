package vds

import (
	"testing"
	"time"
	"virturalDevice/internal/mock"
)

func NewMockVDS() *VDS {
	conn := mock.NewConn()
	repo := mock.NewVDRepository(time.Millisecond * 500)
	sender := mock.NewSender()
	codec := mock.NewCodec()
	return NewVDS(conn, repo, sender, codec)
}

func TestBasicLifeCycle(t *testing.T) {
	vds := NewMockVDS()
	vds.Start()

	time.Sleep(3 * time.Second)
	vds.Stop()
}

func TestRegisterDevice(t *testing.T) {
	vds := NewMockVDS()

	vds.Start()

	time.Sleep(3 * time.Second)
}
