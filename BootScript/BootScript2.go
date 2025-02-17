package main

import (
	"log"
	"os/exec"
)

func main() {
	cmd := exec.Command("powershell", "-Command", "Start-Process 'C:\\YOUR\\FULL\\PATH\\HERE\\TTK4145group20\\Elevators\\1\\SimElevatorServer.exe' -WorkingDirectory 'C:\\YOUR\\FULL\\PATH\\HERE\\TTK4145group20\\Elevators\\1'")
	err := cmd.Start()
	if err != nil {
		log.Fatalf("Failed to start SimElevatorServer.exe: %v", err)
	}
	log.Println("SimElevatorServer.exe started successfully")

	cmd2 := exec.Command("powershell", "-Command", "Start-Process 'C:\\YOUR\\FULL\\PATH\\HERE\\TTK4145group20\\Elevators\\2\\SimElevatorServer.exe' -WorkingDirectory 'C:\\YOUR\\FULL\\PATH\\HERE\\TTK4145group20\\Elevators\\2'")
	err = cmd2.Start()
	if err != nil {
		log.Fatalf("Failed to start SimElevatorServer.exe: %v", err)
	}
	log.Println("SimElevatorServer.exe started successfully")

	cmd3 := exec.Command("powershell", "-Command", "Start-Process 'C:\\YOUR\\FULL\\PATH\\HERE\\TTK4145group20\\Elevators\\3\\SimElevatorServer.exe' -WorkingDirectory 'C:\\YOUR\\FULL\\PATH\\HERE\\TTK4145group20\\Elevators\\3'")
	err = cmd3.Start()
	if err != nil {
		log.Fatalf("Failed to start SimElevatorServer.exe: %v", err)
	}
	log.Println("SimElevatorServer.exe started successfully")

	cmd4 := exec.Command("cmd", "/C", "start", "powershell", "-NoExit", "-Command", "cd 'C:\\YOUR\\FULL\\PATH\\HERE\\TTK4145group20\\driver-go'; go run main.go '15657'; pause")
	err = cmd4.Start()
	if err != nil {
		log.Fatalf("Failed to start main.go: %v", err)
	}
	log.Println("main.go started successfully")

	cmd5 := exec.Command("cmd", "/C", "start", "powershell", "-NoExit", "-Command", "cd 'C:\\YOUR\\FULL\\PATH\\HERE\\TTK4145group20\\driver-go'; go run main.go '15658'; pause")
	err = cmd5.Start()
	if err != nil {
		log.Fatalf("Failed to start main.go: %v", err)
	}
	log.Println("main.go started successfully")

	cmd6 := exec.Command("cmd", "/C", "start", "powershell", "-NoExit", "-Command", "cd 'C:\\YOUR\\FULL\\PATH\\HERE\\TTK4145group20\\driver-go'; go run main.go '15659'; pause")
	err = cmd6.Start()
	if err != nil {
		log.Fatalf("Failed to start main.go: %v", err)
	}
	log.Println("main.go started successfully")
}
