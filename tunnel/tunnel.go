package tunnel

import (
	"runtime"
	"sync"
	"tuntap/tun"
)

type Packet struct {
	sync.Mutex
	buffer   []byte
	packet   []byte
}

type Tunnel struct {
	client  bool
	key	int
	WG	sync.WaitGroup
	net struct {
		socket	*UDPScoket
		port    int
		addr    [4]byte
	}
	queue struct {
		inbound []chan *Packet
		outbound []chan *Packet
		encryption [][]chan *Packet
		decryption [][]chan *Packet
	}
	tun struct {
		tunnel tun.Device
		queues    int
	}
}

func NewInstance(tunTunnel tun.Device, key int, addr [4]byte, client bool, queues int) *Tunnel {
	tunnel := new(Tunnel)
	tunnel.client = client
	tunnel.key = key
	tunnel.tun.queues = queues
	tunnel.tun.tunnel = tunTunnel
	tunnel.net.port = 12346
	tunnel.net.addr = addr

	if tunnel.client {
		tunnel.net.socket = CreateUDPScoket(tunnel.net.port, tunnel.net.addr, tunnel.tun.queues, 1)
	} else {
		tunnel.net.socket = CreateUDPScoket(tunnel.net.port, tunnel.net.addr, tunnel.tun.queues, 0)
	}

	tunnel.queue.outbound = make([]chan *Packet, queues)
	tunnel.queue.inbound = make([]chan *Packet, queues)

	enc := runtime.NumCPU()/queues
	if enc < PortNum {
		enc = PortNum
	}
	tunnel.queue.encryption = make([][]chan *Packet, queues)
	tunnel.queue.decryption = make([][]chan *Packet, queues)

	for i := 0; i < queues; i += 1 {
		tunnel.queue.outbound[i] = make(chan *Packet, IOBufferLen)
		tunnel.queue.inbound[i] = make(chan *Packet, IOBufferLen)
		tunnel.queue.encryption[i] = make([]chan *Packet, enc)
		tunnel.queue.decryption[i] = make([]chan *Packet, enc)
		for j := 0; j < enc; j += 1 {
			tunnel.queue.encryption[i][j] = make(chan *Packet, CryptionBufferLen)
			tunnel.queue.decryption[i][j] = make(chan *Packet, CryptionBufferLen)
			go tunnel.RoutineDecryption(i, j)
			go tunnel.RoutineEncryption(i, j)
		}
		go tunnel.RoutineReadFromUDP(i, enc)
		go tunnel.RoutineWriteToTUN(i)
		go tunnel.RoutineReadFromTUN(i, enc)
		go tunnel.RoutineWriteToUDP(i)
	}
	tunnel.WG.Add(1)

	return tunnel
}
