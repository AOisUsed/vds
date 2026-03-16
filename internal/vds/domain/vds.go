package domain

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"runtime"
	"sync"
	"time"
	"virturalDevice/internal/vds/domain/aggregator"
	"virturalDevice/internal/vds/domain/codec"
	"virturalDevice/internal/vds/domain/connection"
	"virturalDevice/internal/vds/domain/dispatcher"
	"virturalDevice/internal/vds/domain/ingressrouter"
	"virturalDevice/internal/vds/domain/message"
	"virturalDevice/internal/vds/domain/repository"
	"virturalDevice/internal/vds/domain/sender"
	"virturalDevice/internal/vds/domain/virtualdevice"
	"virturalDevice/internal/vds/domain/virtualdevice/params"
)

type VDS struct {
	// 设备
	vdById map[string]*virtualdevice.VirtualDevice // 设备id-实体映射

	// vds 消息入口
	conn       connection.Connection
	incomingCh chan message.Message

	// 数据
	vdRepository repository.VDRepository // 虚拟设备信息仓库

	// 消息处理
	ingressRouter *ingressrouter.IngressRouter // 消息入口路由
	aggregator    *aggregator.Aggregator       // 消息集合器
	dispatcher    *dispatcher.Dispatcher       //消息分发器
	sender        sender.Sender                //消息出口发送器

	// 编解码器
	codec codec.Codec

	// 设备状态同步
	paramsSyncerById map[string]chan struct{} // 设备参数同步

	rwMutex   sync.RWMutex
	startOnce sync.Once
	stopOnce  sync.Once
}

// NewVDS 初始化VDS(无设备)
func NewVDS(conn connection.Connection, repo repository.VDRepository, sender sender.Sender, codec codec.Codec) *VDS {
	vds := &VDS{
		vdById:           make(map[string]*virtualdevice.VirtualDevice),
		conn:             conn,
		incomingCh:       make(chan message.Message, 100), // 设置100的缓存
		vdRepository:     repo,
		sender:           sender,
		codec:            codec,
		paramsSyncerById: make(map[string]chan struct{}),
	}
	vds.ingressRouter = ingressrouter.NewIngressRouter(vds.incomingCh)
	vds.aggregator = aggregator.NewAggregator()
	vds.dispatcher = dispatcher.NewDispatcher(vds.aggregator.OutChan(), vds.vdRepository, vds.codec, sender, runtime.NumCPU()*2)
	return vds
}

// SendMessage 由vds处理，通过解析 message 的内部信息，调用对应的设备发送消息给目标设备
func (vds *VDS) SendMessage(msg message.Message) error {
	vds.rwMutex.RLock()
	defer vds.rwMutex.RUnlock()

	srcId := msg.SrcID
	dstId := msg.DstID
	device, ok := vds.vdById[srcId]
	if !ok {
		return errors.New(fmt.Sprintf("%v设备不存在", srcId))
	}
	device.SendMessage(dstId, msg.Payload)
	return nil
}

// listenConnection 持续从连接读取数据并处理
func (vds *VDS) listenConnection() {
	defer close(vds.incomingCh)

	for {
		data, err := vds.conn.Receive()
		if err != nil {
			if err != io.EOF {
				log.Printf("无法从连接读取数据:%v \n", err)

				time.Sleep(3 * time.Second) // 出现EOF外其他error后, 5秒后再重新尝试读数据
				continue
			}
			log.Printf("连接关闭，停止从中读取数据\n")
			return
		}
		msg, err := vds.codec.Decode(data)
		if err != nil {
			log.Printf("无法解码接收到的数据:%v \n", err)
			return
		}
		vds.incomingCh <- msg
	}
}

// SubscribeDeviceMessage 订阅某个设备的消息接收
//
// 点对点模式：多个订阅者竞争消息接收，无法共享
func (vds *VDS) SubscribeDeviceMessage(id string) (<-chan message.Message, error) {
	vds.rwMutex.RLock()
	defer vds.rwMutex.RUnlock()

	vd, ok := vds.vdById[id]
	if !ok {
		return nil, errors.New(fmt.Sprintf("vds中无此设备%v，无法订阅消息", id))
	}
	msgCh := vd.SubscribeMessage()
	return msgCh, nil
}

// syncDeviceParams 把设备参数同步到数据仓库中 (并发不安全)
func (vds *VDS) syncDeviceParams(ctx context.Context, id string) error {
	vd, ok := vds.vdById[id]
	if !ok {
		vds.rwMutex.RUnlock()
		return errors.New(fmt.Sprintf("vds中无设备%v", id))
	}

	parameters := vd.Params()
	err := vds.vdRepository.SetParams(ctx, id, parameters)
	return err
}

// updateDeviceParams 更新设备参数 (仅内存中更新,不同步到数据仓库) (并发不安全)
func (vds *VDS) updateDeviceParams(id string, params params.Params) error {
	vd, ok := vds.vdById[id]
	if !ok {
		return errors.New(fmt.Sprintf("vds中无此设备%v，无法更新参数", id))
	}
	vd.UpdateParams(params)
	return nil
}

func (vds *VDS) updateAndSyncParams(id string, params params.Params) error {
	vds.rwMutex.RLock()
	defer vds.rwMutex.RUnlock()

	// 本地设备内更新参数
	err := vds.updateDeviceParams(id, params)
	if err != nil {
		return err
	}
	// 交给upda

}

// registerDeviceConn 数据仓库中添加设备连接信息
func (vds *VDS) registerDeviceConn(ctx context.Context, id string) error {
	err := vds.vdRepository.SetConnection(ctx, id, vds.conn)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			log.Printf("注册设备%v连接信息取消：%v\n", id, err)
		} else {
			log.Printf("注册设备%v连接信息失败：%v\n", id, err)
		}
	}
	return err
}

// deregisterDeviceConn 删除数据仓库中设备连接信息
func (vds *VDS) deregisterDeviceConn(ctx context.Context, id string) error {
	err := vds.vdRepository.RemoveConnection(ctx, id)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			log.Printf("删除设备%v连接信息取消：%v\n", id, err)
		} else {
			log.Printf("删除设备%v连接信息失败：%v\n", id, err)
		}
	}
	return err
}

// activateDevice vds中创建设备，与前后消息通道连接，并运行设备。
//
// 如果同id的设备存在会删除原设备，替换为新设备
func (vds *VDS) activateDevice(id string, opts ...virtualdevice.Option) {
	vds.rwMutex.Lock()
	defer vds.rwMutex.Unlock()

	if _, exists := vds.vdById[id]; exists {
		vds.terminateDevice(id)
	}
	routerOutCh := vds.ingressRouter.CreateOutboundCh(id)
	vd := virtualdevice.NewVirtualDevice(id, routerOutCh, opts...)
	vdOutCh := vd.OutChan()
	vds.aggregator.Watch(vdOutCh)
	go vd.Run()
	vds.vdById[id] = vd
}

// terminateDevice 停止并移除设备和及相关消息收发通道。
//
// 多次调用，则第一次调用后都是无操作
func (vds *VDS) terminateDevice(id string) {
	vds.rwMutex.Lock()
	defer vds.rwMutex.Unlock()

	if _, exists := vds.vdById[id]; !exists {
		return
	}
	vds.ingressRouter.RemoveOutboundCh(id)
	vds.vdById[id].Stop()
	delete(vds.vdById, id)
}

func (vds *VDS) runParamsSyncer(id string) {
	// 创建 参数同步处理器
	vds.rwMutex.Lock()
	if _, exists := vds.paramsSyncerById[id]; exists {
		return
	}

	devId := id
	paramsSyncer := make(chan struct{}, 1)
	vds.paramsSyncerById[id] = paramsSyncer
	vds.rwMutex.Unlock()

	// 运行 参数同步处理器
	for range paramsSyncer {
		err := vds.syncDeviceParams(context.Background(), devId)
		if err != nil {
			// 如果失败，重试
			select {
			case paramsSyncer <- struct{}{}:
			default:
			}
			continue
		}
	}

}

// ActivateAndRegisterDevice 连接并注册设备连接信息到数据库中,更新
func (vds *VDS) ActivateAndRegisterDevice(ctx context.Context, id string, opts ...virtualdevice.Option) error {

	// 创建并运行设备
	vds.activateDevice(id, opts...)

	// 注册设备到数据仓库中
	err := vds.registerDeviceConn(ctx, id)
	if err != nil {
		// 失败则回滚
		vds.terminateDevice(id)
		return err
	}

	log.Printf("注册设备%v连接信息成功\n", id)
	return nil
}

// TerminateAndDeregisterDevice 停止，断开vd设备，并删除数据库中设备连接信息
func (vds *VDS) TerminateAndDeregisterDevice(ctx context.Context, id string) error {

	err := vds.deregisterDeviceConn(ctx, id)
	if err != nil {
		return err
	}
	log.Printf("删除设备%v连接信息成功\n", id)
	vds.terminateDevice(id)
	return err
}

// Start 启动vds服务
//
// 第一次执行后，后续调用都是无操作
func (vds *VDS) Start() {
	vds.startOnce.Do(func() {
		//log.Println("正在启动 vds ")
		go vds.dispatcher.Run()
		go vds.ingressRouter.Run()
		go vds.listenConnection()
	})
}

// Stop 停止vds服务
//
// 第一次执行后，后续调用都是无操作
func (vds *VDS) Stop() {
	vds.stopOnce.Do(func() {
		vds.rwMutex.RLock()
		defer vds.rwMutex.RUnlock()

		err := vds.conn.Close()
		if err != nil {
			log.Printf("连接关闭失败:%v \n", err)
		}

		vds.ingressRouter.Stop()
		for id, vd := range vds.vdById {
			go func(id string, vd *virtualdevice.VirtualDevice) {
				_ = vds.deregisterDeviceConn(context.Background(), id) // 尝试删除数据仓库中设备连接信息，删除失败也要停止本地goroutine
				vd.Stop()
			}(id, vd)

		}
		vds.aggregator.Stop()
		vds.dispatcher.Stop()

		log.Println("vds 停止")
	})
}
