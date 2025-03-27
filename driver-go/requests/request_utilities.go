/*
Package requests provides utility functions for managing elevator request distribution in a multi-elevator system.

This package includes:
  - `duplicateRequest`: Checks whether a given request already exists in the request queue to avoid redundancy.
  - `calcChosenElevator`: Determines the most suitable elevator to handle a request based on factors like
    current position, direction, and operational state.

The package ensures efficient request assignment while minimizing elevator travel time and optimizing responsiveness.

Credits: https://github.com/perkjelsvik/TTK4145-sanntid
*/
package requests

import (
	"Driver-go/elevator/types"
)

func duplicateRequest(request types.ButtonEvent, elevList [types.N_ELEVATORS]types.ElevInfo, id int) bool {
	if request.Btn == types.BT_Cab && elevList[id].RequestsQueue[request.Floor][types.BT_Cab] {
		return true
	}
	for elevator := 0; elevator < types.N_ELEVATORS; elevator++ {
		if elevList[id].RequestsQueue[request.Floor][request.Btn] {
			return true
		}
	}
	return false
}

func calcChosenElevator(request types.ButtonEvent, elevList [types.N_ELEVATORS]types.ElevInfo, id int, aliveList [types.N_ELEVATORS]bool) int {
	if request.Btn == types.BT_Cab {
		return id
	}

	minCost := (types.N_BUTTONS * types.N_FLOORS) * types.N_ELEVATORS
	bestElevator := id
	for elevator := 0; elevator < types.N_ELEVATORS; elevator++ {
		if !aliveList[elevator] {
			continue
		}
		cost := request.Floor - elevList[elevator].Floor

		if cost == 0 && elevList[elevator].State != types.EB_Moving {
			bestElevator = elevator
			return bestElevator
		}

		if cost < 0 {
			cost = -cost
			if elevList[elevator].Dir == types.ED_Up {
				cost += 3
			}
		} else if cost > 0 {
			if elevList[elevator].Dir == types.ED_Down {
				cost += 3
			}
		}

		if cost == 0 && elevList[elevator].State == types.EB_Moving {
			cost += 4
		}

		if elevList[elevator].State == types.EB_DoorOpen {
			cost++
		}

		if cost < minCost {
			minCost = cost
			bestElevator = elevator
		}
	}
	return bestElevator
}
