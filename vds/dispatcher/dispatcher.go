// Package dispatcher
// 消息分发器负责计算消息可达目标，并发送至对应目标
package dispatcher

import (
	"context"
	"log"
	"sync"
	"time"
	"virturalDevice/message"
	"virturalDevice/vds/repository"
	"virturalDevice/vds/sender"
)

// Dispatcher 消息分发器
type Dispatcher struct {
	incomingCh   <-chan message.Message // 接收来自消息集中器的消息
	vdRepository repository.VDRepository
	sender       sender.Sender

	workerPool *DispatchWorkerPool
}

func NewDispatcher(incomingCh <-chan message.Message, vdRepository repository.VDRepository, sender sender.Sender, numWorkers int) *Dispatcher {
	return &Dispatcher{
		incomingCh:   incomingCh,
		vdRepository: vdRepository,
		sender:       sender,
		workerPool:   NewDispatchWorkerPool(incomingCh, numWorkers),
	}
}

// Serve 运行消息分发器
func (d *Dispatcher) Serve() { // todo:未完成
	for i := 0; i < d.workerPool.numWorkers; i++ {
		go d.worker()
	}
}

// Stop 停止消息分发器
func (d *Dispatcher) Stop() {
	close(d.workerPool.done)
}

// worker 实际执行dispatch的worker
func (d *Dispatcher) worker(ctx context.Context) {
	for msg := range d.incomingCh {
		func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
			defer cancel()
			d.dispatch(ctx, msg)
		}()
	}
}

// dispatch 分发消息
func (d *Dispatcher) dispatch(ctx context.Context, incomingMsg message.Message) {
	select {
	case <-ctx.Done():
		log.Printf("dispatcher context done")
		return
	default:
		if incomingMsg.DstID != "" {
			d.dispatchUnicast(ctx, incomingMsg)
		} else {
			d.dispatchMulticast(ctx, incomingMsg)
		}
	}
}

// dispatchUnicast 分发单播消息（消息有明确目标地址的情况）
func (d *Dispatcher) dispatchUnicast(ctx context.Context, msg message.Message) {
	srcState, err := d.vdRepository.GetVDStateById(ctx, msg.SrcID)
	if err != nil {
		log.Printf("无法获取消息来源设备 %s 的状态: %v", msg.SrcID, err)
		return
	}

	dstState, err := d.vdRepository.GetVDStateById(ctx, msg.DstID)
	if err != nil {
		log.Printf("无法获取消息目标设备 %s 的状态: %v", msg.DstID, err)
		return
	}

	if !srcState.IsCompatibleWith(dstState) {
		log.Printf("消息来源 %s 设备与 消息目标 %s 设备无法沟通", msg.SrcID, msg.DstID)
		return
	}

	dstAddr, err := d.vdRepository.GetVDConnById(ctx, msg.DstID)
	if err != nil {
		log.Printf("无法获取目标设备 %s 的地址: %v", msg.DstID, err)
		return
	}

	if err = d.sender.Send(dstAddr, msg); err != nil {
		log.Printf("无法给 %s 发送消息: %v", msg.DstID, err)
	}
}

// dispatchUnicast 分发多播消息（消息无明确目标地址的情况）
func (d *Dispatcher) dispatchMulticast(ctx context.Context, msg message.Message) {
	dstIDs, err := d.FindValidTargetVDs(ctx, msg.SrcID)
	if err != nil {
		log.Printf("无法找到消息来源设备 %s 可联络的设备: %v", msg.SrcID, err)
		return
	}

	var wg sync.WaitGroup

	// 并发向所有可达设备发送消息
	for _, dstID := range dstIDs {
		wg.Add(1)
		go func(dstID string) {
			defer wg.Done()
			dstAddr, err := d.vdRepository.GetVDConnById(ctx, dstID)
			if err != nil {
				log.Printf("无法获取多播消息目标设备 %s 的地址: %v", dstID, err)
				return
			}

			msgToSend := message.Message{
				SrcID:   msg.SrcID,
				DstID:   dstID,
				Payload: msg.Payload,
			}

			if err = d.sender.Send(dstAddr, msgToSend); err != nil {
				log.Printf("无法获取多播消息目标设备 %s 的地址: %v", dstID, err)
				return
			}
		}(dstID)
	}

	wg.Wait()
}

// FindValidTargetVDs 根据消息来源设备id找到能够到达的目标设备id (指能够接收，且通信参数匹配可以沟通)
func (d *Dispatcher) FindValidTargetVDs(ctx context.Context, srcId string) ([]string, error) {
	// 1. 获得所有在线设备信息 (包括消息来源设备)
	onlineVDById, err := d.vdRepository.GetAllVDStates(ctx)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// 2. 计算得到可达设备id
	var validTargetVDs []string
	srcVDState := onlineVDById[srcId]

	for id, vdState := range onlineVDById {
		if id == srcId {
			continue
		}
		if srcVDState.IsCompatibleWith(vdState) {
			validTargetVDs = append(validTargetVDs, id)
		}
	}
	return validTargetVDs, nil
}
