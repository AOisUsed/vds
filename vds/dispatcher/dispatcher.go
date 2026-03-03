package dispatcher

import (
	"log"
	"net"
	"virturalDevice/message"
	"virturalDevice/vds/address"
	"virturalDevice/vds/repository"
	"virturalDevice/vds/sender"
)

// Dispatcher 消息分发器
type Dispatcher struct {
	incomingCh   <-chan message.Message // 出站消息统一出口
	vdRepository repository.VDRepository
	sender       sender.Sender
}

func NewDispatcher(incomingCh <-chan message.Message, vdRepository repository.VDRepository, sender sender.Sender) *Dispatcher {
	return &Dispatcher{
		incomingCh:   incomingCh,
		vdRepository: vdRepository,
		sender:       sender,
	}
}

// Serve 运行消息分发器
func (d *Dispatcher) Serve() {
	for msg := range d.incomingCh {
		go d.Dispatch(msg) // 使用goroutine并发分发消息，因为dispatch耗时，不应阻塞dispatcher接收新的消息
	}
}

// Dispatch 根据消息中信息和注册中心中vds消息接收通道分发单条消息到对应vds
func (d *Dispatcher) Dispatch(incomingMsg message.Message) {
	// 如果有明确的目标id(点对点发送):
	// 直接查询目标是否可达，如果可达则发送
	if incomingMsg.DstID != "" {
		dstId := incomingMsg.DstID
		srcState, err := d.vdRepository.GetVDStateById(dstId)
		if err != nil {
			log.Println(err)
			return
		}

		dstState, err := d.vdRepository.GetVDStateById(incomingMsg.DstID)
		if err != nil {
			log.Println(err)
			return
		}

		if srcState.IsCompatibleWith(dstState) {
			dstAddr, err := d.vdRepository.GetVDAddrById(dstId)
			if err != nil {
				log.Println(err)
				return
			}

			if err = d.sender.Send(dstAddr, incomingMsg); err != nil {
				log.Println(err)
				return
			}
		}
		return
	}

	// 如果没有明确的目标地址，执行以下分支：
	// 获得所有可达的目标设备的 ID
	srcId := incomingMsg.SrcID
	dstIds, err := d.FindReachableVDs(srcId)
	if err != nil {
		log.Println(err)
		return
	}

	var dstAddrById map[string]address.Address

	// 根据可达设备 ID 获得目标设备地址
	for _, dstId := range dstIds {
		dstAddr, err := d.vdRepository.GetVDAddrById(dstId)
		if err != nil {
			log.Println(err)
			continue
		}
		dstAddrById[dstId] = dstAddr
	}

	// 调用 Sender 向目标设备发送消息
	for dstId, dstAddr := range dstAddrById {
		msgToSend := message.Message{
			SrcID: srcId,
			DstID: dstId,
			Body:  incomingMsg.Body,
		}
		if err = d.sender.Send(dstAddr, msgToSend); err != nil {
			log.Println(err)
			continue
		}
	}
}

// FindReachableVDs 根据消息来源设备id找到能够到达的目标设备id
func (d *Dispatcher) FindReachableVDs(srcId string) ([]string, error) {
	// 1. 获得所有在线设备信息 (包括消息来源设备)
	onlineVDById, err := d.vdRepository.GetAllVDStates()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// 2. 计算得到可达设备id
	var reachableVDIds []string
	srcVDState := onlineVDById[srcId]

	for id, vdState := range onlineVDById {
		if id == srcId {
			continue
		}
		if srcVDState.IsCompatibleWith(vdState) {
			reachableVDIds = append(reachableVDIds, id)
		}
	}
	return reachableVDIds, nil
}

func (d *Dispatcher) Send(dstAddr net.UDPAddr, msg message.Message) {

}
