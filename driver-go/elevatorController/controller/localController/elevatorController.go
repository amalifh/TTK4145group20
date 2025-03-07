package localController

import (
	"Driver-go/elevator/types"
)

// DirnBehaviourPair holds a direction and behaviour pair.
type DirnBehaviourPair struct {
	Drn       types.ElevDirection
	Behaviour types.ElevBehaviour
}

// CurrentElevator holds the global elevator state.
// It is updated by all FSM functions.
var CurrentElevator types.Elevator

// OnRequestButtonPress handles a button press event.
func OnRequestButtonPress(floor int, btn int) {
	CurrentElevator.Requests[floor][btn] = true
	if CurrentElevator.Behaviour == types.EB_Idle {
		pair := chooseDirection(CurrentElevator)
		CurrentElevator.Direction = pair.Drn
		CurrentElevator.Behaviour = pair.Behaviour
	}
}

// OnFloorArrival handles a floor sensor event.
func OnFloorArrival(floor int) {
	CurrentElevator.Floor = floor
	if CurrentElevator.Behaviour == types.EB_Moving && shouldStop(CurrentElevator) {
		CurrentElevator.Behaviour = types.EB_DoorOpen
		CurrentElevator = clearRequestsAtCurrentFloor(CurrentElevator)
	}
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
		e.Requests[e.Floor][types.BT_Cab] = false
		switch e.Direction {
		case types.ED_Up:
			if !requestsAbove(e) && !e.Requests[e.Floor][types.BT_HallUp] {
				e.Requests[e.Floor][types.BT_HallDown] = false
			}
			e.Requests[e.Floor][types.BT_HallUp] = false
		case types.ED_Down:
			if !requestsBelow(e) && !e.Requests[e.Floor][types.BT_HallDown] {
				e.Requests[e.Floor][types.BT_HallUp] = false
			}
			e.Requests[e.Floor][types.BT_HallDown] = false
		case types.ED_Stop:
			fallthrough
		default:
			e.Requests[e.Floor][types.BT_HallUp] = false
			e.Requests[e.Floor][types.BT_HallDown] = false
		}
	}
	return e
}
