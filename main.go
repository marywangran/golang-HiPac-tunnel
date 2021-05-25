package main

import (
//	"runtime"
	"fmt"
	"os"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"strconv"
	"tuntap/tunnel"
	"tuntap/tun"
)

func init() {
//	runtime.LockOSThread()
//	runtime.GOMAXPROCS(48)
}

func main() {
	var client bool = false
	var queues int = 4
	var addr [4]byte
	var key	int

	if len(os.Args) == 5 {
		if os.Args[1] == "client" {
			client = true
		}
		queues, _ = strconv.Atoi(os.Args[2])
		ip := net.ParseIP(os.Args[3]).To4()
		for i := 0; i < len(ip); i += 1 {
			addr[i] = ip[i]
		}
		key, _ = strconv.Atoi(os.Args[4])
	} else {
		fmt.Println("./tuntap server|client(mode) 10(queues) 192.168.56.1(IP address) key(pre shared key)")
		os.Exit(1)
	}
	go func() {
		log.Println(http.ListenAndServe("127.0.0.1:6061", nil))
	}()

	tun := func() (tun.Device) {
		return tun.CreateTUN("wg2", 1500, queues)
	} ()

	instance := tunnel.NewInstance(tun, key, addr, client, queues)
	instance.WG.Wait()
}
