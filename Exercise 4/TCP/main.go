package main

import (
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"time"
)

func Primary(lastValue int) {
	ln, _ := net.Listen("tcp", ":20022")
	//exec.Command("gnome-terminal", "--", "go", "run", "main.go").Run()            // Linux
	exec.Command("cmd", "/C", "start", "powershell", "go", "run", "main.go").Run() // Windows
	conn, _ := ln.Accept()
	fmt.Println("[Primary initialized]")

	var str string
	for {
		lastValue++
		str = strconv.Itoa(int(lastValue))
		data := []byte(str)
		if _, err := conn.Write(data); err != nil {
			fmt.Println(lastValue)
			fmt.Println("[Error writing. Reinitializing Primary]")
			conn.Close()
			ln.Close()
			Primary(lastValue)
			continue
		}
		fmt.Println(lastValue)
		time.Sleep(2000 * time.Millisecond)
	}
}

func Backup() int {
	addr, _ := net.ResolveTCPAddr("tcp", "localhost:20022")
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		fmt.Println("[Error dialing]")
		return 0
	}
	fmt.Println("[Backup initialized]")

	var num int
	for {
		buf := make([]byte, 1024)
		value, err := conn.Read(buf)
		if err != nil {
			fmt.Println("[Sending num:", num, "]")
			conn.Close()
			return num
		} else {
			data := buf[0:value]
			strdata := string(data)
			num, _ = strconv.Atoi(strdata)
			fmt.Println(num)
		}
	}
}

func main() {
	lastValue := 0
	lastValue = Backup()
	Primary(lastValue)
}
