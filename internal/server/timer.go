package server

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

type Timer struct {
	TimerMux       sync.RWMutex
	StartTime      string        `json:"startTime"`
	StartTimestamp time.Time     `json:"-"`        // Store the actual start time
	Duration       time.Duration `json:"duration"` // Duration of the current phase (focus or break)
	IsCompleted    bool          `json:"isCompleted"`
	Time           uint16        `json:"time"`
	IsRunning      bool          `json:"isRunning"`
	IsBreak        bool          `json:"isBreak"`
	IsPaused       bool          `json:"isPaused"`
	FocusTime      uint16        `json:"focusTime"`
	ShortBreakTime uint16        `json:"shortBreakTime"`
	LongBreakTime  uint16        `json:"longBreakTime"`
	Sets           uint8         `json:"sets"`
	CompletedSets  uint8         `json:"completedSets"`
	Pace           float32       `json:"pace"`
	AutoContinue   bool          `json:"autoContinue"`
	CountdownTimer *time.Timer   `json:"-"`
	UpdateTicker   *time.Ticker  `json:"-"`
}

func (t *Timer) BeginFocusTime(r *Room) {
	t.TimerMux.Lock()
	defer t.TimerMux.Unlock()
	//just starting
	if t.StartTime == "" {
		t.StartTime = time.Now().String()
		t.StartTimestamp = time.Now()
		fmt.Println("TImer FocusTime", int(t.FocusTime))
		t.Duration = time.Duration(t.FocusTime) * time.Second
		fmt.Println("Timer Duration", t.Duration, t.Duration)
		t.UpdateTicker = time.NewTicker(time.Duration((0.01/t.Pace)*3600) * time.Second)

	} else {
		t.UpdateTicker.Reset(time.Duration((0.01/t.Pace)*3600) * time.Second)
	}
	//After first param 0 , run 2nd param
	//var is a *timer to stop /cancel func from happening
	t.CountdownTimer = time.AfterFunc(time.Duration(t.FocusTime)*time.Second, func() {
		//SetBreak Resets Timer for Break && sets IsBreak bool
		t.SetBreak(r)
	})

	go func() {
		for {
			select {
			case <-t.UpdateTicker.C:
				r.update_protocol() // Periodically updates session and unpaused user distance
			}
		}
	}()

}

func (t *Timer) SetBreak(r *Room) error {
	//stop updating
	t.UpdateTicker.Stop()

	//set isBreak to True
	t.IsBreak = true
	//set timer.completedsets + 1
	t.CompletedSets++
	//reset time tracking
	t.StartTimestamp = time.Now()
	t.Duration = time.Duration(t.ShortBreakTime) * time.Second
	//get snapshot of all hikers data
	r.HikersMux.RLock()
	hikersSnapshot := make(map[string]*Client, len(r.Hikers))
	for k, v := range r.Hikers {
		hikersSnapshot[k] = v
	}
	r.HikersMux.RUnlock()

	//check which break should be set
	//if completed all sets send message to show modal
	if t.CompletedSets == t.Sets {
		if t.AutoContinue == true {
			//set long Break Time
			t.CountdownTimer = time.AfterFunc(time.Duration(t.LongBreakTime)*time.Second, func() {
				t.BeginFocusTime(r)
			})
			return nil
		}

		//Broadcast Completed Message to all hikers
		//EndModal protocol tells UI to display modal to host to end or continue
		//UI will send endSession protocol, if continue session protocol isnt sent within x seconds
		r.broadcast("endModal", map[string]interface{}{
			"type":    "broadcast",
			"session": r.Session,
			"timer":   r.Timer,
			"hikers":  hikersSnapshot,
			"message": "Congrats, You Finished!",
		})
		return nil
	}
	//Timers Break Timer begins, will call BeginFocusTime once breakTime is Reached
	t.CountdownTimer = time.AfterFunc(time.Duration(t.ShortBreakTime)*time.Second, func() {
		t.BeginFocusTime(r)
	})

	//broadcast to users its break time
	message := map[string]interface{}{
		"type":    "broadcast",
		"session": r.Session,
		"timer":   r.Timer,
		"hikers":  hikersSnapshot,
	}

	//Broadcast "shortBreak" protocol tells ui to switch to break mode
	r.broadcast("shortBreak", message)
	return nil
}

func (t *Timer) RemainingTime() time.Duration {
	t.TimerMux.RLock()
	defer t.TimerMux.RUnlock()

	intString := strconv.Itoa(int(t.Duration))
	remaining, _ := time.ParseDuration(intString + "s")

	elapsed := time.Since(t.StartTimestamp)
	remaining = remaining - elapsed
	if remaining < 0 {
		remaining = 0
	}
	return remaining
}
