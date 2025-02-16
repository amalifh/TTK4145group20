package elevator

// Constants for the number of floors and buttons
const (
	N_FLOORS  = 4 // Total number of floors in the elevator
	N_BUTTONS = 3 // Number of button types (HallUp, HallDown, Cab)
)

// Enum for elevator direction
type Dirn int

const (
	D_Down Dirn = -1 // Moving down
	D_Stop Dirn = 0  // Stopped (idle state)
	D_Up   Dirn = 1  // Moving up
)

// Enum for button types (HallUp, HallDown, Cab)
type ButtonType int

const (
	B_HallUp   ButtonType = iota // HallUp button (used to request the elevator to go up from the hall)
	B_HallDown                   // HallDown button (used to request the elevator to go down from the hall)
	B_Cab                        // Cab button (inside the elevator, used to select a floor)
)

// ButtonEvent struct captures information about a button press event
type ButtonEvent struct {
	Floor  int        // The floor where the button was pressed
	Button ButtonType // The type of button that was pressed
}

// Enum for elevator behavior states
type ElevatorBehaviour int

const (
	EB_Idle     ElevatorBehaviour = iota // Elevator is idle (not moving and not opening doors)
	EB_DoorOpen                          // Doors are open
	EB_Moving                            // Elevator is moving between floors
)

// Enum for request clearing behavior
type ClearRequestVariant int

const (
	CV_All    ClearRequestVariant = iota // Everyone enters the elevator, even if going in the "wrong" direction
	CV_InDirn                            // Only passengers traveling in the current direction enter
)

// Config struct holds the elevator's configuration
type Config struct {
	DoorOpenDuration_s  float64             // Duration (in seconds) that the door remains open after a request
	ClearRequestVariant ClearRequestVariant // Defines how requests are cleared
}

// Elevator struct represents the state of the elevator
type Elevator struct {
	Floor               int                       // Current floor of the elevator
	Dirn                Dirn                      // Current direction of the elevator
	Requests            [N_FLOORS][N_BUTTONS]bool // 2D array to track which floor has which button pressed
	Behaviour           ElevatorBehaviour         // Current behavior of the elevator
	ObstructionDetected bool                      // Flag to indicate if the obstruction switch is active
	StopButtonPressed   bool                      // Flag to indicate if the stop button is pressed
	Config              Config                    // Configuration for the elevator
}

// ElevatorUninitialized creates an uninitialized elevator state
// Used to represent an elevator in an uninitialized state before starting operation
func ElevatorUninitialized() Elevator {
	return Elevator{
		Floor:     -1,      // Uninitialized floor
		Dirn:      D_Stop,  // Initial direction is stopped
		Behaviour: EB_Idle, // Initial behavior is idle
		Config: Config{
			ClearRequestVariant: CV_InDirn, // Default clearing behavior: all requests are handled
			DoorOpenDuration_s:  1.5,       // Default duration for door open: 1.5 seconds
		},
	}
}
