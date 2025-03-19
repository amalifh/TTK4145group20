package main

import (
	"log"
	"os/exec"
)

func main() {
	cmd := exec.Command("gnome-terminal", "--", "sh", "-c", "cd /home/student/Documents/TTK4145group20/Elevators/1 && ./SimElevatorServer.exe; exec bash")
	err := cmd.Start()
	if err != nil {
		log.Fatalf("Failed to start SimElevatorServer: %v", err)
	}
	log.Println("SimElevatorServer started successfully")

	cmd2 := exec.Command("gnome-terminal", "--", "sh", "-c", "cd /home/student/Documents/TTK4145group20/Elevators/2 && ./SimElevatorServer.exe; exec bash")
	err = cmd2.Start()
	if err != nil {
		log.Fatalf("Failed to start SimElevatorServer: %v", err)
	}
	log.Println("SimElevatorServer started successfully")

	cmd3 := exec.Command("gnome-terminal", "--", "sh", "-c", "cd /home/student/Documents/TTK4145group20/Elevators/3 && ./SimElevatorServer.exe; exec bash")
	err = cmd3.Start()
	if err != nil {
		log.Fatalf("Failed to start SimElevatorServer: %v", err)
	}
	log.Println("SimElevatorServer started successfully")

	cmd4 := exec.Command("gnome-terminal", "--", "sh", "-c", "cd /home/student/Documents/TTK4145group20/driver-go && go run main.go '15657'; exec bash")
	err = cmd4.Start()
	if err != nil {
		log.Fatalf("Failed to start main.go: %v", err)
	}
	log.Println("main.go started successfully")

	cmd5 := exec.Command("gnome-terminal", "--", "sh", "-c", "cd /home/student/Documents/TTK4145group20/driver-go && go run main.go '15658'; exec bash")
	err = cmd5.Start()
	if err != nil {
		log.Fatalf("Failed to start main.go: %v", err)
	}
	log.Println("main.go started successfully")

	cmd6 := exec.Command("gnome-terminal", "--", "sh", "-c", "cd /home/student/Documents/TTK4145group20/driver-go && go run main.go '15659'; exec bash")
	err = cmd6.Start()
	if err != nil {
		log.Fatalf("Failed to start main.go: %v", err)
	}
	log.Println("main.go started successfully")
}
