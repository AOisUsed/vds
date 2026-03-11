package vds

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"
	"virturalDevice/internal/mock"
)

// 单机使用vds (repo仅在此vds中)
func NewMockVDS() *VDS {
	conn := mock.NewConn()
	repo := mock.NewVDRepository(time.Millisecond * 600)
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

func TestBasicRegisterDevice(t *testing.T) {
	vds := NewMockVDS()
	var wg sync.WaitGroup

	wg.Add(1)
	// 启动 vds
	vds.Start()

	// 模拟成功注册
	fmt.Println("\n 模拟注册")
	ctx := context.Background()
	err := vds.RegisterDeviceConn(ctx, "1")
	if err != nil {
		log.Println(err.Error())
	}

	time.Sleep(2 * time.Second)

	// 模拟注册进行中用户取消
	fmt.Println("\n 模拟注册中取消")

	ctxWithCancel, cancel := context.WithCancel(context.Background())

	// 模拟并发取消注册
	go func() {
		cancel()
	}()

	err = vds.RegisterDeviceConn(ctxWithCancel, "2")

	if err != nil {
		log.Println(err.Error())
	}

	time.Sleep(2 * time.Second)
	fmt.Println("\n 开始停止 vds")

	// 停止 vds
	vds.Stop()
	wg.Done()
	wg.Wait()
}

func TestConcurrentRegisterDevice(t *testing.T) {
	vds := NewMockVDS()
	var wg sync.WaitGroup
	wg.Add(1)
	// 启动 vds
	vds.Start()

	var registerWg sync.WaitGroup

	// 测试并发 register device
	for i := 0; i < 3; i++ {
		registerWg.Add(1)
		go func() {
			defer registerWg.Done()
			_ = vds.RegisterDeviceConn(context.Background(), strconv.Itoa(i))
		}()
	}

	registerWg.Wait() // 保证并发的注册结束后才停止vds
	// 停止 vds
	vds.Stop()
	wg.Done()
	wg.Wait()
}

// 线性注册设备，然后并发删除注册信息
func TestConcurrentDeregisterDevice(t *testing.T) {
	vds := NewMockVDS()
	var wg sync.WaitGroup
	wg.Add(1)
	// 启动 vds
	vds.Start()

	// 注册15个设备
	for i := 0; i < 15; i++ {
		_ = vds.RegisterDeviceConn(context.Background(), strconv.Itoa(i))
	}
	fmt.Printf("已完成设备注册\n\n")

	// 删除10个设备的注册信息，每3个中选一个取消删除注册信息 (并发，可能成功也可能失败)
	var deregisterWg sync.WaitGroup
	for i := 0; i < 10; i++ {
		deregisterWg.Add(1)
		ctxWithCancel, cancel := context.WithCancel(context.Background())
		go func() {
			if i%3 == 0 {
				cancel()
			}

		}()

		go func() {
			defer deregisterWg.Done()
			_ = vds.DeregisterDeviceConn(ctxWithCancel, strconv.Itoa(i))
		}()
	}

	deregisterWg.Wait() // 等待 deregister 流程结束
	// 停止 vds
	fmt.Println("\n 即将停止vds")
	vds.Stop()
	wg.Done()
	wg.Wait()
}

func TestConcurrentRegisterDeregisterDevice(t *testing.T) {
	vds := NewMockVDS()
	var wg sync.WaitGroup
	wg.Add(1)
	// 启动 vds
	vds.Start()

	var regderegWg sync.WaitGroup

	// 并发注册设备连接信息
	for i := 0; i < 5; i++ {
		regderegWg.Add(1)
		go func() {
			defer regderegWg.Done()
			_ = vds.RegisterDeviceConn(context.Background(), strconv.Itoa(i))
		}()
	}

	// 并发删除设备连接信息
	for i := 0; i < 5; i++ {
		regderegWg.Add(1)
		go func() {
			defer regderegWg.Done()
			_ = vds.DeregisterDeviceConn(context.Background(), strconv.Itoa(i))
		}()
	}

	regderegWg.Wait()

	fmt.Println("\n 即将开始关闭vds")
	// 停止 vds
	vds.Stop()

	wg.Done()
	fmt.Printf("\ngoroutine数量:%v\n", runtime.NumGoroutine())
	wg.Wait()
}

func TestBasicCommunication(t *testing.T) {
	var wg sync.WaitGroup

	// vds1, vds2 公用的repository (临时使用)
	repo := mock.NewVDRepository(time.Millisecond * 600)

	// 存入 3个设备的 params，用于比对是否匹配
	err := repo.SetVDParamsById(context.Background(), "1", mock.NewRadioParams())
	if err != nil {
		return
	}

	err = repo.SetVDParamsById(context.Background(), "2", mock.NewRadioParams())
	if err != nil {
		return
	}

	err = repo.SetVDParamsById(context.Background(), "3", mock.NewRadioParams())
	if err != nil {
		return
	}

	// 创造并启动 vds1
	vds1 := NewVDS(mock.NewConn(), repo, mock.NewSender(), mock.NewCodec())
	wg.Add(1)
	vds1.Start()

	// 创造并启动 vds2
	vds2 := NewVDS(mock.NewConn(), repo, mock.NewSender(), mock.NewCodec())
	wg.Add(1)
	vds2.Start()

	// vds1 中注册设备1连接信息
	err = vds1.RegisterDeviceConn(context.Background(), "1")
	if err != nil {
		log.Println(err.Error())
	}

	// vds2 中注册设备2连接信息
	err = vds2.RegisterDeviceConn(context.Background(), "2")
	if err != nil {
		log.Println(err.Error())
	}

	// vds2 中注册设备3连接信息
	err = vds2.RegisterDeviceConn(context.Background(), "3")
	if err != nil {
		log.Println(err.Error())
	}

	// 设备1 给 设备2 发送消息
	vds1.DeviceByID("1").Send("2", []byte("message 1->2"))

	// 设备2 给 设备1 发送消息
	vds2.DeviceByID("2").Send("1", []byte("message 2->1"))

	// 设备1 发出广播
	vds1.DeviceByID("1").Send("", []byte("1 发出 广播"))

	// 设备2 发出广播
	vds2.DeviceByID("2").Send("", []byte("2 发出 广播"))

	time.Sleep(2 * time.Second)

	fmt.Println("\n 即将开始关闭vds")
	// 停止 vds
	vds1.Stop()
	wg.Done()

	vds2.Stop()
	wg.Done()

	fmt.Printf("\ngoroutine数量:%v\n", runtime.NumGoroutine())
	wg.Wait()
}
