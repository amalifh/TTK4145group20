package main

import (
	"Driver-go/elevio"
	"fmt"
)

func main() {

	numFloors := 4
	var currentFloor int
	//var buttonKind elevio.ButtonType
	var requestQueue []elevio.ButtonEvent // List of requested floors

	elevio.Init("localhost:15657", numFloors)

	var d elevio.MotorDirection = elevio.MD_Down
	elevio.SetMotorDirection(d)
	var initQueue elevio.ButtonEvent
	initQueue.Button = 2
	initQueue.Floor = 0
	requestQueue = append(requestQueue, initQueue)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	for {
		select {
		case a := <-drv_buttons:
			// If the requested floor is the current floor, just turn on the lamp and ignore it
			if a.Floor == currentFloor && d == elevio.MD_Stop {
				fmt.Println("Already at requested floor, no movement needed.")
				continue
			}

			if len(requestQueue) == 0 && currentFloor != a.Floor {
				requestQueue = append(requestQueue, a)
			} else if currentFloor != a.Floor {
				// Add the floor to the queue if it's not already there
				for i := 0; i < len(requestQueue); i++ {
					if requestQueue[i] != a {
						requestQueue = append(requestQueue, a)
					}
				}
			}
			//buttonKind = a.Button

			fmt.Printf("Length: %d\n", len(requestQueue))

			// Decide the motor direction based on the first request in queue
			if len(requestQueue) > 0 {
				if currentFloor > requestQueue[0].Floor {
					d = elevio.MD_Down
				} else if currentFloor < requestQueue[0].Floor {
					d = elevio.MD_Up
				} else {
					requestQueue = requestQueue[1:]
				}
			}

			fmt.Printf("Button pressed: %+v\n", a)
			elevio.SetButtonLamp(a.Button, a.Floor, true)
			elevio.SetMotorDirection(d)

		case a := <-drv_floors:
			currentFloor = a
			fmt.Printf("Reached floor: %d\n", currentFloor)

			// Check if we reached the target floor
			if len(requestQueue) > 0 && currentFloor == requestQueue[0].Floor {
				d = elevio.MD_Stop
				elevio.SetMotorDirection(d)
				elevio.SetButtonLamp(requestQueue[0].Button, requestQueue[0].Floor, false)
				elevio.SetButtonLamp(2, requestQueue[0].Floor, false)

				// Remove the floor from the queue
				requestQueue = requestQueue[1:]
				fmt.Printf("Length: %d\n", len(requestQueue))

				// Move to the next request
				if len(requestQueue) > 0 {
					if currentFloor > requestQueue[0].Floor {
						d = elevio.MD_Down
					} else if currentFloor < requestQueue[0].Floor {
						d = elevio.MD_Up
					}
					elevio.SetMotorDirection(d)
				}
			}

		case a := <-drv_obstr:
			fmt.Printf("Obstruction: %+v\n", a)
			if a {
				elevio.SetMotorDirection(elevio.MD_Stop)
			} else if len(requestQueue) > 0 {
				// Resume movement
				if currentFloor > requestQueue[0].Floor {
					d = elevio.MD_Down
				} else if currentFloor < requestQueue[0].Floor {
					d = elevio.MD_Up
				}
				elevio.SetMotorDirection(d)
			}

		case a := <-drv_stop:
			fmt.Printf("%+v\n", a)
			for f := 0; f < numFloors; f++ {
				for b := elevio.ButtonType(0); b < 3; b++ {
					elevio.SetButtonLamp(b, f, false)
				}
			}
		}
	}
}

// Helper function to check if a floor is already in the queue
func contains(queue []int, floor int) bool {
	for _, f := range queue {
		if f == floor {
			return true
		}
	}
	return false
}
