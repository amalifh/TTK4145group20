package main

import (
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

func Primary() {
	addr, _ := net.ResolveUDPAddr("udp", "localhost:20022")
	conn, _ := net.DialUDP("udp", nil, addr)
	exec.Command("gnome-terminal", "--", "go", "run", "main.go").Run()
	var str string
	lastValue := Backup()
	for{
		lastValue++
		str = strconv.Itoa(int(lastValue))
		data := []byte(str)
		conn.Write(data)
		fmt.Println(lastValue)
		time.Sleep(1500 * time.Millisecond)		
	}
	
}

func Backup() int {
	udpAddr, err := net.ResolveUDPAddr("udp", ":20022")
	if err != nil {
		return 0
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println(err)
		return 0 
	}
	fmt.Println("Backup connected")
	defer conn.Close()
	for {
		buf := make([]byte, 1024)
		conn.SetDeadline(time.Now().Add(3 * time.Second))
		value, err := conn.Read(buf)
		data := buf[0:value]
		strdata := string(data)
		num, _:= strconv.Atoi(strdata)
		fmt.Println(num)
		if err != nil{
			fmt.Println(err)
			fmt.Println("Sending num:", num)
			return num
		}
	}
}

func main() {
	//var lastValue int
	var wg sync.WaitGroup

	wg.Add(1)
	//fmt.Println("LastValue: ", lastValue)
	Primary()
	wg.Wait()
	
}
