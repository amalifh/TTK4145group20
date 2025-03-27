/*
Package controller provides utility functions to determine elevator movement, stopping conditions,
and direction conversion for an elevator system.

Functions:
- shouldStop: Determines whether the elevator should stop at the current floor based on requests and direction.
- chooseDirection: Determines the elevator's next movement direction based on pending requests.
- DirectionConverter: Converts an elevator's movement direction into a corresponding motor command.
- isRequestAbove: Checks if there are any active requests above the current floor.
- isRequestBelow: Checks if there are any active requests below the current floor.

These functions support the finite state machine (FSM) logic in managing elevator behavior efficiently.
*/
package controller

import (
	"Driver-go/elevator/types"
)

func shouldStop(elevator types.ElevInfo) bool {
	switch elevator.Dir {
	case types.ED_Up:
		return elevator.RequestsQueue[elevator.Floor][types.BT_Up] ||
			elevator.RequestsQueue[elevator.Floor][types.BT_Cab] ||
			!isRequestAbove(elevator)
	case types.ED_Down:
		return elevator.RequestsQueue[elevator.Floor][types.BT_Down] ||
			elevator.RequestsQueue[elevator.Floor][types.BT_Cab] ||
			!isRequestBelow(elevator)
	case types.ED_Stop:
	default:
	}
	return false
}

func chooseDirection(elevator types.ElevInfo) types.ElevDirection {
	switch elevator.Dir {
	case types.ED_Stop:
		if isRequestAbove(elevator) {
			return types.ED_Up
		} else if isRequestBelow(elevator) {
			return types.ED_Down
		} else {
			return types.ED_Stop
		}

	case types.ED_Up:
		if isRequestAbove(elevator) {
			return types.ED_Up
		} else if isRequestBelow(elevator) {
			return types.ED_Down
		} else {
			return types.ED_Stop
		}

	case types.ED_Down:
		if isRequestBelow(elevator) {
			return types.ED_Down
		} else if isRequestAbove(elevator) {
			return types.ED_Up
		} else {
			return types.ED_Stop
		}
	}
	return types.ED_Stop
}

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

func isRequestAbove(elevator types.ElevInfo) bool {
	for floor := elevator.Floor + 1; floor < types.N_FLOORS; floor++ {
		for btn := 0; btn < types.N_BUTTONS; btn++ {
			if elevator.RequestsQueue[floor][btn] {
				return true
			}
		}
	}
	return false
}

func isRequestBelow(elevator types.ElevInfo) bool {
	for floor := 0; floor < elevator.Floor; floor++ {
		for btn := 0; btn < types.N_BUTTONS; btn++ {
			if elevator.RequestsQueue[floor][btn] {
				return true
			}
		}
	}
	return false
}

func toClearHallDown(elevator types.ElevInfo) bool {
	if !elevator.RequestsQueue[elevator.Floor][types.BT_Down] {
		return false
	}
	switch elevator.Dir {
	case types.ED_Down, types.ED_Stop:
		return true
	case types.ED_Up:
		return !isRequestBelow(elevator) && !elevator.RequestsQueue[elevator.Floor][types.BT_Up]
	}
	return false
}

func toClearHallUp(elevator types.ElevInfo) bool {
	if !elevator.RequestsQueue[elevator.Floor][types.BT_Cab] {
		return false
	}
	switch elevator.Dir {
	case types.ED_Up, types.ED_Stop:
		return true
	case types.ED_Down:
		return !isRequestAbove(elevator) && !elevator.RequestsQueue[elevator.Floor][types.BT_Down]
	}
	return false
}

func toClearCab(elevator types.ElevInfo) bool {
	return elevator.RequestsQueue[elevator.Floor][types.BT_Cab]
}

func clearRequests(elevator types.ElevInfo, floor int) types.ElevInfo {
	switch elevator.CV {
	case types.CV_InDirn:
		// Clear the cab request if the helper indicates it should be cleared.
		if toClearCab(elevator) {
			elevator.RequestsQueue[floor][types.BT_Cab] = false
		}

		// Clear the hall up request if indicated.
		if toClearHallUp(elevator) {
			elevator.RequestsQueue[floor][types.BT_Up] = false
		}

		// Clear the hall down request if indicated.
		if toClearHallDown(elevator) {
			elevator.RequestsQueue[floor][types.BT_Down] = false
		}
	case types.CV_All:
		// Clear all button requests unconditionally.
		for btn := 0; btn < types.N_BUTTONS; btn++ {
			elevator.RequestsQueue[floor][btn] = false
		}
	}
	return elevator
}
