package controller

import (
	"Driver-go/elevator/driver"
	"Driver-go/elevator/types"
	"fmt"
	"time"
)

const (
	DOOR_OPEN_TIME   = 3 * time.Second
	MOBILITY_TIMEOUT = 10 * time.Second
)

type FsmChannels struct {
	RequestsComplete chan int
	Elevator         chan types.ElevInfo
	NewRequest       chan types.ButtonEvent
	ArrivedAtFloor   chan int
	ObstructionChan  chan bool
}

// ElevatorHandler manages the elevator's behavior and state.
func ElevatorHandler(ch FsmChannels) {
	elevator := types.ElevInfo{
		State:         types.EB_Idle,
		Dir:           types.ED_Stop,
		Floor:         driver.GetFloor(),
		RequestsQueue: [types.N_FLOORS][types.N_BUTTONS]bool{},
		CV:            types.CV_All, // Should we pass it as a parameter?
	}

	// Create the timers
	doorTimer := time.NewTimer(DOOR_OPEN_TIME)
	mobilityTimer := time.NewTimer(MOBILITY_TIMEOUT)
	doorTimer.Stop()
	mobilityTimer.Stop()
	requestCleared := false
	obstructionActive := false
	pendingStop := false
	ch.Elevator <- elevator

	// Main event loop
	for {
		select {
		case newRequest := <-ch.NewRequest:
			if newRequest.Done {
				elevator.RequestsQueue[newRequest.Floor][types.BT_Up] = false
				elevator.RequestsQueue[newRequest.Floor][types.BT_Down] = false
				requestCleared = true
			} else {
				elevator.RequestsQueue[newRequest.Floor][newRequest.Btn] = true
			}

			switch elevator.State {
			case types.EB_Idle:
				elevator.Dir = chooseDirection(elevator)
				driver.SetMotorDirection(DirectionConverter(elevator.Dir))
				if elevator.Dir == types.ED_Stop {
					elevator.State = types.EB_DoorOpen
					driver.SetDoorOpenLamp(true)
					doorTimer.Reset(DOOR_OPEN_TIME)
					go func() { ch.RequestsComplete <- newRequest.Floor }()
					elevator.RequestsQueue[elevator.Floor] = [types.N_BUTTONS]bool{}
				} else {
					elevator.State = types.EB_Moving
					mobilityTimer.Reset(MOBILITY_TIMEOUT)
				}

			case types.EB_Moving:
				fallthrough
			case types.EB_DoorOpen:
				if elevator.Floor == newRequest.Floor {
					doorTimer.Reset(DOOR_OPEN_TIME)
					go func() { ch.RequestsComplete <- newRequest.Floor }()
					elevator.RequestsQueue[elevator.Floor] = [types.N_BUTTONS]bool{}
				}

			case types.EB_Undefined:
			default:
				fmt.Println("Fatal error: Reboot system")
			}
			ch.Elevator <- elevator

		case elevator.Floor = <-ch.ArrivedAtFloor:
			fmt.Println("Arrived at floor", elevator.Floor)
			if pendingStop {
				pendingStop = false
				driver.SetDoorOpenLamp(true)
				mobilityTimer.Stop()
				elevator.State = types.EB_DoorOpen
				// Ensure motor is stopped.
				driver.SetMotorDirection(types.MD_Stop)
				doorTimer.Reset(DOOR_OPEN_TIME)
				elevator = clearRequests(elevator, elevator.Floor)
				go func() { ch.RequestsComplete <- elevator.Floor }()
			} else if obstructionActive {
				// When an obstruction is active, simply maintain door open.
				driver.SetDoorOpenLamp(true)
				doorTimer.Reset(DOOR_OPEN_TIME)
				elevator.State = types.EB_DoorOpen
			}else if shouldStop(elevator) ||
				(!shouldStop(elevator) && elevator.RequestsQueue == [types.N_FLOORS][types.N_BUTTONS]bool{} && requestCleared) {
				requestCleared = false
				driver.SetDoorOpenLamp(true)
				mobilityTimer.Stop()
				elevator.State = types.EB_DoorOpen
				driver.SetMotorDirection(types.MD_Stop)
				doorTimer.Reset(DOOR_OPEN_TIME)
				elevator = clearRequests(elevator, elevator.Floor)
				go func() { ch.RequestsComplete <- elevator.Floor }()

			} else if elevator.State == types.EB_Moving {
				mobilityTimer.Reset(3 * time.Second)
			}
			ch.Elevator <- elevator

		// Not working
		case obstructed := <-ch.ObstructionChan:
			fmt.Printf("Obstruction Event: %+v\n", obstructed)
			if obstructed {
				// Set the obstruction flag.
				obstructionActive = true
				// If the elevator is moving, check if it's between floors.
				if elevator.State == types.EB_Moving {
					if driver.GetFloor() == -1 {
						// Elevator is between floors; set pending stop.
						pendingStop = true
						// Let the elevator continue moving.
					} else {
						// Elevator is at a valid floor; stop immediately.
						driver.SetMotorDirection(types.MD_Stop)
						driver.SetDoorOpenLamp(true)
						elevator.State = types.EB_DoorOpen
						doorTimer.Reset(DOOR_OPEN_TIME)
						mobilityTimer.Stop()
						elevator = clearRequests(elevator, elevator.Floor)
						go func() { ch.RequestsComplete <- elevator.Floor }()
					}
				} else if elevator.State == types.EB_DoorOpen {
					// Already stopped with doors open; ensure door remains open.
					driver.SetDoorOpenLamp(true)
					doorTimer.Reset(DOOR_OPEN_TIME)
				}
			} else {
				// Obstruction cleared.
				obstructionActive = false
				// If we are in door-open state and no pending stop, close door and resume normal operation.
				if elevator.State == types.EB_DoorOpen {
					doorTimer.Stop()
					driver.SetDoorOpenLamp(false)
					elevator.Dir = chooseDirection(elevator)
					if elevator.Dir == types.ED_Stop {
						elevator.State = types.EB_Idle
						mobilityTimer.Stop()
					} else {
						elevator.State = types.EB_Moving
						mobilityTimer.Reset(MOBILITY_TIMEOUT)
						driver.SetMotorDirection(DirectionConverter(elevator.Dir))
					}
				}
			}
			ch.Elevator <- elevator

			case <-doorTimer.C:
				if obstructionActive {
					doorTimer.Reset(DOOR_OPEN_TIME)
					break
				}
				driver.SetDoorOpenLamp(false)
				elevator.Dir = chooseDirection(elevator)
				if elevator.Dir == types.ED_Stop {
					elevator.State = types.EB_Idle
					mobilityTimer.Stop()
				} else {
					elevator.State = types.EB_Moving
					mobilityTimer.Reset(MOBILITY_TIMEOUT)
					driver.SetMotorDirection(DirectionConverter(elevator.Dir))
				}
				ch.Elevator <- elevator

			case <-mobilityTimer.C:
				driver.SetMotorDirection(types.MD_Stop)
				elevator.State = types.EB_Undefined
				fmt.Println("\x1b[1;1;33m", "Engine Error - Go offline", "\x1b[0m")
				for i := 0; i < 10; i++ {
					if i%2 == 0 {
						driver.SetStopLamp(true)
					} else {
						driver.SetStopLamp(false)
					}
					time.Sleep(time.Millisecond * 200)
				}
				driver.SetMotorDirection(DirectionConverter(elevator.Dir))
				ch.Elevator <- elevator
				mobilityTimer.Reset(MOBILITY_TIMEOUT)
		}
	}
}
