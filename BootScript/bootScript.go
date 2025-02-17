package main

import (
	"log"
	"os/exec"
	"runtime"
	"time"
)

func StartSims() {
	cmd1 := exec.Command("C:\\Users\\aukne\\OneDrive\\Documents\\00 Øvinger\\6. Semester Elsys\\Sanntidsprogrammering\\TTK4145group20\\Elevators\\1\\SimElevatorServer.exe")
	cmd2 := exec.Command("C:\\Users\\aukne\\OneDrive\\Documents\\00 Øvinger\\6. Semester Elsys\\Sanntidsprogrammering\\TTK4145group20\\Elevators\\2\\SimElevatorServer.exe")
	cmd3 := exec.Command("C:\\Users\\aukne\\OneDrive\\Documents\\00 Øvinger\\6. Semester Elsys\\Sanntidsprogrammering\\TTK4145group20\\Elevators\\3\\SimElevatorServer.exe")

	if err := cmd1.Start(); err != nil {
		log.Fatal(err)
	}
	if err := cmd2.Start(); err != nil {
		log.Fatal(err)
	}
	if err := cmd3.Start(); err != nil {
		log.Fatal(err)
	}
}

func StartMains() {
	cmd1 := exec.Command("cmd", "/C", "start", "powershell", "-NoExit", "go", "run", "C:\\Users\\aukne\\OneDrive\\Documents\\00 Øvinger\\6. Semester Elsys\\Sanntidsprogrammering\\TTK4145group20\\driver-go\\main.go", "15657")
	cmd2 := exec.Command("cmd", "/C", "start", "powershell", "-NoExit", "go", "run", "C:\\Users\\aukne\\OneDrive\\Documents\\00 Øvinger\\6. Semester Elsys\\Sanntidsprogrammering\\TTK4145group20\\driver-go\\main.go", "15658")
	cmd3 := exec.Command("cmd", "/C", "start", "powershell", "-NoExit", "go", "run", "C:\\Users\\aukne\\OneDrive\\Documents\\00 Øvinger\\6. Semester Elsys\\Sanntidsprogrammering\\TTK4145group20\\driver-go\\main.go", "15659")

	if err := cmd1.Start(); err != nil {
		log.Fatal(err)
	}
	if err := cmd2.Start(); err != nil {
		log.Fatal(err)
	}
	if err := cmd3.Start(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	runtime.GOMAXPROCS(1)
	StartSims()
	time.Sleep(3 * time.Second)
	StartMains()
}
