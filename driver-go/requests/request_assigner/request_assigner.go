/*
	This package is basically a bridge between the code and the external hall_request_assigner program,
	which acts as the brain for assigning requests in a multi-elevator system.
		- The code collects the local state and peer states.
		- It packages that into JSON.
		- It sends that to hall_request_assigner, which calculates the optimal assignments.
		- It takes the output and returns only the assignments for this elevator.
*/

/*
	hall_request_assigner Pseudo Code (Python):
		import json
		import sys

		def load_input():
			import argparse
			parser = argparse.ArgumentParser()
			parser.add_argument("-i", required=True)
			args = parser.parse_args()
			data = json.loads(args.i)
			return data

		def assign_requests(input_data):
			assignments = {elevator: [[False, False] for _ in range(len(input_data["hallRequests"]))] for elevator in input_data["states"]}

			for floor, (up, down) in enumerate(input_data["hallRequests"]):
				if up:
					best_elevator = find_best_elevator(input_data, floor, "up")
					if best_elevator:
						assignments[best_elevator][floor][0] = True

				if down:
					best_elevator = find_best_elevator(input_data, floor, "down")
					if best_elevator:
						assignments[best_elevator][floor][1] = True

			return assignments

		def find_best_elevator(input_data, floor, direction):
			best_elevator = None
			best_score = float('inf')

			for elevator, state in input_data["states"].items():
				if state["behaviour"] == "idle":
					score = abs(state["floor"] - floor)
				elif state["direction"] == direction:
					if (direction == "up" and state["floor"] <= floor) or (direction == "down" and state["floor"] >= floor):
						score = abs(state["floor"] - floor)
					else:
						score = abs(state["floor"] - floor) + 10  # Penalize turning around
				else:
					score = abs(state["floor"] - floor) + 20  # Penalize opposite direction

				if score < best_score:
					best_score = score
					best_elevator = elevator

			return best_elevator

		def main():
			input_data = load_input()
			assignments = assign_requests(input_data)
			print(json.dumps(assignments))

		if __name__ == "__main__":
			main()
*/

package request_assigner

import (
	"Driver-go/elevator"
	"encoding/json"
	"fmt"
	"os/exec"
)

type ElevState struct {
	Behavior    string        
	Floor       int            
	Direction   string         
	CabRequests [N_FLOORS]bool 
}

type Input struct {
	HallRequests [N_FLOORS][2]bool       
	States       map[string]ElevState
}

/*
	- Collect information about the current state of all elevators (yours and your peers').
	- Convert the data into a format that an external executable (called hall_request_assigner) can understand.
	- Run the hall_request_assigner program, which calculates which elevator should handle which requests.
	- Take the output from hall_request_assigner and return the assigned requests for the local elevator (the one running this code).
*/

func RequestAssigner(
	hallRequests [N_FLOORS][N_HALL_BUTTONS]Request_t, // 2D array representing hall button requests (each floor has 2 buttons: UP and DOWN).
	allCabRequests map[string][N_FLOORS]Request_t, // Map of cab requests for each elevator. Each elevator has its own requests (like floor buttons inside the elevator).
	latestInfoElevators map[string]ElevatorInfo_t, // Latest known state (floor, behavior, direction) for each elevator.
	peerList []string, // List of known peers (other elevators in the system).
	localID string, // The ID of this elevator (the one running the code).
) [N_FLOORS][N_BUTTONS]bool {

	hraExecutablePath := "hall_request_assigner"
	/*
		This part converts the hallRequests array (Request_t) into a simpler [N_FLOORS][N_HALL_BUTTONS]bool array:
			- It only marks true for requests that are already ASSIGNED.
			- This simplifies data for the external program.
	*/
	boolHallRequests := [N_FLOORS][N_HALL_BUTTONS]bool{}
	for floor := 0; floor < N_FLOORS; floor++ {
		for button := 0; button < N_HALL_BUTTONS; button++ {
			if hallRequests[floor][button].State == ASSIGNED {
				boolHallRequests[floor][button] = true
			}
		}
	}

	/*
		This part gathers the current state of all elevators into a format that the hall_request_assigner expects:
			- inputStates is a map, where the key is the elevator ID, and the value is an HRAElevState struct.
			- It loops over all elevators (both local and peers).
			- It skips elevators that:
				- Are not known in latestInfoElevators.
				- Are marked as unavailable.
				- Are not in the peerList (unless it's the local elevator itself).
	*/
	inputStates := map[string]HRAElevState{}

	for id, cabRequests := range allCabRequests {
		elevatorInfo, exists := latestInfoElevators[id]
		if !exists {
			continue
		}

		if !elevatorInfo.Available {
			continue
		}

		if !sliceContains(peerList, id) && id != localID {
			continue
		}

		boolCabRequests := [N_FLOORS]bool{}
		for floor := 0; floor < N_FLOORS; floor++ {
			if cabRequests[floor].State == ASSIGNED {
				boolCabRequests[floor] = true
			}
		}
		// This is the core of how the elevator state is packaged for the external process.
		inputStates[id] = HRAElevState{
			Behavior:    behaviourToString(elevatorInfo.Behaviour),
			Floor:       elevatorInfo.Floor,
			Direction:   directionToString(elevatorInfo.Direction),
			CabRequests: boolCabRequests,
		}

	}

	if len(inputStates) == 0 {
		return [N_FLOORS][N_BUTTONS]bool{}
	}

	// If there are no valid elevators, it returns an empty request set.
	input := HRAInput{ // The HRAInput struct holds the hall requests and all elevator states.
		HallRequests: boolHallRequests,
		States:       inputStates,
	}

	// This is converted into JSON with:
	jsonBytes, err := json.Marshal(input)
	if err != nil {
		fmt.Println("json.Marshal error: ", err)
		return [N_FLOORS][N_BUTTONS]bool{}
	}
	/*
		This is where the real logic happens — the decision-making is offloaded to an external process:
			- It calls hall_request_assigner, passing the elevator state data as a JSON string.
			- --includeCab probably tells it to consider cab requests too, not just hall requests.
	*/
	ret, err := exec.Command(hraExecutablePath, "-i", string(jsonBytes), "--includeCab").CombinedOutput()
	if err != nil {
		fmt.Println("exec.Command error: ", err)
		fmt.Println(string(ret))
		return [N_FLOORS][N_BUTTONS]bool{}
	}

	/*
		Parse the output from hall_request_assigner:
			- This maps each elevator ID to a 2D array of floor-button assignments.
			- This is unmarshaled into a map[string][N_FLOORS][N_BUTTONS]bool.
	*/
	output := new(map[string][N_FLOORS][N_BUTTONS]bool)
	err = json.Unmarshal(ret, &output)
	if err != nil {
		fmt.Println("json.Unmarshal error: ", err)
		return [N_FLOORS][N_BUTTONS]bool{}
	}

	// This extracts only this elevator's assignments from the global result.
	return (*output)[localID]
}

func behaviourToString(b Behaviour_t) string {
	switch b {
	case IDLE:
		return "idle"
	case MOVING:
		return "moving"
	case DOOR_OPEN:
		return "doorOpen"
	}
	return "idle"
}

func directionToString(d Direction_t) string {
	switch d {
	case DIR_DOWN:
		return "down"
	case DIR_UP:
		return "up"
	case DIR_STOP:
		return "stop"
	}
	return "stop"
}

func sliceContains(slice []string, elem string) bool {
	for _, element := range slice {
		if element == elem {
			return true
		}
	}
	return false
}
