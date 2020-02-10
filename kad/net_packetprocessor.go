package kad

import (
	"encoding/hex"
	"hahajing/com"
	"strings"
	"time"
)

// PacketProcessor process sending and receiving of packets from socket.
// It will generate and send packet(for sending), filter and dispatch packet(for receiving).
type PacketProcessor struct {
	pPrefs          *Prefs
	pContactManager *ContactManager
	pSearchManager  *SearchManager
	pPacketReqGuard *PacketReqGuard

	sendCh chan *Packet
}
func (pp *PacketProcessor) GetSendChannel()  chan *Packet {
	return pp.sendCh
}

func (pp *PacketProcessor) start(pPrefs *Prefs, pContactManager *ContactManager, pSearchManager *SearchManager, pPacketReqGuard *PacketReqGuard, sendCh chan *Packet) {
	pp.pPrefs = pPrefs
	pp.pContactManager = pContactManager
	pp.pSearchManager = pSearchManager
	pp.pPacketReqGuard = pPacketReqGuard

	pp.sendCh = sendCh
}

func (pp *PacketProcessor) sendMyDetails(opcode byte, pContact *Contact) {
	version := pContact.getVersion()
	if version < kademliaVersion2_47a {
		return
	}

	bi := ByteIO{buf: make([]byte, 32)}

	bi.writeBytes(pp.pPrefs.getKadID().getHash())
	bi.writeUint16(pp.pPrefs.getTCPPort())
	bi.writeUint8(kademliaVersion)

	// I don't fill firewalled fields even if I'm firewalled.
	// It's cheating so that client will add me into its routing table.
	bi.writeUint8(0) // Tag count is 0.

	// contact KAD version check
	if version < kademliaVersion6_49aBeta {
		// low version not support encrytion
		contact := *pContact
		contact.pKadID = nil
		contact.resetUDPKey()

		pContact = &contact
	}

	pp.sendPacket(opcode, pContact, bi.getBuf())
}

func (pp *PacketProcessor) sendPacket(opcode byte, pContact *Contact, buf []byte) {
	// can we pass from guard?
	if !pp.pPacketReqGuard.add(time.Now().Unix(), pContact.ip, opcode) {
		com.HhjLog.Warningf("Sending %s to %s:%d, %s isn't passed by PacketReqGuard\n", getOpcodeStr(opcode), iIP2Str(pContact.ip), pContact.updPort, getVersionStr(pContact.getVersion()))
		return
	}

	version := pContact.getVersion()

	// new sending packet
	var receiverVerifyKey, senderVerifyKey uint32
	var clientKadID *ID
	if version >= kademliaVersion6_49aBeta {
		clientKadID = pContact.getKadID()
		receiverVerifyKey = pContact.getVerifyKey(pp.pPrefs.getPublicIP())
		if clientKadID != nil || receiverVerifyKey != 0 { // in case of encryption
			senderVerifyKey = pp.pPrefs.getUDPVerifyKey(pContact.getIP())
		}
	}

	packet := Packet{
		pKadID:            clientKadID,
		ip:                pContact.getIP(),
		port:              pContact.getUDPPort(),
		opcode:            opcode,
		buf:               buf,
		receiverVerifyKey: receiverVerifyKey,
		senderVerifyKey:   senderVerifyKey,
	}

	// send to socket
	pp.sendCh <- &packet
}

func (pp *PacketProcessor) processPacket(pPacket *Packet) {
	switch pPacket.opcode {
	case kademlia2HelloRes:
		pp.processKademlia2HelloRes(pPacket)
	case kademlia2Res:
		pp.processKademlia2Res(pPacket)
	case kademlia2SearchRes:
		pp.processKademlia2SearchRes(pPacket)
	case kademlia2SearchKeyReq:
		pp.processKademlia2SearchKeyReq(pPacket)
	case kademlia2Ping:
		pp.processKademlia2Ping(pPacket)
	case kademlia2Req:
		//pp.processKademlia2Req(pPacket)
	case kademliaFirewalled2Req:
		//pp.processKademliaFirewalled2Req(pPacket)
	/*case kademlia2SearchSourceReq:
		println(pPacket)*/
	default:
		print(pPacket.opcode)
	}
}
func (pp *PacketProcessor) processKademlia2Req(packet *Packet) {
	bi := ByteIO{buf: packet.buf}
	typeSearch := bi.readUint8()
	targetHash := bi.readBytes(16)
	check := bi.readBytes(16)
	tosend := []*Contact{}
	for _, pContact := range pp.pContactManager.contactMap{
		// send KAD request according to tolerance
		targetID:= ID{}
		targetID.setHash(targetHash)
		distance := targetID.getXor(pContact.getKadID())
		if distance.get32BitChunk(0) >= searchTolerance && len(tosend)<32 {
			tosend = append(tosend,pContact)
		}
	}
	buf := ByteIO{buf: make([]byte, 817)}
	buf.writeBytes(targetHash)
	buf.writeUint8(uint8(len(tosend)))
	com.HhjLog.Infof("Search %d userID %s check %s found: %d",typeSearch,strings.ToUpper(hex.EncodeToString(targetHash[:])),hex.EncodeToString(check[:]),len(tosend))
	for _,pContact := range tosend {
		buf.writeBytes(pContact.getKadID().hash[:])
		buf.writeUint32(pContact.getIP())
		buf.writeUint16(pContact.updPort)
		buf.writeUint16(pContact.getUDPPort())
		buf.writeUint8(pContact.getVersion())
	}
	contact := Contact{
		ip: packet.ip,
		updPort: packet.port,
		pKadID: packet.pKadID,
	}
	pp.sendPacket(kademlia2Res,&contact,buf.buf)
}
func (pp *PacketProcessor) processKademliaFirewalled2Req(packet *Packet) {

	/*bi := ByteIO{buf: packet.buf}
	tcpPort := bi.readUint16()
	userID := bi.readBytes(16)
	connectOptions := bi.readUint8()
	fmt.Println(fmt.Sprintf("firewall TCP Port %d %d user %s options %d",packet.port,tcpPort,userID,connectOptions))*/
	//TCP Connection
	/*CMemFile packetdata(4);
	packetdata.WriteUInt32(ip);
	DebugSend(KadFirewalledRes, ip, port);
	SendPacket(packetdata, KADEMLIA_FIREWALLED_RES, ip, port, senderKey, NULL);*/

}


func (pp *PacketProcessor) processKademlia2Ping(packet *Packet) {

	/*packet := Packet{
		pKadID:            clientKadID,
		ip:                pContact.getIP(),
		port:              pContact.getUDPPort(),
		opcode:            opcode,
		buf:               buf,
		receiverVerifyKey: receiverVerifyKey,
		senderVerifyKey:   senderVerifyKey,
	}*/

	// send to socket
	pp.sendCh <- packet
//	bi.writeUint16(packet.port)
	//pp.sendPacket(kademlia2Pong,contact,bi.getBuf())


}


func (pp *PacketProcessor) processKademlia2HelloRes(pPacket *Packet) {
	msg := Kademlia2HelloResMsg{}
	msg.set(pPacket)
	/*ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, msg.ip)
	println("Hello from "+ ip.String() + " " + string(msg.version))*/
	pp.pContactManager.addKademlia2HelloRes(&msg)
}

func (pp *PacketProcessor) processKademlia2SearchKeyReq(pPacket *Packet) {
	msg := Kademlia2SearchKeyReq{}
	msg.set(pPacket)

}

func (pp *PacketProcessor) processKademlia2Res(pPacket *Packet) {
	msg := Kademlia2ResMsg{}
	if msg.set(pPacket) {
		//pp.pContactManager.addKademlia2Res(&msg)
		if !pp.pSearchManager.addKademlia2Res(&msg) {
			// It's not our response for search key
			pp.pContactManager.addKademlia2Res(&msg)
		}
	}
}

func (pp *PacketProcessor) processKademlia2SearchRes(pPacket *Packet) {
	msg := Kademlia2SearchResMsg{}
	if msg.set(pPacket) {
		pp.pSearchManager.addKademlia2SearchRes(&msg)
	}
}

func (pp *PacketProcessor) sendFindValue(pContact *Contact, pTargetID *ID) {
	bi := ByteIO{buf: make([]byte, 33)}

	// how many contacts we wanted
	//isaac
	//nContactCount := kademliaFindNode
	nContactCount := uint8(100)
	bi.writeUint8(nContactCount)

	// target
	bi.writeBytes(pTargetID.getHash())

	// contact ID for contact check it
	bi.writeBytes(pContact.getKadID().getHash())

	version := pContact.getVersion()
	if version < kademliaVersion2_47a {
		return
	}

	if version < kademliaVersion6_49aBeta {
		// low version not support encrytion
		contact := *pContact
		contact.pKadID = nil
		contact.resetUDPKey()

		pContact = &contact
	}

	pp.sendPacket(kademlia2Req, pContact, bi.getBuf())
}

func (pp *PacketProcessor) sendSearchKeyword(pContact *Contact, targetHash []byte) {
	bi := ByteIO{buf: make([]byte, 33)}

	// target hash
	bi.writeBytes(targetHash)

	bi.writeUint16(uint16(0)) // ???

	version := pContact.getVersion()
	if version < kademliaVersion6_49aBeta {
		// low version not support encrytion
		contact := *pContact
		contact.pKadID = nil
		contact.resetUDPKey()

		pContact = &contact

		if version < kademliaVersion3_47b {
			return
		}
	}

	pp.sendPacket(kademlia2SearchKeyReq, pContact, bi.getBuf())
}




