package elevalgo

import "time"
import "fmt"

func GetWallTime() float64{
    currentTime := time.Now()
	seconds := float64(currentTime.Unix())
	milliseconds := float64(currentTime.Nanosecond()) / 1e6 
	return seconds + (milliseconds / 1000.0) 
}

type ElevatorTimer struct{
	Timer *time.Timer
	timerActive bool
}

func NewTimer() *ElevatorTimer{
	ElevTimer := ElevatorTimer{time.NewTimer(0*time.Second),false}
	<-ElevTimer.Timer.C //To empty the alarmchannel
	return &ElevTimer
}

func (t *ElevatorTimer) Start(d time.Duration){
	// Checks if an alarm has triggered before we had the chance to start again
	if !t.Timer.Stop() && t.timerActive {
		select{
		case <-t.Timer.C:
		default:
		}
	}

	t.Timer.Reset(d)
	t.timerActive = true
}

func (t *ElevatorTimer) Stop() {
	// Checks if an alarm has triggered before we had the chance to stop it
	if !t.Timer.Stop() && t.timerActive {
		select{
		case <-t.Timer.C:
		default:
		}
		
	}
	
	t.timerActive = false
}

func (t *ElevatorTimer) TimedOut() bool{
	if t.timerActive{
		select{
		case <-t.Timer.C:
			t.timerActive = false
			fmt.Println("Timer timed out")
			return true
		default:
			return false
		}
	} else{
		return false
	}
}




