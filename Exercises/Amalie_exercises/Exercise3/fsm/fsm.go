package fsm

import (
	"Exercise3/types"
	"Exercise3/requests"
	"Exercise3/timer"
	"fmt"
)

//starting of by just initializing the elevator
var elevState = types.elevator_uninitialized()

//the whole point of this is to start the processes, having the functions actually working together!
func setAllLights(e types.Elevator) {
	for floor := 0; floor < types.N_floors; floor++ {
		for btn := 0; btn < types.N_buttons; btn++ {
			driver.SetButtonLamp(types.ButtonType(btn), floor, e.E_requests[floor][btn])
		}
	}
}

func fsm_initBetweenFloors() {
	driver.SetMotorDirection(types.Dirn_Down)
	elevState.E_dirn = types.Dirn_Down
	elevState.E_behaviour = types.EB_Moving
}

//function to how the elevator should handle button presses
func fsm_requestButtonPress(btnFloor int, btnType types.Button_type) {
	//if obstruction is on or stop is pressed the elevator is to do nothing
	if (elevState.E_obstruction || elevState.E_stop) {
		return
	} 
	switch elevState.E_behaviour {
	case types.EB_DoorOpen:
		if(request.requests_shouldClearImmediately(elevState, btnFloor, btnType)){
			timer.timerStart(elevState.Config.doorOpen_s)
		} else {
			elevState.E_requests[btnFloor][btnType] = true
		}
	case types.EB_Moving:
		elevState.E_requests[btnFloor][btnType] = true

	case types.EB_idle:
		elevState.E_requests[btnFloor][btnType] = true
		pair := requests_chooseDirn(elevState)
		elevState.E_dirn = pair.dp_dirn
		elevState.E_behaviour = pair.behaviour

		switch elevState.E_behaviour {
		case types.EB_DoorOpen:
			driver.SetDoorOpenLamp(true)
			timer.timerStart(elevState.Config.doorOpen_s)
			elevState = requests.requests_clearAtCurrentFloor(elevState)
			
		case types.EB_Moving:
			driver.SetMotorDirection(elevState.E_dirn)
		
		}


	}
	setAllLights(elevState)
}

func fsm_floorArrival(newFloor int) {
	elevState.E_floor = newFloor

	driver.SetFloorIndicator(elevState.E_floor)

	//wanting to check if elevator should stop at the new floor when moving
	switch elevState.E_behaviour {
	case types.EB_Moving:
		if requests.requests_shouldStop(elevState) {
			driver.SetMotorDirection(types.Dirn_Stop)
			driver.SetDoorOpenLamp(true)
			elevState = requests.requests_clearAtCurrentFloor(elevState)
			timer.timerStart(elevState.Config.doorOpen_s)
			setAllLights(elevState)
			elevState.E_behaviour = types.EB_DoorOpen
		}
	}
}

func fsm_obstruction(obstructed bool) {
	if obstructed {
		elevState.E_obstruction = true
		switch elevState.E_behaviour {
		case types.EB_DoorOpen:
			driver.SetMotorDirection(types.Dirn_Stop)
			driver.SetDoorOpenLamp(true)
			setAllLights(elevState)
		}
	} else {
		if elevState.E_obstruction {
			setAllLights(elevState)
			elevState.E_obstruction = false
			if elevState.E_behaviour == types.EB_Moving {
				driver.SetMotorDirection(elevState.Dirn)
			} else {
				driver.SetDoorOpenLamp(true)
				timer.TimerStart(elevState.Config.doorOpen_s)
			}
		}
	}
}

func fsm_stop() {
	if !elevState.E_stop {
		elevState.E_stop = true
		driver.SetMotorDirection(elevator.D_Stop)
		driver.SetStopLamp(true)
	} else {
		elevState.E_stop = false
		driver.SetStopLamp(false)
		if elevState.E_obstruction {
			return
		}
		driver.SetMotorDirection(elevState.Dirn)
	}
}

func fsm_doorTimeout() {
	switch elevState.E_behaviour {
	case types.EB_DoorOpen:
		pair := requests.requests_chooseDirn(elevState)
		elevState.E_dirn = pair.dp_dirn
		elevState.E_behaviour = pair.behaviour

		switch elevState.E_behaviour {
		case types.EB_DoorOpen:
			timer.timerStart(elevState.Config.doorOpen_s)
			elevState = requests_clearAtCurrentFloor(elevState)
			setAllLights(elevState)
		case types.EB_idle, types.EB_Moving:
			driver.SetDoorOpenLamp(false)
			driver.SetMotorDirection(elevState.E_dirn)

		}


	}
}