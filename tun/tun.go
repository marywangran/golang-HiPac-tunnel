package tun

import (
	"os"
	"unsafe"
	"golang.org/x/sys/unix"
)

type Device interface {
	Read(int, []byte) (int, error)
	Write(int, []byte) (int, error)
}

const (
	cloneDevicePath = "/dev/net/tun"
	ifReqSize       = unix.IFNAMSIZ + 640
)

type NativeTun struct {
	rwFiles                 []*os.File
	queues                  int
}

func (tun *NativeTun) Write(index int, buff []byte) (int, error) {
	return tun.rwFiles[index % tun.queues].Write(buff)
}

func (tun *NativeTun) Read(index int, buff []byte) (int, error) {
	n, _ := tun.rwFiles[index % tun.queues].Read(buff[:])
	return n, nil
}

func CreateTUN(name string, mtu int, queues int) (Device) {

	var fds []*os.File = make([]*os.File, queues)
	var ifr [ifReqSize]byte
	var flags uint16 = unix.IFF_TUN | unix.IFF_MULTI_QUEUE
	nameBytes := []byte(name)
	copy(ifr[:], nameBytes)
	*(*uint16)(unsafe.Pointer(&ifr[unix.IFNAMSIZ])) = flags

	for i := 0; i < len(fds); i++ {
		nfd, _ := unix.Open(cloneDevicePath, os.O_RDWR, 0)
		unix.Syscall(unix.SYS_IOCTL, uintptr(nfd), uintptr(unix.TUNSETIFF), uintptr(unsafe.Pointer(&ifr[0])))
		unix.SetNonblock(nfd, false)

		fds[i] = os.NewFile(uintptr(nfd), cloneDevicePath)
	}
	tun := &NativeTun{
		rwFiles:                 fds,
		queues:                  queues,
	}
	return tun
}
