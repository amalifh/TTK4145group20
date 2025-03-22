package types

const (
	N_buttons = 3
	N_floors  = 4
)

type Dirn int

const (
	Dirn_Down = -1
	Dirn_Stop = 0
	Dirn_UP   = 1
)

type Button_type int

const (
	B_HallUp Button_type = iota
	B_HallDown
	B_Cab
)

// this struct to know what floor and what button is pressed
type Button_event struct {
	Floor  int
	Button Button_type
}

type Elevator_behaviour int

const (
	EB_idle Elevator_behaviour = iota
	EB_DoorOpen
	EB_Moving
)

type ClearRequestVariant int

const (
	CV_All    ClearRequestVariant = iota //everyone enters the elevator
	CV_InDirn                            //only the ones going in the right direction enters
)

type Config struct {
	ClearRequestVariant ClearRequestVariant
	DoorOpen_s          float64
}

type Elevator struct {
	E_floor       int
	E_dirn        Dirn
	E_requests    [N_floors][N_buttons]bool
	E_behaviour   Elevator_behaviour
	E_obstruction bool
	E_stop        bool
	E_config      Config
}

func Elevator_uninitialized() Elevator {
	return Elevator{
		E_floor:     -1,
		E_dirn:      Dirn_Stop,
		E_behaviour: EB_idle,
		E_config: Config{
			ClearRequestVariant: CV_InDirn,
			DoorOpen_s:          3,
		},
	}
}
