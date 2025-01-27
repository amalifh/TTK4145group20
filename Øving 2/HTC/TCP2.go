package main

import (
	"bufio"
	"fmt"
	"net"
	"sync"
	"time"
)

var wg sync.WaitGroup

func connectTo() {
	defer wg.Done()
	addr, _ := net.ResolveTCPAddr("tcp", "10.100.23.204:33546")
	conn, _ := net.DialTCP("tcp", nil, addr)
	conn.Write([]byte("Connect to: 10.100.23.32:20022\x00"))
	//data,_ := bufio.NewReader(conn).ReadString('\x00')
	//fmt.Println("Received: ", string(data))
	time.Sleep(1500 * time.Millisecond)
}

func listenTo() {
	defer wg.Done()
	ln, _ := net.Listen("tcp", ":20022")
	conn, _ := ln.Accept()
	data,_ := bufio.NewReader(conn).ReadString('\x00')
	fmt.Println("Received: ", string(data))
//	conn.Write([]byte("It's working!\x00"))
	time.Sleep(1500 * time.Millisecond)
	conn.Close()
}

func main() {
	wg.Add(2)
	go connectTo()
	go listenTo()
	wg.Wait()
}