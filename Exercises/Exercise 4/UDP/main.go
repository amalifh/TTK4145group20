package main

import (
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

func Primary(lastValue int) {
	addr, _ := net.ResolveUDPAddr("udp", "localhost:20022")
	conn, _ := net.DialUDP("udp", nil, addr)
	//exec.Command("gnome-terminal", "--", "go", "run", "main.go").Run()            // Linux
	exec.Command("cmd", "/C", "start", "powershell", "go", "run", "main.go").Run() // Windows
	var str string

	for {
		lastValue++
		str = strconv.Itoa(int(lastValue))
		data := []byte(str)
		conn.Write(data)
		fmt.Println(lastValue)
		time.Sleep(1500 * time.Millisecond)
	}

}

func Backup() int {
	udpAddr, _ := net.ResolveUDPAddr("udp", ":20022")
	conn, _ := net.ListenUDP("udp", udpAddr)

	fmt.Println("Backup connected")
	defer conn.Close()
	var num int
	for {
		buf := make([]byte, 1024)
		conn.SetDeadline(time.Now().Add(3 * time.Second))
		value, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Sending num:", num)
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
	var wg sync.WaitGroup
	wg.Add(1)
	lastValue := Backup()
	fmt.Println("LastValue: ", lastValue)
	Primary(lastValue)
	wg.Wait()

}
