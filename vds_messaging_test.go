package main

import (
	"fmt"
	"math/rand/v2"
	"strconv"
	"testing"
	"time"
	"virturalDevice/cipher"
	"virturalDevice/message"
	"virturalDevice/registry"
	"virturalDevice/vds"
	"virturalDevice/vds/virtual_device"
)

func TestVDMessage(T *testing.T) {
	aesgcmCipher, err := cipher.NewAESGCMCipher([]byte("1234567890123456"))
	if err != nil {
		panic(err)
	}

	// 创建注册中心
	vdRegistry := registry.NewRegistry()

	// 创建虚拟设备
	virtualDevices := make([]*virtual_device.VirtualDevice, 30)

	for i := range virtualDevices {
		virtualDevices[i] = virtual_device.NewVirtualDevice(strconv.Itoa(i), aesgcmCipher)
	}

	// 创建一些 vds
	vds1 := vds.NewVDS(make(chan message.Task), make(chan message.Task), vdRegistry)
	for i := 0; i < 5; i++ {
		vds1.RegisterDevice(virtualDevices[i])
	}

	vds2 := vds.NewVDS(make(chan message.Task), make(chan message.Task), vdRegistry)
	for i := 5; i < 10; i++ {
		vds2.RegisterDevice(virtualDevices[i])
	}

	vds3 := vds.NewVDS(make(chan message.Task), make(chan message.Task), vdRegistry)
	for i := 10; i < 15; i++ {
		vds3.RegisterDevice(virtualDevices[i])
	}

	vds4 := vds.NewVDS(make(chan message.Task), make(chan message.Task), vdRegistry)
	for i := 15; i < 20; i++ {
		vds4.RegisterDevice(virtualDevices[i])
	}

	vds5 := vds.NewVDS(make(chan message.Task), make(chan message.Task), vdRegistry)
	for i := 20; i < 25; i++ {
		vds5.RegisterDevice(virtualDevices[i])
	}

	vds6 := vds.NewVDS(make(chan message.Task), make(chan message.Task), vdRegistry)
	for i := 25; i < 30; i++ {
		vds6.RegisterDevice(virtualDevices[i])
	}

	// 运行这些 vds
	go vds1.Run()
	go vds2.Run()
	go vds3.Run()
	go vds4.Run()
	go vds5.Run()
	go vds6.Run()

	// 测试消息的发送与接收
	for i := 0; i < 5; i++ {
		srcId := rand.IntN(30)
		dstId := rand.IntN(30)
		fmt.Printf("外部调用%d发送消息给%d\n", srcId, dstId)
		virtualDevices[srcId].Send(strconv.Itoa(dstId), []byte(fmt.Sprintf("%d->%d", srcId, dstId)))
	}
	time.Sleep(3 * time.Second)

}
