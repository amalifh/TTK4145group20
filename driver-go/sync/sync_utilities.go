/*
/*
Package sync provides functions for managing synchronization and acknowledgment of elevator requests in a distributed elevator system.

This package includes:
	- `copyAckList`: Copies acknowledgment data from a network message to the local registered orders, ensuring consistency across elevators.
	- `checkAllAckStatus`: Verifies whether all active elevators have a specific acknowledgment status for a given request.
	- `initializeSync`: Initializes synchronization by receiving elevator and request data or timing out if no data is received.
	- `handleOfflineState`: Manages the elevator's behavior when it detects an offline state, attempting to reinitialize and recover.
	- `handleElevatorDeath`: Handles the death event of an elevator, logging the event and marking the elevator as dead.
	- `handleElevatorUpdate`: Updates the local elevator state with new information from the network, preserving the request queue.
	- `handleRequestUpdate`: Updates the elevator's request queue and acknowledgment status based on incoming request events.
	- `processIncomingMessage`: Processes incoming network messages, updating elevator states and request acknowledgments.
	- `handleSingleModeOperations`: Manages elevator operations in single-mode, ensuring requests are handled correctly when only one elevator is active.
	- `broadcastState`: Broadcasts the elevator's current state and request information to the network.
	- `handlePeerUpdate`: Handles updates to the peer list, managing elevator online/offline status and triggering request reassignment.
	- `handleRequestReassignment`: Reassigns requests from a dead elevator to the most suitable alive elevator.
	- `handleNotAck`: Handles the NOTACK acknowledgment status, updating the registered requests accordingly.
	- `handleAck`: Handles the ACK acknowledgment status, updating the registered requests and elevator's request queue.
	- `handleCompleted`: Handles the COMPLETED acknowledgment status, updating the registered requests and elevator's request queue upon request completion.

These functions help maintain synchronization between elevators by tracking implicit acknowledgments, managing peer updates, and ensuring correct request assignments in a distributed environment.

Credits: https://github.com/perkjelsvik/TTK4145-sanntid
*/

package sync

import (
	"Driver-go/elevator/types"
	"Driver-go/network/peers"
	"Driver-go/requests"
	"fmt"
	"strconv"
	"time"
)

func copyAckList(
	msg types.NetworkMessage,
	registeredOrders [types.N_FLOORS][types.N_BUTTONS - 1]types.AckList,
	elevator, floor, id int,
	btn types.ButtonType) [types.N_FLOORS][types.N_BUTTONS - 1]types.AckList {
	registeredOrders[floor][btn].ImplicitAcks[id] = msg.RegisteredRequests[floor][btn].ImplicitAcks[elevator]
	registeredOrders[floor][btn].ImplicitAcks[elevator] = msg.RegisteredRequests[floor][btn].ImplicitAcks[elevator]
	registeredOrders[floor][btn].ChosenElevator = msg.RegisteredRequests[floor][btn].ChosenElevator
	return registeredOrders
}

func checkAllAckStatus(requestsList [types.N_ELEVATORS]bool, ImplicitAcks [types.N_ELEVATORS]types.Acknowledge, status types.Acknowledge) bool {
	for elev := 0; elev < types.N_ELEVATORS; elev++ {
		if !requestsList[elev] {
			continue
		}
		if ImplicitAcks[elev] != status {
			return false
		}
	}
	return true
}

func initializeSync(elevList *[types.N_ELEVATORS]types.ElevInfo,
	registeredRequests *[types.N_FLOORS][types.N_BUTTONS - 1]types.AckList,
	someUpdate *bool, offline *bool, ch SyncChannels, id int) {

	timeout := make(chan bool)
	go func() {
		time.Sleep(1 * time.Second)
		timeout <- true
	}()

	select {
	case initMsg := <-ch.IncomingMsg:
		*elevList = initMsg.Elevator
		*registeredRequests = initMsg.RegisteredRequests
		*someUpdate = true
	case <-timeout:
		*offline = true
	}
}

func handleOfflineState(
	offline *bool,
	aliveList [types.N_ELEVATORS]bool,
	id int,
	reInitTimer *time.Timer,
	ch SyncChannels,
	elevList *[types.N_ELEVATORS]types.ElevInfo,
	someUpdate *bool,
) {
	if *offline && aliveList[id] {
		*offline = false
		reInitTimer.Reset(1 * time.Second)

		for {
			select {
			case reInitMsg := <-ch.IncomingMsg:
				if reInitMsg.Elevator != *elevList && reInitMsg.ID != id {
					tmpElevator := elevList[id]
					*elevList = reInitMsg.Elevator
					elevList[id] = tmpElevator
					*someUpdate = true
					reInitTimer.Stop()
					return
				}
			case <-reInitTimer.C:
				return
			}
		}
	}
}

func handleElevatorDeath(
	lostID *int,
	recentlyDied *[types.N_ELEVATORS]bool,
) {
	if *lostID != -1 {
		fmt.Printf("ELEVATOR %d DIED\n", *lostID)
		recentlyDied[*lostID] = true
		*lostID = -1
	}
}

func handleElevatorUpdate(newElev types.ElevInfo, elevList *[types.N_ELEVATORS]types.ElevInfo,
	id int, peerTxEnable chan<- bool, someUpdate *bool) {

	oldQueue := elevList[id].RequestsQueue
	if newElev.State == types.EB_Undefined {
		peerTxEnable <- false
	} else if elevList[id].State == types.EB_Undefined {
		peerTxEnable <- true
	}

	elevList[id] = newElev
	elevList[id].RequestsQueue = oldQueue
	*someUpdate = true
}

func handleRequestUpdate(
	newRequest types.ButtonEvent,
	elevList *[types.N_ELEVATORS]types.ElevInfo,
	id int,
	registeredRequests *[types.N_FLOORS][types.N_BUTTONS - 1]types.AckList,
	someUpdate *bool,
) {
	if newRequest.Done {
		elevList[id].RequestsQueue[newRequest.Floor] = [types.N_BUTTONS]bool{}
		*someUpdate = true
		if newRequest.Btn != types.BT_Cab {
			if newRequest.Btn == types.BT_Up {
				registeredRequests[newRequest.Floor][types.BT_Up].ImplicitAcks[id] = types.COMPLETED
			} else if newRequest.Btn == types.BT_Down {
				registeredRequests[newRequest.Floor][types.BT_Down].ImplicitAcks[id] = types.COMPLETED
			}
			fmt.Printf("Completed Request %v at floor %d\n", newRequest.Btn, newRequest.Floor)
		}
	} else {
		if newRequest.Btn == types.BT_Cab {
			elevList[id].RequestsQueue[newRequest.Floor][newRequest.Btn] = true
			*someUpdate = true
		} else {
			registeredRequests[newRequest.Floor][newRequest.Btn].ChosenElevator = newRequest.ChosenElevator
			registeredRequests[newRequest.Floor][newRequest.Btn].ImplicitAcks[id] = types.ACK
			fmt.Printf("New Request ACK %v at floor %d\n\tDesignated to %d\n",
				newRequest.Btn, newRequest.Floor, newRequest.ChosenElevator)
		}
	}
}

func processIncomingMessage(
	msg types.NetworkMessage,
	id int,
	elevList *[types.N_ELEVATORS]types.ElevInfo,
	registeredRequests *[types.N_FLOORS][types.N_BUTTONS - 1]types.AckList,
	someUpdate *bool,
	aliveList [types.N_ELEVATORS]bool,
	updateChan chan<- [types.N_ELEVATORS]types.ElevInfo,
) {
	if msg.ID == id || !aliveList[msg.ID] || !aliveList[id] {
		return
	}

	if msg.Elevator != *elevList {
		tmpElevator := elevList[id]
		*elevList = msg.Elevator
		elevList[id] = tmpElevator
		*someUpdate = true
	}

	for floor := 0; floor < types.N_FLOORS; floor++ {
		for btn := types.BT_Up; btn < types.BT_Cab; btn++ {
			for elevator := 0; elevator < types.N_ELEVATORS; elevator++ {
				if elevator == id || !aliveList[msg.ID] {
					continue
				}

				switch msg.RegisteredRequests[floor][btn].ImplicitAcks[elevator] {
				case types.NOTACK:
					handleNotAck(registeredRequests, msg, elevator, floor, id, btn)
				case types.ACK:
					handleAck(registeredRequests, msg, elevator, floor, id, btn,
						elevList, someUpdate, aliveList)
				case types.COMPLETED:
					handleCompleted(registeredRequests, msg, elevator, floor, id, btn,
						elevList, someUpdate, aliveList)
				}
			}
		}
	}

	if *someUpdate {
		updateChan <- *elevList
		*someUpdate = false
	}
}

func handleSingleModeOperations(
	registeredRequests *[types.N_FLOORS][types.N_BUTTONS - 1]types.AckList,
	elevList *[types.N_ELEVATORS]types.ElevInfo,
	id int,
	someUpdate *bool,
	updateChan chan<- [types.N_ELEVATORS]types.ElevInfo,
) {
	for floor := 0; floor < types.N_FLOORS; floor++ {
		for btn := types.BT_Up; btn < types.BT_Cab; btn++ {
			if registeredRequests[floor][btn].ImplicitAcks[id] == types.ACK &&
				!elevList[id].RequestsQueue[floor][btn] {
				elevList[id].RequestsQueue[floor][btn] = true
				*someUpdate = true
			}

			if registeredRequests[floor][btn].ImplicitAcks[id] == types.COMPLETED {
				registeredRequests[floor][btn].ImplicitAcks[id] = types.NOTACK
			}
		}
	}

	if *someUpdate {
		updateChan <- *elevList
		*someUpdate = false
	}
}

func broadcastState(
	sendMsg *types.NetworkMessage,
	registeredRequests [types.N_FLOORS][types.N_BUTTONS - 1]types.AckList,
	elevList [types.N_ELEVATORS]types.ElevInfo,
	id int,
	offline bool,
	outgoingChan chan<- types.NetworkMessage,
) {
	if !offline {
		sendMsg.RegisteredRequests = registeredRequests
		sendMsg.Elevator = elevList
		sendMsg.ID = id
		outgoingChan <- *sendMsg
	}
}

func handlePeerUpdate(
	p peers.PeerUpdate,
	aliveList *[types.N_ELEVATORS]bool,
	recentlyDied *[types.N_ELEVATORS]bool,
	offline *bool,
	singleModeTicker **time.Ticker,
	elevList [types.N_ELEVATORS]types.ElevInfo,
	aliveChan chan<- [types.N_ELEVATORS]bool,
	timers *struct {
		reassign   *time.Timer
		broadcast  *time.Ticker
		singleMode *time.Ticker
		reInit     *time.Timer
	},
) int {
	fmt.Printf("Peer update:\n Peers: %q\n New: %q\n Lost: %q\n", p.Peers, p.New, p.Lost)
	lostID := -1

	if len(p.Peers) == 0 {
		*offline = true
		(*singleModeTicker).Stop()
	} else if len(p.Peers) == 1 {
		*singleModeTicker = time.NewTicker(100 * time.Millisecond)
	} else {
		(*singleModeTicker).Stop()
	}

	if len(p.New) > 0 {
		newID, _ := strconv.Atoi(p.New)
		aliveList[newID] = true
	} else if len(p.Lost) > 0 {
		lostID, _ = strconv.Atoi(p.Lost[0])
		aliveList[lostID] = false
		if elevList[lostID].RequestsQueue != [types.N_FLOORS][types.N_BUTTONS]bool{} && !recentlyDied[lostID] {
			timers.reassign.Reset(1 * time.Second)
		}
	}

	fmt.Println("Online elevators changed: ", *aliveList)
	go func() { aliveChan <- *aliveList }()
	return lostID
}

func handleRequestReassignment(
	elevList *[types.N_ELEVATORS]types.ElevInfo,
	recentlyDied *[types.N_ELEVATORS]bool,
	id int,
	updateChan chan<- [types.N_ELEVATORS]types.ElevInfo,
) {
	for elevator := 0; elevator < types.N_ELEVATORS; elevator++ {
		if !recentlyDied[elevator] {
			continue
		}

		originalAlive := [types.N_ELEVATORS]bool{}
		for e := 0; e < types.N_ELEVATORS; e++ {
			originalAlive[e] = !recentlyDied[e]
		}

		recentlyDied[elevator] = false

		for floor := 0; floor < types.N_FLOORS; floor++ {
			for btn := types.BT_Up; btn < types.BT_Cab; btn++ {
				if elevList[elevator].RequestsQueue[floor][btn] {
					request := types.ButtonEvent{
						Floor: floor,
						Btn:   btn,
					}

					bestElev := requests.CalcChosenElevator(
						request,
						*elevList,
						id,
						originalAlive,
					)

					elevList[bestElev].RequestsQueue[floor][btn] = true
					elevList[elevator].RequestsQueue[floor][btn] = false
				}
			}
		}
	}
	updateChan <- *elevList
}
func handleNotAck(
	registeredRequests *[types.N_FLOORS][types.N_BUTTONS - 1]types.AckList,
	msg types.NetworkMessage,
	elevator, floor, id int,
	btn types.ButtonType,
) {
	if registeredRequests[floor][btn].ImplicitAcks[id] == types.COMPLETED {
		*registeredRequests = copyAckList(msg, *registeredRequests, elevator, floor, id, btn)
	} else if registeredRequests[floor][btn].ImplicitAcks[elevator] != types.NOTACK {
		registeredRequests[floor][btn].ImplicitAcks[elevator] = types.NOTACK
	}
}

func handleAck(
	registeredRequests *[types.N_FLOORS][types.N_BUTTONS - 1]types.AckList,
	msg types.NetworkMessage,
	elevator, floor, id int,
	btn types.ButtonType,
	elevList *[types.N_ELEVATORS]types.ElevInfo,
	someUpdate *bool,
	aliveList [types.N_ELEVATORS]bool,
) {
	if registeredRequests[floor][btn].ImplicitAcks[id] == types.NOTACK {
		*registeredRequests = copyAckList(msg, *registeredRequests, elevator, floor, id, btn)
	} else if registeredRequests[floor][btn].ImplicitAcks[elevator] != types.ACK {
		registeredRequests[floor][btn].ImplicitAcks[elevator] = types.ACK
	}

	if checkAllAckStatus(aliveList, registeredRequests[floor][btn].ImplicitAcks, types.ACK) &&
		!elevList[id].RequestsQueue[floor][btn] &&
		registeredRequests[floor][btn].ChosenElevator == id {
		elevList[id].RequestsQueue[floor][btn] = true
		*someUpdate = true
	}
}

func handleCompleted(
	registeredRequests *[types.N_FLOORS][types.N_BUTTONS - 1]types.AckList,
	msg types.NetworkMessage,
	elevator, floor, id int,
	btn types.ButtonType,
	elevList *[types.N_ELEVATORS]types.ElevInfo,
	someUpdate *bool,
	aliveList [types.N_ELEVATORS]bool,
) {
	if registeredRequests[floor][btn].ImplicitAcks[id] == types.ACK {
		*registeredRequests = copyAckList(msg, *registeredRequests, elevator, floor, id, btn)
	} else if registeredRequests[floor][btn].ImplicitAcks[elevator] != types.COMPLETED {
		registeredRequests[floor][btn].ImplicitAcks[elevator] = types.COMPLETED
	}

	if checkAllAckStatus(aliveList, registeredRequests[floor][btn].ImplicitAcks, types.COMPLETED) {
		registeredRequests[floor][btn].ImplicitAcks[id] = types.NOTACK
		if registeredRequests[floor][btn].ChosenElevator == id {
			elevList[id].RequestsQueue[floor][btn] = false
			*someUpdate = true
		}
	}
}
