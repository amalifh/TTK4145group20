package network

import "Driver-go/elevator"

func TCPNetworkInit() {
	// Initialize the TCP network
	chanButtonEvent := make(chan elevator.ButtonEvent)
	chanBehaviour := make(chan elevator.ElevatorBehaviour)
	chanElevator := make(chan elevator.Elevator)

}

func TCPNetworkSend() {
	// Send a message over the TCP network

}

func TCPNetworkReceive() {
	// Receive a message over the TCP network

}
