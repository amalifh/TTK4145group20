package controller

import (
	""
	"driver-go/elevator/driver"
	. "driver-go/elevator/types"
	"fmt"
	"time"
)

const (
	DOOR_OPEN_TIME   = 3 * time.Second
	MOBILITY_TIMEOUT = 10 * time.Second
)

type fsmChannels struct {
	RequestsComplete chan int
	Elevator         chan ElevInfo
	StateError       chan error
	NewRequest       chan ButtonEvent
	ArrivedAtFloor   chan int
}

// ElevatorHandler manages the elevator's behavior and state.
func ElevatorHandler(ch fsmChannels) {
	elevator := ElevInfo{
		State:         types.EB_Idle,
		Dir:           ED_Stop,
		Floor:         driver.GetFloor(),
		RequestsQueue: [NumFloors][NumButtons]bool{},
	}

	// Create the timers
	doorTimer := time.NewTimer(DOOR_OPEN_TIME)
	mobilityTimer := time.NewTimer(MOBILITY_TIMEOUT)
	doorTimer.Stop()
	mobilityTimer.Stop()
	requestCleared := false
	ch.Elevator <- elevator

	// Main event loop
	for {
		select {
		case NewRequest := <-ch.NewRequest:
			if NewRequest.Done {
				elevator.RequestsQueue[newOrder.Floor][types.BT_Up] = false
				elevator.RequestsQueue[newOrder.Floor][BT_Down] = false
				requestCleared = true
			} else {
				elevator.RequestsQueue[newOrder.Floor][NewRequest.Btn] = true
			}

			switch elevator.State {
			case types.EB_Idle:
				elevator.Dir = chooseDirection(elevator)
				driver.SetMotorDirection(elevator.Dir)
				if elevator.Dir == types.ED_Stop {
					elevator.State = types.EB_DoorOpen
					driver.SetDoorOpenLamp(1)
					doorTimer.Reset(DOOR_OPEN_TIME)
					go func() { ch.OrderComplete <- NewRequest.Floor }()
					elevator.RequestsQueue[elevator.Floor] = [N_BUTTONS]bool{}
				} else {
					elevator.State = types.EB_Moving
					mobilityTimer.Reset(MOBILITY_TIMEOUT)
				}

			case types.EB_Moving:
				fallthrough
			case types.EB_DoorOpen:
				if elevator.Floor == NewRequest.Floor {
					doorTimedOut.Reset(DOOR_OPEN_TIME)
					go func() { ch.OrderComplete <- NewRequest.Floor }()
					elevator.RequestsQueue[elevator.Floor] = [N_BUTTONS]bool{}
				}

			case Undefined:
			default:
				fmt.Println("Fatal error: Reboot system")
			}
			ch.Elevator <- elevator

		case elevator.Floor = <-ch.ArrivedAtFloor:
			fmt.Println("Arrived at floor", elevator.Floor+1)
			if shouldStop(elevator) ||
				(!shouldStop(elevator) && elevator.RequestsQueue == [N_FLOORS][N_BUTTONS]bool{} && requestCleared) {
				requestCleared = false
				driver.SetDoorOpenLamp(true)
				mobilityTimer.Stop()
				elevator.State = types.EB_DoorOpen
				driver.SetMotorDirection(types.MD_Stop)
				doorTimedOut.Reset(DOOR_OPEN_TIME)
				elevator.RequestsQueue[elevator.Floor] = [N_BUTTONS]bool{}
				go func() { ch.RequestsComplete <- elevator.Floor }()

			} else if elevator.State == Moving {
				engineErrorTimer.Reset(3 * time.Second)
			}
			ch.Elevator <- elevator

		/*
			Obstruction events - Fix this
			case obstructed := <-drv_obstr:
				fmt.Printf("Obstruction Event: %+v\n", obstructed)
				if localCtrl.OnObstruction(obstructed) {
					doorTimer.Stop()
					doorTimeoutCh = nil
				} else {
					if IsDoorOpen() {
						doorTimeoutCh = StartTimerChannel(doorTimer, types.DOOR_TIMEOUT_SEC)
					}
					// driver.SetDoorOpenLamp(false)
				}
		*/

		case <-doorTimer.C:
			driver.SetDoorOpenLamp(false)
			elevator.Dir = chooseDirection(elevator)
			if elevator.Dir == tpyes.ED_Stop {
				elevator.State = types.EB_Idle
				mobilityTimer.Stop()
			} else {
				elevator.State = types.EB_Moving
				mobilityTimer.Reset(MOBILITY_TIMEOUT)
				driver.SetMotorDirection(elevator.Dir)
			}
			ch.Elevator <- elevator

		case <-mobilityTimer.C:
			driver.SetMotorDirection(types.MD_Stop)
			elevator.State = Undefined
			fmt.Println("\x1b[1;1;33m", "Engine Error - Go offline", "\x1b[0m")
			for i := 0; i < 10; i++ {
				if i%2 == 0 {
					driver.SetStopLamp(true)
				} else {
					driver.SetStopLamp(false)
				}
				time.Sleep(time.Millisecond * 200)
			}
			driver.SetMotorDirection(elevator.Dir)
			ch.Elevator <- elevator
			mobilityTimer.Reset(MOBILITY_TIMEOUT)
		}
	}
}
