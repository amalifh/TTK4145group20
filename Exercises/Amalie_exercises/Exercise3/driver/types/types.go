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

//this struct to know what floor and what button is pressed
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
	clearQuestVar ClearRequestVariant
	doorOpen_s    float64
}

type Elevator struct {
	e_floor     int
	e_dirn      Dirn
	e_requests  [N_floors][N_buttons]int
	e_behaviour Elevator_behaviour
}

func elevator_uninitialized() Elevator {
	return Elevator{
		e_floor:     -1,
		e_dirn:      Dirn_Stop,
		e_behaviour: EB_idle,
	}
}
