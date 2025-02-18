package main

import (
	"net"
)

func main() {
	addr := &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 8080,
	}
	socket, err := net.DialUDP("udp", nil, addr)
}
