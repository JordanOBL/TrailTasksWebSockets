package server

import (
	"fmt"
	"sync"
	"time"
)

type Timer struct {
	TimerMux       sync.RWMutex  `json:"-"`
	StartTime      string        `json:"startTime"`
	StartTimestamp time.Time     `json:"-"`        // Store the actual start time
	Duration       time.Duration `json:"duration"` // Duration of the current phase (focus or break)
	IsCompleted    bool          `json:"isCompleted"`
	IsRunning      bool          `json:"isRunning"`
	IsBreak        bool          `json:"isBreak"`
	FocusTime      uint16        `json:"focusTime"`
	ShortBreakTime uint16        `json:"shortBreakTime"`
	LongBreakTime  uint16        `json:"longBreakTime"`
	Sets           uint8         `json:"sets"`
	CompletedSets  uint8         `json:"completedSets"`
	Pace           float32       `json:"pace"`
	AutoContinue   bool          `json:"autoContinue"`
	CountdownTimer *time.Timer   `json:"-"`
	UpdateTicker   *time.Ticker  `json:"-"`
	quit           chan struct{} `json:"-"`
	wg             sync.WaitGroup
}

func (t *Timer) BeginFocusTime(r *Room) {
	t.TimerMux.Lock()
	t.StartTimestamp = time.Now()
	//just starting
	t.IsRunning = true
	t.IsBreak = false
	t.Duration = time.Duration(t.FocusTime) // time.Duration(t.FocusTime) * time.Se
	t.UpdateTicker = time.NewTicker(time.Duration((0.01/t.Pace)*3600) * time.Second)
	t.quit = make(chan struct{})

	if t.StartTime == "" {
		t.StartTime = time.Now().String()

	}
	//After first param 0 , run 2nd param
	//var is a *timer to stop /cancel func from happening
	t.CountdownTimer = time.AfterFunc(time.Duration(t.FocusTime)*time.Second, func() {
		//SetBreak Resets Timer for Break && sets IsBreak bool
		r.update_protocol()
		t.SetBreak(r)
	})
	t.TimerMux.Unlock()

	t.wg.Add(1)
	ticker := t.UpdateTicker
	quit := t.quit
	go func() {
		defer t.wg.Done()
		for {
			select {
			case <-ticker.C:
				r.update_protocol() // Periodically updates session and unpaused user distance
			case <-quit:
				return
			}
		}
	}()

}

func (t *Timer) ExtraSet(r *Room) {
	//lock Timer
	t.TimerMux.Lock()
	defer t.TimerMux.Unlock()
	//increase set by 1
	t.Sets++
	//set long Break Time
	t.CountdownTimer = time.AfterFunc(time.Duration(t.LongBreakTime)*time.Second, func() {
		t.BeginFocusTime(r)
	})
}

func (t *Timer) ExtraSession(r *Room) {
	//lock Timer
	t.TimerMux.Lock()
	defer t.TimerMux.Unlock()
	//increase set by 1
	t.Sets += 3
	//set long Break Time
	t.CountdownTimer = time.AfterFunc(time.Duration(t.LongBreakTime)*time.Second, func() {
		t.BeginFocusTime(r)
	})
}

func (t *Timer) SkipBreak(r *Room) {

	t.TimerMux.Lock()
	defer t.TimerMux.Unlock()
	t.CountdownTimer = time.AfterFunc(time.Duration(t.FocusTime)*time.Second, func() {
		t.BeginFocusTime(r)
	})
}

func (t *Timer) SetBreak(r *Room) error {
	//stop updating
	t.TimerMux.Lock()
	defer t.TimerMux.Unlock()

	//set isBreak to True
	t.IsBreak = true
	//set timer.completedsets + 1
	t.CompletedSets++
	//reset time tracking
	t.StartTimestamp = time.Now()
	t.Duration = time.Duration(t.ShortBreakTime)
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
		r.update_protocol()
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
	fmt.Println("in Remaining Time, Timer.Duration is:", t.Duration)
	//seconds
	elapsed := time.Since(t.StartTimestamp)

	fmt.Println("in Remaining Time, elapsed.seconds() is:", elapsed.Seconds())

	remaining := t.Duration*time.Second - elapsed

	fmt.Println("Final Remaining Time", remaining)
	if remaining < 0 {
		remaining = 0
	}
	return time.Duration(remaining)
}

// StopTicker stops the UpdateTicker and signals the update goroutine to exit.
func (t *Timer) StopTicker() {
	t.TimerMux.Lock()
	if t.UpdateTicker != nil {
		t.UpdateTicker.Stop()
	}
	if t.quit != nil {
		close(t.quit)
		t.quit = nil
	}
	t.TimerMux.Unlock()
	t.wg.Wait()
}
