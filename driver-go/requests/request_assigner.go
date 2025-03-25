package requests

import (
	"Driver-go/elevator/driver"
	"Driver-go/elevator/types"
	"fmt"
)

func RequestAssigner(
	id int,
	bPressedCh chan types.ButtonEvent, // Button pressed
	lUpdateCh chan [types.N_ELEVATORS]types.ElevInfo, // Lights update
	completedRequestsCh chan int,
	newRequestsCh chan types.ButtonEvent,
	elevatorsCh chan types.ElevInfo,
	updatedRequestsCh chan types.ButtonEvent,
	updateSyncCh chan types.ElevInfo,
	assignerUpdatesCh chan [types.N_ELEVATORS]types.ElevInfo,
	aliveElevatorsCh chan [types.N_ELEVATORS]bool) {
	var (
		elevList          [types.N_ELEVATORS]types.ElevInfo
		aliveList         [types.N_ELEVATORS]bool
		completedRequests types.ButtonEvent
	)
	for {
		select {
		case newLocalRequest := <-bPressedCh:
			// If the request is a cab request, the elevator should handle it
			if !aliveList[id] && newLocalRequest.Btn == types.BT_Cab {
				elevList[id].RequestsQueue[newLocalRequest.Floor][types.BT_Cab] = true
				lUpdateCh <- elevList
				go func() { newRequestsCh <- newLocalRequest }()

				// If the elevator is not alive, and the request is not a cab request, do nothing
			} else if !aliveList[id] && newLocalRequest.Btn != types.BT_Cab {
				continue
				// In any other case the request should be assigned to the best elevator
			} else {
				if newLocalRequest.Floor == elevList[id].Floor && elevList[id].State != types.EB_Moving {
					newRequestsCh <- newLocalRequest
				} else {
					if !duplicateRequest(newLocalRequest, elevList, id) {
						fmt.Println("New request at floor ", newLocalRequest.Floor, " for button ", newLocalRequest.Btn)
						newLocalRequest.ChosenElevator = calcChosenElevator(newLocalRequest, elevList, id, aliveList)
						updatedRequestsCh <- newLocalRequest
					}
				}
			}

		case completedRequests.Floor = <-completedRequestsCh:
			completedRequests.Done = true
			for btn := types.BT_Up; btn < types.N_BUTTONS; btn++ {
				if elevList[id].RequestsQueue[completedRequests.Floor][btn] {
					completedRequests.Btn = btn
				}
				for elevator := 0; elevator < types.N_ELEVATORS; elevator++ {
					if btn != types.BT_Cab || elevator == id {
						elevList[elevator].RequestsQueue[completedRequests.Floor][btn] = false
					}
				}
			}

			if aliveList[id] {
				updatedRequestsCh <- completedRequests
			}
			lUpdateCh <- elevList

		case newElev := <-elevatorsCh:
			tmpQueue := elevList[id].RequestsQueue
			if elevList[id].State == types.EB_Undefined && newElev.State != types.EB_Undefined {
				aliveList[id] = true
			}
			elevList[id] = newElev
			elevList[id].RequestsQueue = tmpQueue
			if aliveList[id] {
				updateSyncCh <- elevList[id]
			}

		case copyOnlineList := <-aliveElevatorsCh:
			aliveList = copyOnlineList

		case tmpElevList := <-assignerUpdatesCh:
			newRequest := false
			for elevator := 0; elevator < types.N_ELEVATORS; elevator++ {
				if elevator == id {
					continue
				}
				if elevList[elevator].RequestsQueue != tmpElevList[elevator].RequestsQueue {
					newRequest = true
				}
				elevList[elevator] = tmpElevList[elevator]
			}

			for floor := 0; floor < types.N_FLOORS; floor++ {
				for btn := types.BT_Up; btn < types.N_BUTTONS; btn++ {
					if tmpElevList[id].RequestsQueue[floor][btn] && !elevList[id].RequestsQueue[floor][btn] {
						elevList[id].RequestsQueue[floor][btn] = true
						request := types.ButtonEvent{Floor: floor, Btn: btn, ChosenElevator: id, Done: false}
						go func() { newRequestsCh <- request }()
						newRequest = true
					} else if !tmpElevList[id].RequestsQueue[floor][btn] && elevList[id].RequestsQueue[floor][btn] {
						elevList[id].RequestsQueue[floor][btn] = false
						request := types.ButtonEvent{Floor: floor, Btn: btn, ChosenElevator: id, Done: true}
						go func() { newRequestsCh <- request }()
						newRequest = true
					}
				}
			}

			if newRequest {
				lUpdateCh <- elevList
			}
		}
	}
}

func LightsUpdater(lUpdateCh <-chan [types.N_ELEVATORS]types.ElevInfo, id int) {
	var RequestExists [types.N_ELEVATORS]bool

	for {
		elevs := <-lUpdateCh
		for floor := 0; floor < types.N_FLOORS; floor++ {
			for btn := types.BT_Up; btn < types.N_BUTTONS; btn++ {
				for elevator := 0; elevator < types.N_ELEVATORS; elevator++ {
					RequestExists[elevator] = false
					if elevator != id && btn == types.BT_Cab {
						// Ignore inside Requests for other elevators
						continue
					}
					if elevs[elevator].RequestsQueue[floor][btn] {
						driver.SetButtonLamp(btn, floor, true)
						RequestExists[elevator] = true
					}
				}
				if RequestExists == [types.N_ELEVATORS]bool{} {
					driver.SetButtonLamp(btn, floor, false)
				}
			}
		}
	}
}