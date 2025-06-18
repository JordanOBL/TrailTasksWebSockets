package server

import (
	"encoding/json"
	"fmt"
	"math"
	"time"
)

// Responds to room with Header + "hiker Joined the room"
func (r *Room) join_protocol(h *Client) error {
	fmt.Println("Hiker join protocol, amount of hikers in room before adding:", len(r.Hikers))
	err := r.AddHiker(h)
	if err != nil {
		return fmt.Errorf("error in join_protocol: %v", err)
	}
	fmt.Println("After AddHiker, amount of hikers in room:", len(r.Hikers))

	// Broadcast the updated list
	return r.responseFactory("join", h)
}

// Responds to CLient witha Header + "welcome"
// Joins new user to room
func (r *Room) create_protocol(h *Client) error {
	// Add hiker to room
	r.AddHiker(h)

	return r.responseFactory("create", h)

}

// Protocol_2 toggles the hiker's ready state and sends a response message with the new state.
func (r *Room) ready_protocol(h *Client) error {
	h.IsReady = !h.IsReady

	// send ready responses
	err := r.responseFactory("ready", h)
	if err != nil {
		return fmt.Errorf("error in ready_protocol: %v", err)
	}
	return nil
}
func (r *Room) updateConfig_protocol(cl *Client, timerConfig interface{}, sessionConfig interface{}) error {
	r.Session.SessionMux.Lock()
	defer r.Session.SessionMux.Unlock()

	r.Timer.TimerMux.Lock()
	defer r.Timer.TimerMux.Unlock()
	// Debug print before updating
	fmt.Printf("r.Timer before update protocol: %+v\n", r.Timer)
	fmt.Printf("r.Session before update protocol: %+v\n", r.Session)

	// Marshal the sessionConfig into JSON
	sessionJsonResult, err := json.Marshal(sessionConfig)
	if err != nil {
		return fmt.Errorf("Error marshaling updated session config: %v", err)
	}
	// Unmarshal the JSON result into r.Session
	err = json.Unmarshal(sessionJsonResult, &r.Session)
	if err != nil {
		return fmt.Errorf("Error unmarshaling updated session config: %v", err)
	}

	// Marshal the timerConfig into JSON
	timerJsonResult, err := json.Marshal(timerConfig)
	if err != nil {
		return fmt.Errorf("Error marshaling updated timer config: %v", err)
	}
	// Unmarshal the JSON result into r.Timer
	err = json.Unmarshal(timerJsonResult, &r.Timer)
	if err != nil {
		return fmt.Errorf("Error unmarshaling updated timer config: %v", err)
	}
	r.Timer.Duration = time.Duration(r.Timer.FocusTime) * time.Second // time.Duration(r.Timer.FocusTime + "s") * time.Second
	// Debug print after updating
	fmt.Printf("r.Timer after update protocol: %+v\n", r.Timer)
	fmt.Printf("r.Session after update protocol: %+v\n", r.Session)

	return nil
}

func (r *Room) pauseHiker_protocol(h *Client) error {
	fmt.Println("Hiker pause protocol")
	r.Session.SessionMux.Lock()
	defer r.Session.SessionMux.Unlock()
	h.mux.Lock()
	defer h.mux.Unlock()
	if h.IsPaused {
		return fmt.Errorf("hiker is already paused")
	}
	h.IsPaused = true

	h.Strikes += 1
	if h.Distance > 0.01 {
		h.Distance -= 0.01
	}

	// Calculate session penalty
	r.Session.Strikes += 1
	fmt.Printf("r.Session.Distance: %f\n", r.Session.Distance)
	if r.Session.Distance > 0.00 {
		sessionPenalty := r.Session.calculateStrikePenalty()
		fmt.Printf("sessionPenalty: %f\n", sessionPenalty)
		if (r.Session.Distance - sessionPenalty) < 0.00 {
			r.Session.Distance = 0.00
		} else {
			r.Session.Distance -= sessionPenalty
		}
	}
	levelDistanceFactor := 0.5 // Distance required per level increment
	r.Session.Level = uint8(math.Floor(r.Session.Distance/levelDistanceFactor) + 1)

	return nil
}
func (r *Room) resumeHiker_protocol(h *Client) error {
	if !h.IsPaused {
		return fmt.Errorf("hiker is not paused")
	}
	h.IsPaused = false
	return nil
}
func (r *Room) end_protocol(h *Client) error {
	r.Timer.TimerMux.Lock()
	defer r.Timer.TimerMux.Unlock()
	r.Session.SessionMux.Lock()
	defer r.Session.SessionMux.Unlock()
	r.HikersMux.Lock()
	defer r.HikersMux.Unlock()
	r.Timer.StopTicker()
	r.Timer.CountdownTimer.Stop()
	r.Timer.IsRunning = false
	r.Timer.IsBreak = false
	r.Timer.CompletedSets = 0
	r.Timer.Pace = 2.0
	for _, hiker := range r.Hikers {
		hiker.IsReady = false
		hiker.IsPaused = false
		hiker.Strikes = 0
		hiker.droppedMessages = 0
		hiker.Distance = 0.00
	}
	r.Session.Distance = 0.00
	r.Session.Level = 1
	r.Session.Strikes = 0
	r.Session.BonusTokens = 0
	r.Session.TokensEarned = 0
	r.Session.HighestCompletedLevel = 0

	return nil
}

func (r *Room) start_protocol() {
	r.Timer.TimerMux.Lock()
	r.Timer.IsRunning = true
	r.Timer.TimerMux.Unlock()
	r.Timer.BeginFocusTime(r)

}
func (r *Room) leave_protocol(h *Client) error {
	r.RemoveHiker(h)

	return nil
}

func (r *Room) extraSet_protocol() error {
	r.Timer.ExtraSet(r)
	return nil
}
func (r *Room) extraSession_protocol() error {
	r.Timer.ExtraSession(r)
	return nil
}

func (r *Room) skipBreak_protocol() error {
	r.Timer.SkipBreak(r)
	return nil
}

func (r *Room) update_protocol() error {

	r.Timer.TimerMux.RLock()
	//should not be updating during break or pause
	if r.Timer.IsBreak {
		r.Timer.TimerMux.RUnlock()
		return r.responseFactory("update", r.Hikers[r.Host])

	}
	r.Timer.TimerMux.RUnlock()

	//update all hikers distance +.01 if not paused
	r.HikersMux.Lock()
	for _, hiker := range r.Hikers {
		if !hiker.IsPaused {
			//increase hiker Distance
			hiker.Distance += 0.01
			//increase session total Distance
			r.Session.Distance += 0.01
		}
	}
	levelDistanceFactor := 0.5 // Distance required per level increment
	r.Session.Level = uint8(math.Floor(r.Session.Distance/levelDistanceFactor) + 1)

	if r.Session.Level > r.Session.HighestCompletedLevel {
		r.Session.HighestCompletedLevel = r.Session.Level
	}
	r.HikersMux.Unlock()

	return r.responseFactory("update", r.Hikers[r.Host])
}
