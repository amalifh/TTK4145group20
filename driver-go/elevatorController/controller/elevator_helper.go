package controller

import (
	. "Driver-go/elevator/types"
	localCtrl "Driver-go/elevatorController/controller/localController"
	"Driver-go/elevatorController/timer"
	"time"
)

// StartTimerChannel starts a custom timer for the given duration (in seconds)
// and returns a channel that signals when the timer expires.
func StartTimerChannel(t *timer.Timer, duration int) <-chan bool {
	ch := make(chan bool, 1)
	t.Start(float64(duration))
	go func() {
		for {
			if t.TimedOut() {
				ch <- true
				return
			}
			time.Sleep(50 * time.Millisecond)
		}
	}()
	return ch
}

// TimerStop stops the given custom timer.
func TimerStop(t *timer.Timer) {
	t.Stop()
}

// IsDoorOpen returns true if the elevator's state is DoorOpen.
func IsDoorOpen() bool {
	return localCtrl.CurrentElevator.Behaviour == EB_DoorOpen
}

// IsMoving returns true if the elevator's state is Moving.
func IsMoving() bool {
	return localCtrl.CurrentElevator.Behaviour == EB_Moving
}

// IsIdle returns true if the elevator's state is Idle.
func IsIdle() bool {
	return localCtrl.CurrentElevator.Behaviour == EB_Idle
}

// GetDirection returns the current elevator direction.
func GetDirection() ElevDirection {
	return localCtrl.CurrentElevator.Direction
}
