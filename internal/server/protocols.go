package server

import (
	"encoding/json"
	"fmt"
)

// Responds to room with Header + "hiker Joined the room"
func (r *Room) join_protocol(h *Client) error {
	fmt.Println("Hiker join protocol, amount of hikers in room before adding:", len(r.Hikers))
	r.AddHiker(h)
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

	// Debug print after updating
	fmt.Printf("r.Timer after update protocol: %+v\n", r.Timer)
	fmt.Printf("r.Session after update protocol: %+v\n", r.Session)

	return nil
}

func (r *Room) pauseHiker_protocol(h *Client) error {
	if h.IsPaused {
		return fmt.Errorf("hiker is already paused")
	}

	h.IsPaused = true
	h.Strikes++
	r.SessionMux.Lock()
	r.Session.Strikes++
	r.Session.Distance -= r.Session.calculateStrikePenalty()
	r.SessionMux.Unlock()

	return nil
}
func (r *Room) resumeHiker_protocol(h *Client) error {
	r.HikersMux.RLock()
	if !h.IsPaused {
		r.HikersMux.RUnlock()
		return fmt.Errorf("hiker is not paused")
	}
	r.HikersMux.RUnlock()
	r.HikersMux.Lock()
	h.IsPaused = false
	r.HikersMux.Unlock()
	return nil
}
func (r *Room) end_protocol(h *Client) error {
	r.Timer.UpdateTicker.Stop()
	r.Timer.CountdownTimer.Stop()

	return nil
}

func (r *Room) start_protocol() {
	r.TimerMux.Lock()
	r.Timer.IsRunning = true
	r.TimerMux.Unlock()
	r.Timer.BeginFocusTime(r)

}

func (r *Room) update_protocol() error {

	r.Timer.TimerMux.RLock()
	//should not be updating during break or pause
	if r.Timer.IsPaused || r.Timer.IsBreak {
		r.Timer.TimerMux.RUnlock()

		return nil
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
	r.HikersMux.Unlock()

	return r.responseFactory("update", r.Hikers[r.Host])
}
