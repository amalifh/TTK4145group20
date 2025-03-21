/*
ToDO:
	- Timers are buggy
	- Implement a confirmation message system
	
Point 2:
The project has no clear acknowledgment system that confirms if an elevator has handled a request.
Request states are tracked (NEW, ASSIGNED, COMPLETED), but there is no verification from other elevators that it has been handled.
Our suggestion is to implement a confirmation message system of some sort, ensuring that once a request is completed, all elevators know about it.

Small fixes:
	- In elevator-Main there are several references to functions in localController, but the parameters used when
	passing the function is not equal to the definition done in localController.
	One example is onRequestButtonPress which needs three parameters, only receives two parameters when being used.
*/

package main

import (
	"Driver-go/elevator/driver"
	"Driver-go/elevator/types"
	elevator_controller "Driver-go/elevatorController/controller"
	"Driver-go/network/bcast"
	"Driver-go/network/peers"
	request_control "Driver-go/requests"
	"fmt"
	"os"
)

func main() {
	// Initialize the driver connection to the elevator server
	if len(os.Args) < 3 {
		fmt.Println("Usage: <program> <port> <id>")
		return
	}
	addr := os.Args[1]
	localID := os.Args[2]
	addr = "localhost:" + addr
	driver.Init(addr, types.N_FLOORS)

	// Polls hardware buttons and sends button presses into buttonEventCh.
	buttonEventCh := make(chan types.ButtonEvent)
	go driver.PollButtons(buttonEventCh)

	requestsCh := make(chan [types.N_FLOORS][types.N_BUTTONS]bool)
	completedCh := make(chan types.ButtonEvent)

	drv_buttons := make(chan types.ButtonEvent)
	drv_floors := make(chan int) // Floor sensor events.
	drv_obstr := make(chan bool) // Obstruction switch events.
	drv_stop := make(chan bool)  // Stop button events.

	// Start goroutines to poll elevator inputs.
	go driver.PollButtons(drv_buttons) // Check the double Poll Buttons
	go driver.PollFloorSensor(drv_floors)
	go driver.PollObstructionSwitch(drv_obstr)
	go driver.PollStopButton(drv_stop)

	messageTx := make(chan types.NetworkMessage) // Outgoing messages
	messageRx := make(chan types.NetworkMessage) // Incoming messages
	peerUpdateCh := make(chan peers.PeerUpdate)  // Peer discovery updates

	go peers.Transmitter(types.PEER_PORT, localID, nil) // Sends presence on peer network
	go peers.Receiver(types.PEER_PORT, peerUpdateCh)    // Receives peer updates
	go bcast.Transmitter(types.MSG_PORT, messageTx)     // Broadcasts messages
	go bcast.Receiver(types.MSG_PORT, messageRx)        // Receives broadcast messages

	go request_control.RequestHandler(localID, requestsCh, completedCh, buttonEventCh, messageTx, messageRx, peerUpdateCh) // Handles requests
	elevator_controller.ElevatorHandler(drv_buttons, drv_floors, drv_obstr)                                                // Handles elevator control
}
