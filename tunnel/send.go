package tunnel

import (
//	"fmt"
	"sync"
)

func addToEncryptionBuffer(outboundQueue chan *Packet, encryptionQueue chan *Packet, pktent *Packet) {
	outboundQueue <- pktent
	encryptionQueue <- pktent
}

func (tunnel *Tunnel) RoutineReadFromTUN(queue int, max_enc int) {
	pool := make([]Packet, IOBufferLen, IOBufferLen)
	for i := 0; i < len(pool); i += 1 {
		pool[i].buffer = make([]byte, MaxPacketSzie, MaxPacketSzie)
		pool[i].Mutex = sync.Mutex{}
		pool[i].Lock()
	}
	var pos, enc int = 0, 0
	for {
		pkt := pool[pos % len(pool)]
		size, _ := tunnel.tun.tunnel.Read(queue, pkt.buffer[:])
		pkt.packet = pkt.buffer[:size]
		//fmt.Printf("####### read from tun:%d\n", index)
		addToEncryptionBuffer(tunnel.queue.outbound[queue], tunnel.queue.encryption[queue][enc % max_enc], &pkt)
		pos += 1
		enc += 1
	}
}

func (tunnel *Tunnel) RoutineEncryption(queue int, enc int) {
	key := byte(tunnel.key)
	for {
		pkt, _ := <-tunnel.queue.encryption[queue][enc]
		// encrypt packet
		for i := 0; i < len(pkt.packet); i += 1 {
			pkt.packet[i] += key
		}
		pkt.Unlock()
	}
}

func (tunnel *Tunnel) RoutineWriteToUDP(index int) {
	for {
		pkt, _ := <-tunnel.queue.outbound[index]
		pkt.Lock()
		//fmt.Printf("####### Write to UDP:%d\n", index)
		tunnel.Send(index, pkt.packet)
	}
}
