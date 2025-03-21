package localController

import (
	"Driver-go/elevator/driver"
	"Driver-go/elevator/types"
	"sync"
)

// DirnBehaviourPair holds a direction and behaviour pair.
type DirnBehaviourPair struct {
	Drn       types.ElevDirection
	Behaviour types.ElevBehaviour
}

// CurrentElevator holds the global elevator state.
// It is updated by all FSM functions.
var CurrentElevator types.Elevator
var mtx sync.Mutex

// DirectionConverter converts an elevator direction to a motor command.
func DirectionConverter(dir types.ElevDirection) types.MotorDirection {
	switch dir {
	case types.ED_Up:
		return types.MD_Up
	case types.ED_Down:
		return types.MD_Down
	case types.ED_Stop:
		return types.MD_Stop
	}
	return types.MD_Stop
}

func updateButtonLamps(floor int, requests [types.N_BUTTONS]bool) {
	for btn, active := range requests {
		driver.SetButtonLamp(types.ButtonType(btn), floor, active)
	}
}

// OnRequestButtonPress handles a button press event.
func OnRequestButtonPress(floor int, btn int) {
	mtx.Lock()
	defer mtx.Unlock()
	CurrentElevator.Requests[floor][btn] = true
	if CurrentElevator.Behaviour == types.EB_Idle {
		pair := chooseDirection(CurrentElevator)
		CurrentElevator.Direction = pair.Drn
		CurrentElevator.Behaviour = pair.Behaviour
		driver.SetMotorDirection(DirectionConverter(pair.Drn))
	}
}

// OnFloorArrival handles a floor sensor event.
func OnFloorArrival(floor int) {
	CurrentElevator.Floor = floor
	if CurrentElevator.Behaviour == types.EB_Moving && shouldStop(CurrentElevator) {
		CurrentElevator.Behaviour = types.EB_DoorOpen
		driver.SetMotorDirection(types.MD_Stop)
		driver.SetDoorOpenLamp(true)
		CurrentElevator = clearRequestsAtCurrentFloor(CurrentElevator)
	}
	updateButtonLamps(CurrentElevator.Floor, CurrentElevator.Requests[CurrentElevator.Floor])
}

// OnObstruction handles an obstruction event.
func OnObstruction(obstructed bool) bool {
	if obstructed && CurrentElevator.Behaviour == types.EB_DoorOpen {
		return true
	}
	return false
}

// OnDoorTimeout handles the door timeout event.
func OnDoorTimeout() {
	CurrentElevator = clearRequestsAtCurrentFloor(CurrentElevator)
	pair := chooseDirection(CurrentElevator)
	CurrentElevator.Direction = pair.Drn
	CurrentElevator.Behaviour = pair.Behaviour
	driver.SetMotorDirection(DirectionConverter(pair.Drn))
}

// OnMobilityTimeout handles the mobility timeout event.
func OnMobilityTimeout() {
	CurrentElevator.Behaviour = types.EB_Idle
}

// --- Internal helper functions --- //

func chooseDirection(e types.Elevator) DirnBehaviourPair {
	switch e.Direction {
	case types.ED_Up:
		if requestsAbove(e) {
			return DirnBehaviourPair{types.ED_Up, types.EB_Moving}
		} else if requestsHere(e) {
			return DirnBehaviourPair{types.ED_Down, types.EB_DoorOpen}
		} else if requestsBelow(e) {
			return DirnBehaviourPair{types.ED_Down, types.EB_Moving}
		}
	case types.ED_Down:
		if requestsBelow(e) {
			return DirnBehaviourPair{types.ED_Down, types.EB_Moving}
		} else if requestsHere(e) {
			return DirnBehaviourPair{types.ED_Up, types.EB_DoorOpen}
		} else if requestsAbove(e) {
			return DirnBehaviourPair{types.ED_Up, types.EB_Moving}
		}
	case types.ED_Stop:
		if requestsHere(e) {
			return DirnBehaviourPair{types.ED_Stop, types.EB_DoorOpen}
		} else if requestsAbove(e) {
			return DirnBehaviourPair{types.ED_Up, types.EB_Moving}
		} else if requestsBelow(e) {
			return DirnBehaviourPair{types.ED_Down, types.EB_Moving}
		}
	}
	return DirnBehaviourPair{types.ED_Stop, types.EB_Idle}
}

func requestsAbove(e types.Elevator) bool {
	for f := e.Floor + 1; f < types.N_FLOORS; f++ {
		for btn := 0; btn < types.N_BUTTONS; btn++ {
			if e.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func requestsBelow(e types.Elevator) bool {
	for f := 0; f < e.Floor; f++ {
		for btn := 0; btn < types.N_BUTTONS; btn++ {
			if e.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func requestsHere(e types.Elevator) bool {
	for btn := 0; btn < types.N_BUTTONS; btn++ {
		if e.Requests[e.Floor][btn] {
			return true
		}
	}
	return false
}

func shouldStop(e types.Elevator) bool {
	switch e.Direction {
	case types.ED_Down:
		return e.Requests[e.Floor][types.BT_HallDown] ||
			e.Requests[e.Floor][types.BT_Cab] ||
			!requestsBelow(e)
	case types.ED_Up:
		return e.Requests[e.Floor][types.BT_HallUp] ||
			e.Requests[e.Floor][types.BT_Cab] ||
			!requestsAbove(e)
	case types.ED_Stop:
		fallthrough
	default:
		return true
	}
}

func clearRequestsAtCurrentFloor(e types.Elevator) types.Elevator {
	switch e.Config.ClearRequestVariant {
	case types.CV_All:
		for btn := 0; btn < types.N_BUTTONS; btn++ {
			e.Requests[e.Floor][btn] = false
		}
	case types.CV_InDirn:
		// Always clear the cab button.
		e.Requests[e.Floor][types.BT_Cab] = false

		// Instead of conditionally clearing hall buttons based on direction,
		// clear both hall requests to ensure that no pending requests remain.
		e.Requests[e.Floor][types.BT_HallUp] = false
		e.Requests[e.Floor][types.BT_HallDown] = false

		// (Optionally, if some directional filtering is desired in other contexts,
		// additional logic can be added here—but for clearing at the floor, it's safer to clear both.)
	}
	return e
}
