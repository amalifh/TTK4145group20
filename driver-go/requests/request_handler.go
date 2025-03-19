package request_control

import (
	"Driver-go/elevator/driver"
	. "Driver-go/elevator/types"
	"Driver-go/network/peers"
	"Driver-go/requests/request_assigner"
	"time"
)

func RequestHandler(
	localID string, // Identifier for this elevator
	requestsCh chan<- [N_FLOORS][N_BUTTONS]bool, // Channel for sending the list of requests to other components.
	completedRequestCh <-chan ButtonEvent, // Receives notifications when requests are completed.
	buttonEventCh <-chan ButtonEvent, // Receives notifications when buttons are pressed.
	messageTx chan<- NetworkMessage, // Channel for sending messages to other elevators.
	messageRx <-chan NetworkMessage, // Channel for receiving messages from other elevators.
	peerUpdateCh <-chan peers.PeerUpdate, // Channel for receiving updates about the network.
) {

	sendTicker := time.NewTicker(SEND_TIME_MS * time.Millisecond)
	assignRequestTicker := time.NewTicker(ASSIGN_REQUESTS_TIME_MS * time.Millisecond)

	connectedToNetwork := false
	peerList := []string{}

	hallRequests := [N_FLOORS][N_HALL_BUTTONS]Request{}  // Requests for up/down buttons on each floor.
	allCabRequests := make(map[string][N_FLOORS]Request) // Requests inside elevators (each elevator tracks its own).
	latestInfoElevators := make(map[string]ElevatorInfo) // Most recent status of all known elevators.

	// This elevator tracks its own cab requests and initial status.
	allCabRequests[localID] = [N_FLOORS]Request{}
	latestInfoElevators[localID] = ElevatorInfo{}

	for {
		select {
		/*
			When a button press is detected, it retrieves the Request for that button and floor.
			Different Logic for Cab vs Hall buttons.
				- Cab buttons (inside the elevator): Only this elevator cares.
				- Hall buttons (up/down outside elevator): Network-wide coordination.
			Request Handling Logic:
				- NEW Request: Try to assign it (if all peers know about it).
				- COMPLETE Request: Re-activate it.
				- Set lamps for assigned requests.
		*/
		case btn := <-buttonEventCh:
			request := Request{}
			if btn.Button == BT_Cab {
				request = allCabRequests[localID][btn.Floor]
			} else {
				if !connectedToNetwork {
					break
				}
				request = hallRequests[btn.Floor][btn.Button]
			}

			switch request.State {
			case COMPLETED:
				request.State = NEW
				request.AwareList = []string{localID}
				if IsSubset(peerList, request.AwareList) {
					request.State = ASSIGNED
					request.AwareList = []string{localID}
					driver.SetButtonLamp(btn.Button, btn.Floor, true)
				}
			case NEW:
				if IsSubset(peerList, request.AwareList) {
					request.State = ASSIGNED
					request.AwareList = []string{localID}
					driver.SetButtonLamp(btn.Button, btn.Floor, true)
				}
			}

			if btn.Button == BT_Cab {
				localCabRequest := allCabRequests[localID]
				localCabRequest[btn.Floor] = request
				allCabRequests[localID] = localCabRequest
			} else {
				hallRequests[btn.Floor][btn.Button] = request
			}

		// When a request is completed, mark it COMPLETED, increment a counter, and turn off the lamp.
		case btn := <-completedRequestCh:
			request := Request{}
			if btn.Button == BT_Cab {
				request = allCabRequests[localID][btn.Floor]
			} else {
				request = hallRequests[btn.Floor][btn.Button]
			}

			switch request.State {
			case ASSIGNED:
				request.State = COMPLETED
				request.AwareList = []string{localID}
				request.Count++
				driver.SetButtonLamp(btn.Button, btn.Floor, false)
			}

			if btn.Button == BT_Cab {
				localCabRequest := allCabRequests[localID]
				localCabRequest[btn.Floor] = request
				allCabRequests[localID] = localCabRequest
			} else {
				hallRequests[btn.Floor][btn.Button] = request
			}

		// Every 200ms, send this elevator’s status and requests to all peers (only if connected).
		case <-sendTicker.C:
			info := ElevatorInfo{}
			latestInfoElevators[localID] = info

			newMessage := NetworkMessage{
				SID:            localID,
				Available:      info.Available,
				Behaviour:      info.Behaviour,
				Floor:          info.Floor,
				Direction:      info.Direction,
				SHallRequests:  hallRequests,
				AllCabRequests: allCabRequests,
			}

			if connectedToNetwork {
				messageTx <- newMessage
			}

		/*
			Every second, reassing all request using RequestAssigner. It evaluates:
				- All hall requests.
				- All cab requests.
				- The latest status of all elevators.
				- The list of peers.
			The result is sent to the requestsCh channel.
		*/
		case <-assignRequestTicker.C:
			select {
			case requestsCh <- request_assigner.RequestAssigner(hallRequests, allCabRequests, latestInfoElevators, peerList, localID):
			default:
				// Avoid deadlock
			}
		/*
			This detects when a new peer has joined or left the network. The local elevator:
				- Updates peerList.
				- Tracks if it is connected to the network.
				- If we lose ourselves (in theory), we set connectedToNetwork = false
		*/
		case p := <-peerUpdateCh:
			peerList = p.Peers

			if p.New == localID {
				connectedToNetwork = true
			}

			if IsSubset([]string{localID}, p.Lost) {
				connectedToNetwork = false
			}

		// When a message from another elevator arrives:
		case message := <-messageRx:

			// Ignore messages from self
			if message.SID == localID {
				break
			}

			if !connectedToNetwork {
				// Not accepting messages until we are on the peerlist
				break
			}

			// Update the latest status of the elevator that sent the message.
			latestInfoElevators[message.SID] = ElevatorInfo{
				Available: message.Available,
				Behaviour: message.Behaviour,
				Direction: message.Direction,
				Floor:     message.Floor,
			}
			/*
				Handling incoming requests:
				The elevator updates:
					- Cab requests for the sender elevator.
					- Hall requests.
				For each request, it updates:
					- AwareList: Tracks which elevators are aware.
					- State: Requests move to ASSIGNED if all peers know about them.
					- Lamps: Set lamps for assigned
				This logic ensures:
					- All elevators know about all requests.
					- Requests can only be "assigned" when all elevators know about them.
					- Lamps are synced accross all elevators.
			*/
			for id, cabRequests := range message.AllCabRequests {

				if _, idExist := allCabRequests[id]; !idExist {
					// First informaton about this elevator
					for floor := range cabRequests {
						cabRequests[floor].AwareList = AddToAwareList(cabRequests[floor].AwareList, localID)
					}
					allCabRequests[id] = cabRequests
					continue
				}

				for floor := 0; floor < N_FLOORS; floor++ {
					if !shouldAcceptRequest(allCabRequests[id][floor], cabRequests[floor]) {
						continue
					}

					acceptedRequest := cabRequests[floor]
					acceptedRequest.AwareList = AddToAwareList(acceptedRequest.AwareList, localID)

					if acceptedRequest.State == NEW && IsSubset(peerList, acceptedRequest.AwareList) {
						acceptedRequest.State = ASSIGNED
						acceptedRequest.AwareList = []string{localID}
					}

					if id == localID && acceptedRequest.State == ASSIGNED {
						driver.SetButtonLamp(BT_Cab, floor, true)
					}

					tmpCabRequests := allCabRequests[id]
					tmpCabRequests[floor] = acceptedRequest
					allCabRequests[id] = tmpCabRequests
				}
			}

			for floor := 0; floor < N_FLOORS; floor++ {
				for btn := 0; btn < N_HALL_BUTTONS; btn++ {
					if !shouldAcceptRequest(hallRequests[floor][btn], message.SHallRequests[floor][btn]) {
						continue
					}

					acceptedRequest := message.SHallRequests[floor][btn]
					acceptedRequest.AwareList = AddToAwareList(acceptedRequest.AwareList, localID)

					switch acceptedRequest.State {
					case COMPLETED:
						driver.SetButtonLamp(ButtonType(btn), floor, false)
					case NEW:
						driver.SetButtonLamp(ButtonType(btn), floor, false)
						if IsSubset(peerList, acceptedRequest.AwareList) {
							acceptedRequest.State = ASSIGNED
							acceptedRequest.AwareList = []string{localID}
							driver.SetButtonLamp(ButtonType(btn), floor, true)
						}
					case ASSIGNED:
						driver.SetButtonLamp(ButtonType(btn), floor, true)
					}

					hallRequests[floor][btn] = acceptedRequest
				}
			}
		}
	}
}
