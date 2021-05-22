package tunnel

import (
//	"fmt"
	"golang.org/x/sys/unix"
)

type End struct {
	end     unix.Sockaddr
}

type UDPScoket struct {
	sock     []int
	end	 []End
	queues   int
}

func getSockaddr(port int, addr [4]byte) (sa unix.Sockaddr) {
        address := unix.SockaddrInet4 {
                        Port: port,
                        Addr: addr,
                }
        return &address
}

func CreateUDPScoket(port int, addr [4]byte, queues int, client int) (*UDPScoket) {
	socket := new(UDPScoket)
	socket.sock = make([]int, queues, queues)
	socket.end = make([]End, queues, queues)
	initial := make([]byte, 1, 1)
	initial[0] = 'H'
	for i := 0; i < queues; i += 1 {
		tport := port + i;
		socket.sock[i] = create()
		address := &unix.SockaddrInet4 {
					Port: tport,
					Addr: addr,
		}
		if client == 1 {
			socket.end[i].end = getSockaddr(tport, addr)
			unix.Connect(socket.sock[i], address)
			send(socket.sock[i], &socket.end[i], initial)
		} else {
			unix.Bind(socket.sock[i], address)
		}
	}
	socket.queues = queues
	return socket
}

func (tunnel *Tunnel) Receive(index int, buff []byte) (int) {
	socket := tunnel.net.socket
	n := receive(socket.sock[index], buff, &socket.end[index])
	return n
}

func (tunnel *Tunnel) Send(index int, buff []byte) {
	socket := tunnel.net.socket
	send(socket.sock[index], &socket.end[index], buff)
}

func create() (int) {
	fd, _ := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, 0)
	unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
	unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
	return fd
}

func send(sock int, end *End, buff []byte) {
	if end.end != nil {
		unix.Sendto(sock, buff, 0, end.end)
//		fmt.Printf("send internal #####   %d\n", sock)
	}
}

func receive(sock int, buff []byte, end *End) (int) {
	size, dst, _ := unix.Recvfrom(sock, buff, 0)
	end.end = dst
//	fmt.Printf("receive internal #####  sock:%d\n", sock)
	return size
}
