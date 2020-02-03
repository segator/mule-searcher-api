package kad

import "hahajing/com"



// SocketManager is manager of sockets for distributing sending packets to sockets by round robin.
type SocketManager struct {
	sockets []*Socket
	config *com.Config
	round   int

	sendCh chan *Packet
}

func (s *SocketManager) start(pPrefs *Prefs, recvCh, sendCh chan *Packet,config *com.Config) bool {
	s.sendCh = sendCh // channel for sending packets
	s.config = config

	// start sockets
	for i := 0; i < s.config.NumberOfSocket; i++ {
		socket := &Socket{no: i}
		sendCh1 := make(chan *Packet, cap(sendCh)/config.NumberOfSocket)
		udpPort := uint16(config.UDPPort + i)
		if !socket.start(pPrefs, recvCh, sendCh1, udpPort) {
			return false
		}

		s.sockets = append(s.sockets, socket)
	}

	// loop to distribute sending packets
	go s.sendRoutine()

	return true
}

func (s *SocketManager) sendRoutine() {
	for {
		packet := <-s.sendCh
		s.sockets[s.round].sendCh <- packet

		s.round++
		s.round = s.round % s.config.NumberOfSocket
	}
}
