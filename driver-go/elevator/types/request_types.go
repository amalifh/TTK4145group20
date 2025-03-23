package types

type RequestState int

const (
	COMPLETED RequestState = iota
	NEW
	ASSIGNED
)

const (
	PEER_PORT               = 30052 // Peer discovery port
	MSG_PORT                = 30051 // Broadcast message port
	SEND_TIME_MS            = 200   // Send elevator status every 200ms
	ASSIGN_REQUESTS_TIME_MS = 1000  // RE-assign request every 1000ms
)

type Request struct {
	State     RequestState
	Count     int
	AwareList []string
}

type ElevatorInfo struct {
	Available bool
	Behaviour ElevBehaviour
	Direction ElevDirection
	Floor     int
}

type NetworkMessage struct {
	SID            string // Sender ID
	Available      bool
	Behaviour      ElevBehaviour
	Direction      ElevDirection
	Floor          int
	SHallRequests  [N_FLOORS][N_HALL_BUTTONS]Request
	AllCabRequests map[string][N_FLOORS]Request
}
