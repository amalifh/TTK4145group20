package timer

import "time"

var (
	timerEndTime time.Time
	timerActive bool
)

func timerStart(duration float64) {
	timerEndTime = time.Now().Add(time.Duration(duration*float64(time.Second)))
	timerActive = true
}

func timerStop() bool {
	timerActive = false
}

//checking if timer has expired
func timerTimedOut() bool {
	return timerActive && (time.Now().After(timerEndTime))
}