package tunnel

import (
	"sync"
	//"fmt"
)

func addToDecryptionBuffer(inboundQueue chan *Packet, decryptionQueue chan *Packet, pktent *Packet) {
	inboundQueue <- pktent
	decryptionQueue <- pktent
}

func (tunnel *Tunnel) RoutineReadFromUDP(queue int, max_enc int) {
	pool := make([]Packet, IOBufferLen, IOBufferLen)
	for i := 0; i < len(pool); i += 1 {
		pool[i].buffer = make([]byte, MaxPacketSzie, MaxPacketSzie)
		pool[i].Mutex = sync.Mutex{}
		pool[i].Lock()
	}
	var pos, enc int = 0, 0
	for {
		pkt := pool[pos % len(pool)]
		//fmt.Printf("####### Receive from UDP:%d\n", queue)
		size := tunnel.Receive(queue, pkt.buffer[:])
		if pkt.buffer[0] == 'H' {
			continue
		}
		pkt.packet = pkt.buffer[:size]
		addToDecryptionBuffer(tunnel.queue.inbound[queue], tunnel.queue.decryption[queue][enc % max_enc], &pkt)
		pos += 1
		enc += 1
	}
}

func (tunnel *Tunnel) RoutineDecryption(queue int, enc int) {
	key := byte(tunnel.key)
	for {
		pkt, _ := <-tunnel.queue.decryption[queue][enc]
		// decrypt packet
		for i := 0; i < len(pkt.packet); i += 1 {
			pkt.packet[i] -= key
		}
		pkt.Unlock()
	}
}

func (tunnel *Tunnel) RoutineWriteToTUN(index int) {
	for {
		pkt, _ := <-tunnel.queue.inbound[index]
		pkt.Lock()
		//fmt.Printf("####### Write to TUN:%d\n", index)
		tunnel.tun.tunnel.Write(index, pkt.buffer[:len(pkt.packet)])
	}
}
