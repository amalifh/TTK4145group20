package elevatorController

import "Driver-go/elevator/types"

// Struct to hold the direction and behavior pair
type DirnBehaviourPair struct {
	Dirn      types.ElevDirection
	Behaviour types.ElevBehaviour
}

// Check if there are any requests for floors above the current floor
func RequestsAbove(e types.Elevator) bool {
	for f := e.Floor + 1; f < types.N_FLOORS; f++ {
		for btn := 0; btn < types.N_BUTTONS; btn++ {
			if e.Requests[f][btn] {
				return true // Return true if any request exists above
			}
		}
	}
	return false // No requests above
}

// Check if there are any requests for floors below the current floor
func RequestsBelow(e types.Elevator) bool {
	for f := 0; f < e.Floor; f++ {
		for btn := 0; btn < types.N_BUTTONS; btn++ {
			if e.Requests[f][btn] {
				return true // Return true if any request exists below
			}
		}
	}
	return false // No requests below
}

// Check if there are any requests for the current floor
func RequestsHere(e types.Elevator) bool {
	for btn := 0; btn < types.N_BUTTONS; btn++ {
		if e.Requests[e.Floor][btn] {
			return true // Return true if any request exists at the current floor
		}
	}
	return false // No requests at the current floor
}

// Determine the next movement direction and behavior based on the requests
func RequestsChooseDirection(e types.Elevator) DirnBehaviourPair {
	switch e.Direction {
	case types.ED_Up:
		// If moving up, check if there are requests above
		if RequestsAbove(e) {
			return DirnBehaviourPair{types.ED_Up, types.EB_Moving}
		} else if RequestsHere(e) {
			// If there are requests at the current floor, open doors
			return DirnBehaviourPair{types.ED_Down, types.EB_DoorOpen}
		} else if RequestsBelow(e) {
			// If no requests above, but requests below, move down
			return DirnBehaviourPair{types.ED_Down, types.EB_Moving}
		}
	case types.ED_Down:
		// If moving down, check if there are requests below
		if RequestsBelow(e) {
			return DirnBehaviourPair{types.ED_Down, types.EB_Moving}
		} else if RequestsHere(e) {
			// If there are requests at the current floor, open doors
			return DirnBehaviourPair{types.ED_Up, types.EB_DoorOpen}
		} else if RequestsAbove(e) {
			// If no requests below, but requests above, move up
			return DirnBehaviourPair{types.ED_Up, types.EB_Moving}
		}
	case types.ED_Stop:
		// If stopped, check if there are requests at the current floor
		if RequestsHere(e) {
			return DirnBehaviourPair{types.ED_Stop, types.EB_DoorOpen}
		} else if RequestsAbove(e) {
			// If no requests at the current floor, but requests above, move up
			return DirnBehaviourPair{types.ED_Up, types.EB_Moving}
		} else if RequestsBelow(e) {
			// If no requests at the current floor, but requests below, move down
			return DirnBehaviourPair{types.ED_Down, types.EB_Moving}
		}
	}
	// Default: stop the elevator if no requests
	return DirnBehaviourPair{types.ED_Stop, types.EB_Idle}
}

// Check if the elevator should stop at the current floor based on requests
func RequestsShouldStop(e types.Elevator) bool {
	switch e.Direction {
	case types.ED_Down:
		// If moving down, stop if there are requests or no further requests below
		return e.Requests[e.Floor][types.BT_HallDown] ||
			e.Requests[e.Floor][types.BT_Cab] ||
			!RequestsBelow(e)

	case types.ED_Up:
		// If moving up, stop if there are requests or no further requests above
		return e.Requests[e.Floor][types.BT_HallUp] ||
			e.Requests[e.Floor][types.BT_Cab] ||
			!RequestsAbove(e)

	case types.ED_Stop:
		// If stopped, always return true
		fallthrough
	default:
		return true
	}
}

// Check if the current request should be cleared immediately
func RequestsShouldClearImmediately(e types.Elevator, btnFloor int, btnType types.ButtonType) bool {
	switch e.Config.ClearRequestVariant {
	case types.CV_All:
		// If clearing all requests, clear immediately if on the requested floor
		return e.Floor == btnFloor
	case types.CV_InDirn:
		// If clearing in direction, only clear if on requested floor and in the correct direction
		return e.Floor == btnFloor &&
			((e.Direction == types.ED_Up && btnType == types.BT_HallUp) ||
				(e.Direction == types.ED_Down && btnType == types.BT_HallDown) ||
				e.Direction == types.ED_Stop ||
				btnType == types.BT_Cab)
	default:
		// No immediate clearing for other configurations
		return false
	}
}

// Clear requests at the current floor based on the configured clearing behavior
func RequestsClearAtCurrentFloor(e types.Elevator) types.Elevator {
	switch e.Config.ClearRequestVariant {
	case types.CV_All:
		// Clear all requests at the current floor
		for btn := 0; btn < types.N_BUTTONS; btn++ {
			e.Requests[e.Floor][btn] = false
		}
	case types.CV_InDirn:
		// Clear requests at the current floor but only for the cab or direction-specific buttons
		e.Requests[e.Floor][types.BT_Cab] = false
		switch e.Direction {
		case types.ED_Up:
			// If moving up, clear down request if no requests above
			if !RequestsAbove(e) && !e.Requests[e.Floor][types.BT_HallUp] {
				e.Requests[e.Floor][types.BT_HallDown] = false
			}
			e.Requests[e.Floor][types.BT_HallUp] = false
		case types.ED_Down:
			// If moving down, clear up request if no requests below
			if !RequestsBelow(e) && !e.Requests[e.Floor][types.BT_HallDown] {
				e.Requests[e.Floor][types.BT_HallUp] = false
			}
			e.Requests[e.Floor][types.BT_HallDown] = false
		case types.ED_Stop:
			fallthrough
		default:
			// Clear both up and down requests for all directions
			e.Requests[e.Floor][types.BT_HallUp] = false
			e.Requests[e.Floor][types.BT_HallDown] = false
		}
	}
	return e
}