package main

import (
	"fmt"
	"net"
)

//var buffer int = byte[1024]
//var addr string =

func main() {
	p := make([]byte, 1024)
	addr := net.UDPAddr{
		Port: 30000,
		IP:   net.ParseIP("129.241.187.255"),
	}
	msg, _ := net.ListenUDP("udp", &addr)

	_, remoteaddr, _ := msg.ReadFromUDP(p)
	fmt.Printf("Message read from %v %s\n", remoteaddr, p)

}
