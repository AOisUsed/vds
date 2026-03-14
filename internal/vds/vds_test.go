package vds

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"
	"virturalDevice/internal/mock"
	"virturalDevice/internal/vds/virtualdevice"
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

	var regDeregWg sync.WaitGroup

	// 并发注册设备连接信息
	for i := 0; i < 5; i++ {
		regDeregWg.Add(1)
		go func() {
			defer regDeregWg.Done()
			_ = vds.RegisterDeviceConn(context.Background(), strconv.Itoa(i))
		}()
	}

	// 并发删除设备连接信息
	for i := 0; i < 5; i++ {
		regDeregWg.Add(1)
		go func() {
			defer regDeregWg.Done()
			_ = vds.DeregisterDeviceConn(context.Background(), strconv.Itoa(i))
		}()
	}

	regDeregWg.Wait()

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
	repo := mock.NewVDRepository(time.Millisecond * 1)
	// 创造并启动 vds1
	vds1 := NewVDS(mock.NewConn(), repo, mock.NewSender(), mock.NewCodec())
	wg.Add(1)
	vds1.Start()

	// 创造并启动 vds2
	vds2 := NewVDS(mock.NewConn(), repo, mock.NewSender(), mock.NewCodec())
	wg.Add(1)
	vds2.Start()

	// vds1 中注册设备1连接信息，更新设备参数
	err := vds1.ConnectAndRegisterDevice(context.Background(), "1")
	if err != nil {
		log.Println(err.Error())
	}
	err = vds1.UpdateDeviceParams(context.Background(), "1")
	if err != nil {
		log.Println(err.Error())
	}

	// vds2 中注册设备2连接信息，更新设备参数
	err = vds2.ConnectAndRegisterDevice(context.Background(), "2")
	if err != nil {
		log.Println(err.Error())
	}
	err = vds2.UpdateDeviceParams(context.Background(), "2")
	if err != nil {
		log.Println(err.Error())
	}

	// vds2 中注册设备3连接信息，更新设备参数
	err = vds2.ConnectAndRegisterDevice(context.Background(), "3")
	if err != nil {
		log.Println(err.Error())
	}
	err = vds2.UpdateDeviceParams(context.Background(), "3")
	if err != nil {
		log.Println(err.Error())
	}

	// 设备1 给 设备2 发送消息
	vds1.Device("1").Send("2", []byte("message 1->2"))

	// 设备2 给 设备1 发送消息
	vds2.Device("2").Send("1", []byte("message 2->1"))

	// 设备1 发出广播
	vds1.Device("1").Send("", []byte("1发出的广播"))

	// 设备2 发出广播
	vds2.Device("2").Send("", []byte("2发出的广播"))

	time.Sleep(2 * time.Second)

	fmt.Println("\n 即将开始关闭vds")
	// 停止 vds
	vds1.Stop()
	wg.Done()

	vds2.Stop()
	wg.Done()

	wg.Wait()
}

func TestConcurrentCommunication(t *testing.T) {
	var wg sync.WaitGroup

	// 公用的repository (临时使用)
	repo := mock.NewVDRepository(time.Millisecond * 600)

	// 创造并启动数个vds
	var vdss []*VDS
	for i := 0; i < 5; i++ {
		vdss = append(vdss, NewVDS(mock.NewConn(), repo, mock.NewSender(), mock.NewCodec()))
		wg.Add(1)
		vdss[i].Start()
	}

	idg := NewIdGenerator()

	for _, vds := range vdss {
		// 每个 vds 中产生数个 vd
		numVD := rand.Int() % 50
		//numVD := 5
		// 每个 vds 并发产生多个 vd,并发送消息
		go func(vds *VDS) {
			for j := 0; j < numVD; j++ {
				id := idg.Next()
				go func(vds *VDS) {
					err := vds.ConnectAndRegisterDevice(context.Background(), id)
					if err != nil {
						log.Println(err.Error())
					}
					err = vds.UpdateDeviceParams(context.Background(), id)

					dstId := rand.Int() % idg.Max()
					if dstId%3 != 0 {
						vds.Device(id).Send(strconv.Itoa(dstId), []byte(fmt.Sprintf("message %v->%d", id, dstId)))
					} else {
						vds.Device(id).Send("", []byte(fmt.Sprintf(" %v broadcast message ", id)))
					}

				}(vds)
			}
		}(vds)
	}

	// 等待发送完成
	time.Sleep(5 * time.Second)
	fmt.Println("\n 即将开始关闭vds")
	// 停止 vds
	for _, vds := range vdss {
		vds.Stop()
		wg.Done()
	}

	fmt.Printf("goroutine number: %v \n", runtime.NumGoroutine())

	wg.Wait()
}

func TestBasicParamMatchCommunication(t *testing.T) {
	var wg sync.WaitGroup

	// 公用的repository (临时使用)
	repo := mock.NewVDRepository(time.Millisecond * 20)

	// 创造并启动数个vds
	var vdss []*VDS
	for i := 0; i < 2; i++ {
		vdss = append(vdss, NewVDS(mock.NewConn(), repo, mock.NewSender(), mock.NewCodec()))
		wg.Add(1)
		vdss[i].Start()
	}

	idg := NewIdGenerator()
	for _, vds := range vdss {
		// 每个 vds 中产生数个 vd
		numVD := rand.Int() % 6
		// 每个 vds 并发产生多个 vd,并发送消息
		go func(vds *VDS) {
			for j := 0; j < numVD; j++ {
				id := idg.Next()
				go func(vds *VDS, j int) {
					mode := j % 3
					err := vds.ConnectAndRegisterDevice(context.Background(), id,
						virtualdevice.WithParams(
							mock.NewRadioParams(mock.WithMode(mode)), // todo：还没写完
						),
					)
					if err != nil {
						log.Println(err.Error())
					}
					log.Printf("设备%v的mode是%v\n", id, mode)
					err = vds.UpdateDeviceParams(context.Background(), id)

					dstId := rand.Int() % idg.Max()
					if dstId%3 == 0 {
						vds.Device(id).Send(strconv.Itoa(dstId), []byte(fmt.Sprintf("message %v->%d", id, dstId)))
					} else {
						vds.Device(id).Send("", []byte(fmt.Sprintf(" %v broadcast message ", id)))
					}

				}(vds, j)
			}
		}(vds)
	}

	time.Sleep(5 * time.Second)

	// 停止 vds
	for _, vds := range vdss {
		vds.Stop()
		wg.Done()
	}

	wg.Wait()
}
