/*
Package sync manages synchronization and communication between distributed elevators in a multi-elevator control system.

This package provides:
- `SyncChannels`: A struct that defines channels for inter-process communication, handling updates to request assignments, elevator states, network messages, and peer connections.
- `Synchronise`: A function that ensures consistent state synchronization across elevators by exchanging messages, handling peer updates, and managing request acknowledgments.

The synchronization mechanism ensures that elevator assignments remain consistent across the network, accounts for elevator failures, and facilitates reassignment of lost requests to maintain system reliability.

Credits: https://github.com/perkjelsvik/TTK4145-sanntid
*/
package sync

import (
	"fmt"
	"strconv"
	"time"

	"Driver-go/elevator/types"
	"Driver-go/network/peers"
)

type SyncChannels struct {
	UpdateReqAssigner chan [types.N_ELEVATORS]types.ElevInfo
	UpdateSync        chan types.ElevInfo
	RequestsUpdate    chan types.ButtonEvent
	AliveElevators    chan [types.N_ELEVATORS]bool
	IncomingMsg       chan types.NetworkMessage
	OutgoingMsg       chan types.NetworkMessage
	PeerUpdate        chan peers.PeerUpdate
	PeerTxEnable      chan bool
}

func Synchronise(ch SyncChannels, id int) {
	var (
		registeredRequests [types.N_FLOORS][types.N_BUTTONS - 1]types.AckList
		elevList           [types.N_ELEVATORS]types.ElevInfo
		sendMsg            types.NetworkMessage
		aliveList          [types.N_ELEVATORS]bool
		recentlyDied       [types.N_ELEVATORS]bool
		someUpdate         bool
		offline            bool
	)

	timeout := make(chan bool)
	go func() { time.Sleep(1 * time.Second); timeout <- true }()

	select {
	case initMsg := <-ch.IncomingMsg:
		elevList = initMsg.Elevator
		registeredRequests = initMsg.RegisteredRequests
		someUpdate = true
	case <-timeout:
		offline = true
	}

	lostID := -1
	reassignTimer := time.NewTimer(5 * time.Second)
	broadcastTicker := time.NewTicker(100 * time.Millisecond)
	singleModeTicker := time.NewTicker(100 * time.Millisecond)
	reassignTimer.Stop()
	singleModeTicker.Stop()

	for {

		if offline {
			if aliveList[id] {
				offline = false
				reInitTimer := time.NewTimer(1 * time.Second)
			REINIT:
				for {
					select {
					case reInitMsg := <-ch.IncomingMsg:
						if reInitMsg.Elevator != elevList && reInitMsg.ID != id {
							tmpElevator := elevList[id]
							elevList = reInitMsg.Elevator
							elevList[id] = tmpElevator
							someUpdate = true
							reInitTimer.Stop()
							break REINIT
						}
					case <-reInitTimer.C:
						break REINIT
					}
				}
			}
		}

		if lostID != -1 {
			fmt.Println("ELEVATOR", lostID, "DIED")
			recentlyDied[lostID] = true
			lostID = -1
		}

		select {
		case newElev := <-ch.UpdateSync:
			oldQueue := elevList[id].RequestsQueue
			if newElev.State == types.EB_Undefined {
				ch.PeerTxEnable <- false
			} else if newElev.State != types.EB_Undefined && elevList[id].State == types.EB_Undefined {
				ch.PeerTxEnable <- true
			}

			elevList[id] = newElev
			elevList[id].RequestsQueue = oldQueue
			someUpdate = true

		case newRequest := <-ch.RequestsUpdate:
			if newRequest.Done {
				elevList[id].RequestsQueue[newRequest.Floor] = [types.N_BUTTONS]bool{}
				someUpdate = true
				if newRequest.Btn != types.BT_Cab {
					registeredRequests[newRequest.Floor][types.BT_Up].ImplicitAcks[id] = types.COMPLETED
					registeredRequests[newRequest.Floor][types.BT_Down].ImplicitAcks[id] = types.COMPLETED
					fmt.Println("Completed Request", newRequest.Btn, "at floor", newRequest.Floor)
				}
			} else {
				if newRequest.Btn == types.BT_Cab {
					elevList[id].RequestsQueue[newRequest.Floor][newRequest.Btn] = true
					someUpdate = true
				} else {
					registeredRequests[newRequest.Floor][newRequest.Btn].ChosenElevator = newRequest.ChosenElevator
					registeredRequests[newRequest.Floor][newRequest.Btn].ImplicitAcks[id] = types.ACK
					fmt.Println("New Request ACK", newRequest.Btn, "at floor", newRequest.Floor)
					fmt.Println("\tdesignated to", registeredRequests[newRequest.Floor][newRequest.Btn].ChosenElevator)
				}
			}

		case msg := <-ch.IncomingMsg:
			if !shouldProcessMessage(msg, id, aliveList) {
				continue
			} else if updatedList, updated := updateElevatorState(msg, id, elevList); updated {
				elevList = updatedList
				someUpdate = true
			}
			for elevator := 0; elevator < types.N_ELEVATORS; elevator++ {
				if elevator == id || !aliveList[msg.ID] || !aliveList[id] {
					continue
				}
				updatedRequests, updatedElevList, updated := processAcksForElevator(msg, elevator, id, registeredRequests, elevList, aliveList)
				if updated {
					registeredRequests = updatedRequests
					elevList = updatedElevList
					someUpdate = true
				}
			}
			if someUpdate {
				ch.UpdateReqAssigner <- elevList
				someUpdate = false
			}

		case <-singleModeTicker.C:
			for floor := 0; floor < types.N_FLOORS; floor++ {
				for btn := types.BT_Up; btn < types.BT_Cab; btn++ {
					if registeredRequests[floor][btn].ImplicitAcks[id] == types.ACK &&
						!elevList[id].RequestsQueue[floor][btn] {
						elevList[id].RequestsQueue[floor][btn] = true
						someUpdate = true
					}
					if registeredRequests[floor][btn].ImplicitAcks[id] == types.COMPLETED {
						registeredRequests[floor][btn].ImplicitAcks[id] = types.NOTACK
					}

				}
			}
			if someUpdate {
				ch.UpdateReqAssigner <- elevList
				someUpdate = false
			}

		case <-broadcastTicker.C:
			if !offline {
				sendMsg.RegisteredRequests = registeredRequests
				sendMsg.Elevator = elevList
				sendMsg.ID = id
				ch.OutgoingMsg <- sendMsg
			}

		case p := <-ch.PeerUpdate:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)
			if len(p.Peers) == 0 {
				offline = true
				singleModeTicker.Stop()
			} else if len(p.Peers) == 1 {
				singleModeTicker = time.NewTicker(100 * time.Millisecond)
			} else {
				singleModeTicker.Stop()
			}

			if len(p.New) > 0 {
				newID, _ := strconv.Atoi(p.New)
				aliveList[newID] = true
			} else if len(p.Lost) > 0 {
				lostID, _ = strconv.Atoi(p.Lost[0])
				aliveList[lostID] = false
				if elevList[lostID].RequestsQueue != [types.N_FLOORS][types.N_BUTTONS]bool{} && !recentlyDied[lostID] {
					reassignTimer.Reset(1 * time.Second)
				}
			}
			fmt.Println("Online elevators changed: ", aliveList)
			tmpList := aliveList
			go func() { ch.AliveElevators <- tmpList }()

		case <-reassignTimer.C:
			for elevator := 0; elevator < types.N_ELEVATORS; elevator++ {
				if !recentlyDied[elevator] {
					continue
				}
				recentlyDied[elevator] = false
				for floor := 0; floor < types.N_FLOORS; floor++ {
					for btn := types.BT_Up; btn < types.BT_Cab; btn++ {
						if elevList[elevator].RequestsQueue[floor][btn] {
							elevList[id].RequestsQueue[floor][btn] = true
							elevList[elevator].RequestsQueue[floor][btn] = false
						}
					}
				}
			}
			ch.UpdateReqAssigner <- elevList
			someUpdate = false
		}
	}
}
