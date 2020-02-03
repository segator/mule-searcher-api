package kad

import (
	"hahajing/com"
	"time"
)

const (
	kadTimer               = 1
	kadPacketReqGuardTimer = 60

	kadSearchReqChSize = 50000
)

// Kad x
type Kad struct {
	prefs           Prefs
	contactManager  ContactManager
	packetProcesser PacketProcessor
	packetReqGuard  PacketReqGuard
	searchManager   SearchManager
	config *com.Config
	socketManager  SocketManager
	recvCh, sendCh chan *Packet

	// externs
	SearchReqCh chan *SearchReq
}

// Start x
func (k *Kad) Start(config *com.Config) bool {
	k.config=config
	k.SearchReqCh = make(chan *SearchReq, kadSearchReqChSize)

	socketChSize := bootstrapSearchContactNbr * int(kademliaFindNode) * int(kademliaFindNode)
	k.recvCh = make(chan *Packet, socketChSize)
	k.sendCh = make(chan *Packet, socketChSize)

	// start should be from bottom to up layer
	k.prefs.start(config)
	k.socketManager.start(&k.prefs, k.recvCh, k.sendCh,config)
	k.packetReqGuard.start()
	k.packetProcesser.start(&k.prefs, &k.contactManager, &k.searchManager, &k.packetReqGuard, k.sendCh)
	k.searchManager.start(&k.packetProcesser, &k.contactManager.onliner,config)

	k.contactManager.start(&k.prefs, &k.packetProcesser, &k.packetReqGuard,config)

	go k.scheduleRoutine()

	return true
}

func (k *Kad) scheduleRoutine() {
	tick := time.NewTicker(kadTimer * time.Second)
	packetReqGuardTimer := time.NewTicker(kadPacketReqGuardTimer * time.Second)

	for {
		select {
		case pPacket := <-k.recvCh:
			k.packetProcesser.processPacket(pPacket)
		case <-tick.C:
			//com.HhjLog.Infof("Contacts: %d\n", len(k.contactManager.contactMap))
			k.contactManager.tickProcess()
			k.searchManager.tickProcess()
		case <-packetReqGuardTimer.C:
			k.packetReqGuard.timerProcess()

		case pSearchReq := <-k.SearchReqCh:
			k.searchManager.newSearch(pSearchReq)
		}
	}
}
