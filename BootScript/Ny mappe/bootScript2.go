package main

import (
	"log"
	"os/exec"
)

func main() {
	cmd := exec.Command("cmd", "/c", "start", "powershell", "-NoExit", "-Command", "& {Start-Process 'C:\\Users\\aukne\\OneDrive\\Documents\\00 Øvinger\\6. Semester Elsys\\Sanntidsprogrammering\\TTK4145group20\\Elevators\\1\\SimElevatorServer.exe' -WorkingDirectory 'C:\\Users\\aukne\\OneDrive\\Documents\\00 Øvinger\\6. Semester Elsys\\Sanntidsprogrammering\\TTK4145group20\\Elevators\\1'; pause}")
	err := cmd.Start()
	if err != nil {
		log.Fatalf("Failed to start SimElevatorServer.exe: %v", err)
	}
	log.Println("SimElevatorServer.exe started successfully")

	cmd2 := exec.Command("cmd", "/c", "start", "powershell", "-NoExit", "-Command", "& {Start-Process 'C:\\Users\\aukne\\OneDrive\\Documents\\00 Øvinger\\6. Semester Elsys\\Sanntidsprogrammering\\TTK4145group20\\Elevators\\2\\SimElevatorServer.exe' -WorkingDirectory 'C:\\Users\\aukne\\OneDrive\\Documents\\00 Øvinger\\6. Semester Elsys\\Sanntidsprogrammering\\TTK4145group20\\Elevators\\1'; pause}")
	err = cmd2.Start()
	if err != nil {
		log.Fatalf("Failed to start SimElevatorServer.exe: %v", err)
	}
	log.Println("SimElevatorServer.exe started successfully")

	cmd3 := exec.Command("cmd", "/c", "start", "powershell", "-NoExit", "-Command", "& {Start-Process 'C:\\Users\\aukne\\OneDrive\\Documents\\00 Øvinger\\6. Semester Elsys\\Sanntidsprogrammering\\TTK4145group20\\Elevators\\3\\SimElevatorServer.exe' -WorkingDirectory 'C:\\Users\\aukne\\OneDrive\\Documents\\00 Øvinger\\6. Semester Elsys\\Sanntidsprogrammering\\TTK4145group20\\Elevators\\1'; pause}")
	err = cmd3.Start() // Updated from cmd.Start() to cmd3.Start()
	if err != nil {
		log.Fatalf("Failed to start SimElevatorServer.exe: %v", err)
	}
	log.Println("SimElevatorServer.exe started successfully")
}
