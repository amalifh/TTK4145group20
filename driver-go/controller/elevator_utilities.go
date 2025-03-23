package elevatorStateMachine

import (
	"Driver-go/elevator/types"
)

func shouldStop(elevator ElevInfo) bool {
	switch elevator.Dir {
	case types.ED_Up:
		return elevator.Queue[elevator.Floor][BtnUp] ||
			elevator.Queue[elevator.Floor][BtnInside] ||
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

func chooseDirection(elevator Elev) Direction {
	switch elevator.Dir {
	case types.ED_Stop:
		if ordersAbove(elevator) {
			return types.ED_Up
		} else if ordersBelow(elevator) {
			return types.ED_Down
		} else {
			return types.ED_Stop
		}

	case types.ED_Up:
		if ordersAbove(elevator) {
			return types.ED_Up
		} else if ordersBelow(elevator) {
			return types.ED_Down
		} else {
			return types.ED_Stop
		}

	case types.ED_Down:
		if ordersBelow(elevator) {
			return types.ED_Down
		} else if ordersAbove(elevator) {
			return types.ED_Up
		} else {
			return types.ED_Stop
		}
	}
	return types.ED_Stop
}

func isRequestAbove(elevator Elev) bool {
	for floor := elevator.Floor + 1; floor < NumFloors; floor++ {
		for btn := 0; btn < NumButtons; btn++ {
			if elevator.Queue[floor][btn] {
				return true
			}
		}
	}
	return false
}

func isRequestBelow(elevator Elev) bool {
	for floor := 0; floor < elevator.Floor; floor++ {
		for btn := 0; btn < NumButtons; btn++ {
			if elevator.Queue[floor][btn] {
				return true
			}
		}
	}
	return false
}

func clearRequestsAtCurrentFloor(e types.ElevInfo) types.ElevInfo {
	switch e.CV {
	case types.CV_All:
		for btn := 0; btn < types.N_BUTTONS; btn++ {
			e.RequestsQueue[e.Floor][btn] = false
		}
	case types.CV_InDirn:
		// Always clear the cab button.
		e.RequestsQueue[e.Floor][types.BT_Cab] = false

		// Instead of conditionally clearing hall buttons based on direction,
		// clear both hall requests to ensure that no pending requests remain.
		e.RequestsQueue[e.Floor][types.BT_Up] = false
		e.RequestsQueue[e.Floor][types.BT_Down] = false

		// (Optionally, if some directional filtering is desired in other contexts,
		// additional logic can be added here—but for clearing at the floor, it's safer to clear both.)
	}
	return e
}
