/*
This package handles the direct interaction with the elevator hardware via a TCP connection,
ensuring thread-safe access to the hardware resources using a mutex.

Key Functions:
  - Init: Establishes a TCP connection to the elevator server, initializes global variables,
    and sets up the hardware interface.
  - SetMotorDirection, SetButtonLamp, SetFloorIndicator, SetDoorOpenLamp, SetStopLamp:
    Send commands to control various elevator components such as the motor, button lamps,
    and indicators.
  - PollButtons, PollFloorSensor, PollStopButton, PollObstructionSwitch: Continuously poll
    the hardware for changes and send events through designated channels.
  - GetButton, GetFloor, GetStop, GetObstruction: Read the current state of specific hardware
    inputs and convert raw data to usable boolean or integer values.
  - read, write: Provide low-level, thread-safe I/O operations to communicate with the hardware.
  - toByte, toBool: Utility functions to convert between boolean values and byte representations.

Credits: https:github.com/TTK4145/driver-go
*/
package driver

import (
	"Driver-go/elevator/types"
	"fmt"
	"net"
	"sync"
	"time"
)

const _pollRate = 25 * time.Millisecond

var _initialized bool = false
var _numFloors = types.N_FLOORS
var _mtx sync.Mutex
var _conn net.Conn

func Init(addr string, numFloors int) {
	if _initialized {
		fmt.Println("Driver already initialized!")
		return
	}
	_numFloors = numFloors
	_mtx = sync.Mutex{}
	var err error
	_conn, err = net.Dial("tcp", addr)
	if err != nil {
		panic(err.Error())
	}
	_initialized = true
}

func SetMotorDirection(dir types.MotorDirection) {
	write([4]byte{1, byte(dir), 0, 0})
}

func SetButtonLamp(button types.ButtonType, floor int, value bool) {
	write([4]byte{2, byte(button), byte(floor), toByte(value)})
}

func SetFloorIndicator(floor int) {
	write([4]byte{3, byte(floor), 0, 0})
}

func SetDoorOpenLamp(value bool) {
	write([4]byte{4, toByte(value), 0, 0})
}

func SetStopLamp(value bool) {
	write([4]byte{5, toByte(value), 0, 0})
}

func PollButtons(receiver chan<- types.ButtonEvent) {
	prev := make([][3]bool, _numFloors)
	for {
		time.Sleep(_pollRate)
		for f := 0; f < _numFloors; f++ {
			for b := types.ButtonType(0); b < 3; b++ {
				v := GetButton(b, f)
				if v != prev[f][b] && v {
					receiver <- types.ButtonEvent{
						Floor: f,
						Btn:   types.ButtonType(b),
					}
				}
				prev[f][b] = v
			}
		}
	}
}

func PollFloorSensor(receiver chan<- int) {
	prev := -1
	for {
		time.Sleep(_pollRate)
		v := GetFloor()
		if v != prev && v != -1 {
			receiver <- v
		}
		prev = v
	}
}

func PollStopButton(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(_pollRate)
		v := GetStop()
		if v != prev {
			receiver <- v
		}
		prev = v
	}
}

func PollObstructionSwitch(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(_pollRate)
		v := GetObstruction()
		if v != prev {
			receiver <- v
		}
		prev = v
	}
}

func GetButton(button types.ButtonType, floor int) bool {
	a := read([4]byte{6, byte(button), byte(floor), 0})
	return toBool(a[1])
}

func GetFloor() int {
	a := read([4]byte{7, 0, 0, 0})
	if a[1] != 0 {
		return int(a[2])
	} else {
		return -1
	}
}

func GetStop() bool {
	a := read([4]byte{8, 0, 0, 0})
	return toBool(a[1])
}

func GetObstruction() bool {
	a := read([4]byte{9, 0, 0, 0})
	return toBool(a[1])
}

func read(in [4]byte) [4]byte {
	_mtx.Lock()
	defer _mtx.Unlock()

	_, err := _conn.Write(in[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}

	var out [4]byte
	_, err = _conn.Read(out[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}

	return out
}

func write(in [4]byte) {
	_mtx.Lock()
	defer _mtx.Unlock()

	_, err := _conn.Write(in[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}
}

func toByte(a bool) byte {
	var b byte = 0
	if a {
		b = 1
	}
	return b
}

func toBool(a byte) bool {
	var b bool = false
	if a != 0 {
		b = true
	}
	return b
}
