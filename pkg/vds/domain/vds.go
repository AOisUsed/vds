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
	"virturalDevice/pkg/vds/domain/aggregator"
	"virturalDevice/pkg/vds/domain/codec"
	"virturalDevice/pkg/vds/domain/connection"
	"virturalDevice/pkg/vds/domain/dispatcher"
	"virturalDevice/pkg/vds/domain/ingressrouter"
	"virturalDevice/pkg/vds/domain/message"
	"virturalDevice/pkg/vds/domain/repository"
	"virturalDevice/pkg/vds/domain/sender"
	"virturalDevice/pkg/vds/domain/virtualdevice"
	"virturalDevice/pkg/vds/domain/virtualdevice/params"
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
	paramsSyncerTriggerById map[string]chan struct{} // 设备参数同步触发通道

	rwMutex   sync.RWMutex
	startOnce sync.Once
	stopOnce  sync.Once

	ctx    context.Context
	cancel context.CancelFunc
}

// NewVDS 初始化VDS(无设备)
func NewVDS(conn connection.Connection, repo repository.VDRepository, sender sender.Sender, codec codec.Codec) *VDS {
	vds := &VDS{
		vdById:                  make(map[string]*virtualdevice.VirtualDevice),
		conn:                    conn,
		incomingCh:              make(chan message.Message, 100), // 设置100的缓存
		vdRepository:            repo,
		sender:                  sender,
		codec:                   codec,
		paramsSyncerTriggerById: make(map[string]chan struct{}),
	}
	vds.ingressRouter = ingressrouter.NewIngressRouter(vds.incomingCh)
	vds.aggregator = aggregator.NewAggregator()
	vds.dispatcher = dispatcher.NewDispatcher(vds.aggregator.OutChan(), vds.vdRepository, vds.codec, sender, runtime.NumCPU()*2)
	return vds
}

// SendMessage 由vds处理，通过解析 message 的内部信息，调用对应的消息来源设备发送消息给目标设备
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
		data, err := vds.conn.Receive(vds.ctx)
		if err != nil && !errors.Is(err, context.Canceled) {
			if err != io.EOF {
				log.Printf("无法从连接读取数据:%v \n", err)
				select {
				case <-vds.ctx.Done():
					return
				case <-time.After(time.Second * 5): // 出现EOF外其他error后, 5秒后再重新尝试读数据
					continue
				}
			}
			log.Printf("连接关闭，停止从中读取数据\n")
			return
		}
		msg, err := vds.codec.Decode(data)
		if err != nil {
			log.Printf("无法解码接收到的数据:%v \n", err)
			continue
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

// syncDeviceParams 把设备参数同步到数据仓库中
func (vds *VDS) syncDeviceParams(ctx context.Context, id string) error {
	vds.rwMutex.RLock()
	vd, ok := vds.vdById[id]
	vds.rwMutex.RUnlock()

	if !ok {
		return errors.New(fmt.Sprintf("vds中无设备%v", id))
	}

	parameters := vd.Params()
	err := vds.vdRepository.SetParams(ctx, id, parameters)
	return err
}

// removeDeviceParams 把数据仓库中的设备参数移除
func (vds *VDS) removeDeviceParams(ctx context.Context, id string) error {
	err := vds.vdRepository.RemoveParams(ctx, id)
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

// triggerParamsSyncUnsafe 触发同步本地设备参数到数据仓库的操作 (具体同步操作见 runParamsSyncListener 方法) (并发不安全)
func (vds *VDS) triggerParamsSyncUnsafe(id string) {
	ch, exists := vds.paramsSyncerTriggerById[id]
	if !exists {
		log.Printf("%v设备已下线", id)
		return
	}

	// 策略：丢弃冗余触发
	// 由于参数同步以“执行时刻的最新值”为准，积压多个同步任务只会重复上传相同的最新数据，浪费资源。
	// 因此，若通道已满（表示已有任务在处理或排队），则直接丢弃本次触发信号，确保最多只有一个待处理任务。

	select {
	case ch <- struct{}{}:
	default:
		//log.Printf("设备 %v 同步请求过于频繁，已丢弃冗余触发", id)
	}
}

// triggerParamsSync 触发同步本地设备参数到数据仓库的操作 (具体同步操作见 runParamsSyncListener 方法)
func (vds *VDS) triggerParamsSync(id string) {
	vds.rwMutex.RLock()
	defer vds.rwMutex.RUnlock()
	vds.triggerParamsSyncUnsafe(id)
}

// runParamsSyncListener 创建并启动参数同步监听器：持续监听“把本地设备的参数上传到数据中心中”的请求,并执行。有失败重传机制 (并发不安全)
func (vds *VDS) runParamsSyncListener(id string) {
	// 创建 参数同步器
	devId := id
	syncTrigger := make(chan struct{}, 1)
	vds.rwMutex.Lock()
	vds.paramsSyncerTriggerById[id] = syncTrigger
	vds.rwMutex.Unlock()

	// 循环运行 参数同步处理器
	// 工作模式：接收到 syncTrigger 信号，则上传设备参数
	for range syncTrigger {
		err := vds.syncDeviceParams(context.Background(), devId)
		if err != nil {
			// 如果上传失败，通过给needSync发送通知，传达重试意图
			time.Sleep(1 * time.Second) // 退避重试: 失败过一段时间后才重试，防止雪崩。可以改变sleep的时长调整重试延时
			vds.triggerParamsSyncUnsafe(devId)
		}
		//time.Sleep(300 * time.Millisecond) // 可以在这里添加休眠时间，防止用户高频率更改设备参数产生的数据仓库写入压力
		log.Printf("设备 %v 同步了设备参数到数据仓库中", id)
	}
}

// stopParamsSyncListener 停止监听参数同步请求,并移除对应的触发器(即channel) (并发不安全)
func (vds *VDS) stopParamsSyncListener(id string) {
	close(vds.paramsSyncerTriggerById[id])
	delete(vds.paramsSyncerTriggerById, id)
}

// updateAndSyncParams 更新设备参数并异步上传到数据仓库中
func (vds *VDS) updateAndSyncParams(id string, params params.Params) error {
	vds.rwMutex.RLock()
	defer vds.rwMutex.RUnlock()

	// 本地设备内更新参数
	err := vds.updateDeviceParams(id, params)
	if err != nil {
		return err
	}

	// 交给 paramsSyncer 来同步设备参数到数据仓库
	vds.triggerParamsSyncUnsafe(id)
	return nil
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

// activateDevice vds中创建设备，与前后消息通道连接，开启参数同步监听器，并运行设备。
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
	go vds.runParamsSyncListener(id)
	go vd.Run()
	vds.vdById[id] = vd
}

// terminateDevice 停止并移除设备和及相关消息收发通道,停止参数同步监听器。
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
	vds.stopParamsSyncListener(id)
	delete(vds.vdById, id)
}

// ActivateAndRegisterDevice 连接并注册设备连接信息，同步设备参数到数据库中
func (vds *VDS) ActivateAndRegisterDevice(ctx context.Context, id string, opts ...virtualdevice.Option) error {
	// 创建并运行设备
	vds.activateDevice(id, opts...)

	// 注册设备连接到数据仓库中
	err := vds.registerDeviceConn(ctx, id)
	if err != nil {
		// 失败则回滚
		vds.terminateDevice(id)
		return err
	}
	// 同步设备参数到数据仓库中
	err = vds.syncDeviceParams(ctx, id)
	if err != nil {
		// 失败则回滚
		vds.terminateDevice(id)
		return err
	}

	log.Printf("注册设备%v连接信息和参数成功\n", id)
	return nil
}

// TerminateAndDeregisterDevice 停止，断开vd设备，并删除数据库中设备连接信息
func (vds *VDS) TerminateAndDeregisterDevice(ctx context.Context, id string) error {
	vds.terminateDevice(id)
	err := vds.deregisterDeviceConn(ctx, id)
	if err != nil {
		return err
	}
	err = vds.removeDeviceParams(ctx, id)
	if err != nil {
		// 注意: 此处报错，会出现 repo 中设备连接信息已经被抹除，但是参数信息还存在的情况,即不保证 参数和连接 的一致性
		return err
	}
	log.Printf("删除设备%v连接信息和参数信息成功\n", id)
	return err
}

// Start 启动vds服务
//
// 第一次执行后，后续调用都是无操作
func (vds *VDS) Start() {
	vds.startOnce.Do(func() {
		vds.ctx, vds.cancel = context.WithCancel(context.Background())
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
		vds.rwMutex.Lock()
		defer vds.rwMutex.Unlock()

		if vds.cancel != nil {
			vds.cancel()
		}
		if vds.conn != nil {
			err := vds.conn.Close()
			if err != nil {
				log.Printf("连接关闭失败:%v \n", err)
			}
		}

		vds.ingressRouter.Stop()

		var repoWg sync.WaitGroup
		for id, vd := range vds.vdById {

			// 尝试删除数据仓库中设备连接信息和参数信息，删除失败也要停止本地设备
			repoWg.Add(1)
			go func() {
				defer repoWg.Done()
				_ = vds.deregisterDeviceConn(context.Background(), id)
			}()
			repoWg.Add(1)
			go func() {
				defer repoWg.Done()
				_ = vds.removeDeviceParams(context.Background(), id)
			}()
			vds.stopParamsSyncListener(id)
			vd.Stop()
		}
		vds.aggregator.Stop()
		vds.dispatcher.Stop()

		repoWg.Wait()
		//log.Println("vds 停止")
	})
}
