/*
Package requests manages elevator request assignments and button light updates in a multi-elevator system.

This package contains functions for handling incoming requests, distributing them among active elevators,
updating request queues, and ensuring proper synchronization between elevators. It also manages button
lamp states to reflect the current status of requests.

Functions:
  - RequestAssigner: Listens for button press events, assigns requests to the most suitable elevator, updates request queues,
    and communicates status changes to other elevators.
  - LightsUpdater: Monitors request queues and updates the corresponding button lamps to indicate active requests.

The package operates using multiple channels for inter-module communication, ensuring a distributed approach to request
handling and elevator coordination.

Credits: https://github.com/perkjelsvik/TTK4145-sanntid
*/
package requests

import (
	"Driver-go/elevator/driver"
	"Driver-go/elevator/types"
	"fmt"
)

func RequestAssigner(
	id int,
	bPressedCh chan types.ButtonEvent,
	lUpdateCh chan [types.N_ELEVATORS]types.ElevInfo,
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
			if !aliveList[id] && newLocalRequest.Btn == types.BT_Cab {
				elevList[id].RequestsQueue[newLocalRequest.Floor][types.BT_Cab] = true
				lUpdateCh <- elevList
				go func() { newRequestsCh <- newLocalRequest }()

			} else if !aliveList[id] && newLocalRequest.Btn != types.BT_Cab {
				continue
			} else {
				if newLocalRequest.Floor == elevList[id].Floor && elevList[id].State != types.EB_Moving {
					newRequestsCh <- newLocalRequest
				} else {
					if !duplicateRequest(newLocalRequest, elevList, id) {
						fmt.Println("New request at floor ", newLocalRequest.Floor, " for button ", newLocalRequest.Btn)
						newLocalRequest.ChosenElevator = CalcChosenElevator(newLocalRequest, elevList, id, aliveList)
						updatedRequestsCh <- newLocalRequest
					}
				}
			}

		case completedRequests.Floor = <-completedRequestsCh:
			completedRequests.Done = true
			currentFloor := completedRequests.Floor
			currentDirn := elevList[id].Dir
			clearVariant := elevList[id].CV

			for btn := types.BT_Up; btn <= types.BT_Down; btn++ {
				for elevator := 0; elevator < types.N_ELEVATORS; elevator++ {
					if elevator >= len(elevList) {
						continue
					}

					switch clearVariant {
					case types.CV_All:
						elevList[elevator].RequestsQueue[currentFloor][btn] = false

					case types.CV_InDirn:
						if elevator == id {
							if (currentDirn == types.ED_Up && btn == types.BT_Up) ||
								(currentDirn == types.ED_Down && btn == types.BT_Down) {
								elevList[elevator].RequestsQueue[currentFloor][btn] = false
							}
						}
					}
				}
			}

			if clearVariant == types.CV_All || (clearVariant == types.CV_InDirn && id == completedRequests.ChosenElevator) {
				elevList[id].RequestsQueue[currentFloor][types.BT_Cab] = false
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
	for elevs := range lUpdateCh {
		for floor := 0; floor < types.N_FLOORS; floor++ {
			// Handle hall buttons (Up and Down)
			for btn := types.BT_Up; btn <= types.BT_Down; btn++ {
				hasRequest := false
				for _, elev := range elevs {
					if elev.RequestsQueue[floor][btn] {
						hasRequest = true
						break
					}
				}
				driver.SetButtonLamp(btn, floor, hasRequest)
			}

			// Handle cab button (only for this elevator)
			cabRequest := elevs[id].RequestsQueue[floor][types.BT_Cab]
			driver.SetButtonLamp(types.BT_Cab, floor, cabRequest)
		}
	}
}
