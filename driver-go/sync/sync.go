package sync

import (
	"fmt"
	"strconv"
	"time"

	"Driver-go/elevator/types"
	"Driver-go/network/peers"
)

// SyncChannels contains all channels between governor - sync and sync - network
type SyncChannels struct {
	UpdateReqAssigner  chan [types.N_ELEVATORS]types.ElevInfo // Send updated elevator information to the request_assigner.  
	UpdateSync      chan types.ElevInfo // Receive updates about the state of a specific elevator.
	RequestsUpdate     chan types.ButtonEvent // Send updates about button requests made by users.
	AliveElevators chan [types.N_ELEVATORS]bool // Send information about which elevators are alive.
	IncomingMsg     chan types.NetworkMessage // Receive messages from the network.
	OutgoingMsg     chan types.NetworkMessage // Send messages to the network.
	PeerUpdate      chan peers.PeerUpdate // Handle updates about the state of peers (other elevators).
	PeerTxEnable    chan bool // Enable or disable peer communication 
}

// Synchronise called as goroutine; forwards data to network, synchronises data from network.
func Synchronise(ch SyncChannels, id int) {
	var (
		registeredRequests [types.N_FLOORS][types.N_BUTTONS - 1]types.AckList
		elevList         [types.N_ELEVATORS]types.ElevInfo
		sendMsg          types.NetworkMessage
		aliveList       [types.N_ELEVATORS]bool
		recentlyDied     [types.N_ELEVATORS]bool
		someUpdate       bool
		offline          bool
	)

	timeout := make(chan bool)
	go func() { time.Sleep(1 * time.Second); timeout <- true }() // If no message is received within 1 second, the system enters an offline state.

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
					fmt.Println("We Finished Request", newRequest.Btn, "at floor", newRequest.Floor+1)
				}
			} else {
				if newRequest.Btn == types.BT_Cab {
					elevList[id].RequestsQueue[newRequest.Floor][newRequest.Btn] = true
					someUpdate = true
				} else {
					registeredRequests[newRequest.Floor][newRequest.Btn].ChosenElevator = newRequest.ChosenElevator
					registeredRequests[newRequest.Floor][newRequest.Btn].ImplicitAcks[id] = types.ACK
					fmt.Println("We acknowledged a new Request", newRequest.Btn, "at floor", newRequest.Floor+1)
					fmt.Println("\tdesignated to", registeredRequests[newRequest.Floor][newRequest.Btn].ChosenElevator)
				}
			}

		case msg := <-ch.IncomingMsg:
			if msg.ID == id || !aliveList[msg.ID] || !aliveList[id] {
				continue
			} else {
				if msg.Elevator != elevList {
					tmpElevator := elevList[id]
					elevList = msg.Elevator
					elevList[id] = tmpElevator
					someUpdate = true
				}
				for elevator := 0; elevator < types.N_ELEVATORS; elevator++ {
					if elevator == id || !aliveList[msg.ID] || !aliveList[id] {
						continue
					}
					for floor := 0; floor < types.N_FLOORS; floor++ {
						for btn := types.BT_Up; btn < types.BT_Cab; btn++ {
							switch msg.RegisteredRequests[floor][btn].ImplicitAcks[elevator] {
							case types.NOTACK:
								if registeredRequests[floor][btn].ImplicitAcks[id] == types.COMPLETED {
									registeredRequests = copyAckList(msg, registeredRequests, elevator, floor, id, btn)
								} else if registeredRequests[floor][btn].ImplicitAcks[elevator] != types.NOTACK {
									registeredRequests[floor][btn].ImplicitAcks[elevator] = types.NOTACK
								}

							case types.ACK:
								if registeredRequests[floor][btn].ImplicitAcks[id] == types.NOTACK {
									fmt.Println("Request ", btn, "from ", msg.ID, "in floor", floor+1, "has been acked!")
									registeredRequests = copyAckList(msg, registeredRequests, elevator, floor, id, btn)
								} else if registeredRequests[floor][btn].ImplicitAcks[elevator] != types.ACK {
									registeredRequests[floor][btn].ImplicitAcks[elevator] = types.ACK
								}
								if checkAllAckStatus(aliveList, registeredRequests[floor][btn].ImplicitAcks, types.ACK) &&
									!elevList[id].RequestsQueue[floor][btn] &&
									registeredRequests[floor][btn].ChosenElevator == id {
									fmt.Println("We've been assigned a new request!")
									elevList[id].RequestsQueue[floor][btn] = true
									someUpdate = true
								}

							case types.COMPLETED:
								if registeredRequests[floor][btn].ImplicitAcks[id] == types.ACK {
									registeredRequests = copyAckList(msg, registeredRequests, elevator, floor, id, btn)
								} else if registeredRequests[floor][btn].ImplicitAcks[elevator] != types.COMPLETED {
									registeredRequests[floor][btn].ImplicitAcks[elevator] = types.COMPLETED
								}

								if checkAllAckStatus(aliveList, registeredRequests[floor][btn].ImplicitAcks, types.COMPLETED) {
									registeredRequests[floor][btn].ImplicitAcks[id] = types.NOTACK
									if registeredRequests[floor][btn].ChosenElevator == id {
										elevList[id].RequestsQueue[floor][btn] = false
										someUpdate = true
									}
								}
							}
						}
					}
				}
				if someUpdate {
					ch.UpdateReqAssigner <- elevList
					someUpdate = false
				}
			}

		// This ticker is used to periodically check for new requests in single elevator mode.
		// It checks if any button requests are still pending acknowledgment and assigns them to the appropriate elevator.
		case <-singleModeTicker.C:
			for floor := 0; floor < types.N_FLOORS; floor++ {
				for btn := types.BT_Up; btn < types.BT_Cab; btn++ {
					if registeredRequests[floor][btn].ImplicitAcks[id] == types.ACK &&
						!elevList[id].RequestsQueue[floor][btn] {
						fmt.Println("We've been assigned a new request!")
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

		// This ticker is used to periodically broadcast the updated elevator and request status to the network.
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