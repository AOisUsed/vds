package dispatcher

import (
	"context"
	"errors"
	"log"
	"sync"
	"virturalDevice/internal/vds/codec"
	"virturalDevice/internal/vds/repository"
	"virturalDevice/internal/vds/sender"
	"virturalDevice/internal/vds/virtualdevice/message"
)

// Dispatcher 消息分发器
type Dispatcher struct {
	incomingCh   <-chan message.Task // 接收来自消息集中器的消息
	vdRepository repository.VDRepository
	codec        codec.Codec
	sender       sender.Sender

	workerPool *WorkerPool
}

func NewDispatcher(incomingCh <-chan message.Task, vdRepository repository.VDRepository, codec codec.Codec, sender sender.Sender, numWorkers int) *Dispatcher {
	return &Dispatcher{
		incomingCh:   incomingCh,
		vdRepository: vdRepository,
		codec:        codec,
		sender:       sender,
		workerPool:   NewDispatchWorkerPool(incomingCh, numWorkers),
	}
}

// Run 运行消息分发器, 创建工人并开始工作 (阻塞)
func (d *Dispatcher) Run() {
	log.Println("正在启动 dispatcher")
	d.workerPool.Start(d.dispatch)
	d.workerPool.Wait()
}

// Stop 停止消息分发器 (注意：强制停止，会使 incomingCh 上游阻塞，)
func (d *Dispatcher) Stop() {
	d.workerPool.Stop()
	log.Println("dispatcher 停止")
}

// dispatch 分发消息
func (d *Dispatcher) dispatch(messagingTask message.Task) {
	ctx := messagingTask.Ctx
	msg := messagingTask.Message
	select {
	case <-ctx.Done():
		log.Printf("%v->%v 消息输送取消", msg.SrcID, msg.DstID)
		return
	default:
		if msg.DstID != "" {
			d.dispatchUnicast(ctx, msg)
		} else {
			d.dispatchMulticast(ctx, msg)
		}
	}
}

// dispatchUnicast 分发单播消息（消息有明确目标地址的情况）
func (d *Dispatcher) dispatchUnicast(ctx context.Context, msg message.Message) {
	srcParams, err := d.vdRepository.Params(ctx, msg.SrcID)
	if err != nil && !errors.Is(err, context.Canceled) {
		log.Printf("无法获取消息来源设备 %s 的状态: %v\n", msg.SrcID, err)
		return
	}

	dstParams, err := d.vdRepository.Params(ctx, msg.DstID)
	if err != nil && !errors.Is(err, context.Canceled) {
		log.Printf("无法获取消息目标设备 %s 的状态: %v\n", msg.DstID, err)
		return
	}

	if !srcParams.IsCompatibleWith(dstParams) {
		log.Printf("消息来源 %s 设备与 消息目标 %s 设备无法沟通\n", msg.SrcID, msg.DstID)
		return
	}

	dstConn, err := d.vdRepository.Connection(ctx, msg.DstID)
	if err != nil && !errors.Is(err, context.Canceled) {
		log.Printf("无法获取目标设备 %s 的连接: %v\n", msg.DstID, err)
		return
	}

	data, err := d.codec.Encode(msg)
	if err != nil {
		log.Printf("无法解码消息:%v \n", err)
	}

	if err = d.sender.Send(dstConn, data); err != nil {
		log.Printf("无法给 %s 发送消息: %v", msg.DstID, err)
	}
}

// dispatchUnicast 分发多播消息（消息无明确目标地址的情况）
func (d *Dispatcher) dispatchMulticast(ctx context.Context, msg message.Message) {
	// 找到所有可达设备
	dstIDs, err := d.FindValidTargetVDs(ctx, msg.SrcID)
	if err != nil && !errors.Is(err, context.Canceled) {
		log.Printf("无法找到消息来源设备 %s 可联络的设备: %v\n", msg.SrcID, err)
		return
	}

	// 并发向所有可达设备发送消息
	var wg sync.WaitGroup
	for _, dstID := range dstIDs {

		if ctx.Err() != nil {
			break // ctx 取消，直接跳出循环，不再启动新的 goroutine
		}

		wg.Add(1)
		go func(dstID string) {
			defer wg.Done()

			// 检查是否 ctx 取消
			select {
			case <-ctx.Done():
				return // 立即退出，不查库也不发送
			default:
			}

			dstConn, err := d.vdRepository.Connection(ctx, dstID)
			if err != nil && !errors.Is(err, context.Canceled) {
				log.Printf("无法获取多播消息目标设备 %s 的连接信息: %v\n", dstID, err)
				return
			}

			msgToSend := message.Message{
				SrcID:   msg.SrcID,
				DstID:   dstID,
				Payload: msg.Payload,
			}

			data, err := d.codec.Encode(msgToSend)
			if err != nil {
				log.Printf("无法解码消息:%v \n", err)
			}

			if err = d.sender.Send(dstConn, data); err != nil && !errors.Is(err, context.Canceled) {
				log.Printf("无法获取多播消息目标设备 %s 的地址: %v\n", dstID, err)
				return
			}
		}(dstID)
	}

	wg.Wait()
}

// FindValidTargetVDs 根据消息来源设备id找到能够到达的目标设备id (指能够接收，且通信参数匹配可以沟通)
func (d *Dispatcher) FindValidTargetVDs(ctx context.Context, srcId string) ([]string, error) {
	// 1. 获得所有在线设备信息 (包括消息来源设备)
	onlineVDById, err := d.vdRepository.AllParams(ctx)
	if err != nil && !errors.Is(err, context.Canceled) {
		log.Printf("无法获取来源设备 %s 能沟通的设备: %v\n", srcId, err)
		return nil, err
	}

	// 2. 计算得到可达设备id
	var validTargetVDs []string
	srcParams := onlineVDById[srcId]

	for id, Params := range onlineVDById {
		if id == srcId {
			continue
		}
		if srcParams.IsCompatibleWith(Params) {
			validTargetVDs = append(validTargetVDs, id)
		}
	}
	return validTargetVDs, nil
}
