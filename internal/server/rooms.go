package server

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type RoomInterface interface {
	AddHiker(c *Client) error
	removeHiker(c *Client) error
	StartTimer()
	PauseTimer()
}

type Room struct {
	Id           string
	Hikers       map[string]*Client
	HikersMux    sync.RWMutex
	Session      *Session
	IncomingMsgs chan *ClientPacket
	Timer        *Timer
	Host         string
}

func (r *Room) handleRoomMessages() {
	for msg := range r.IncomingMsgs {
		if msg == nil {
			fmt.Printf("IncomingMsgs channel closed for room %s\n", r.Id)
			return // Exit the goroutine if the channel is closed
		}
		// Process incoming messages for the room here
		fmt.Printf("Processing  %s message for room %s\n", msg.Header.Protocol, r.Id)
		fmt.Printf("Msgs waiting in rooms msg channel: %v\n", len(r.IncomingMsgs))

		switch msg.Header.Protocol {
		case "create":
			//Create new room
			err := r.create_protocol(msg.Hiker)
			if err != nil {
				fmt.Printf("Error in create protocol: %v", err)
			}
		case "join":
			//join new room
			err := r.join_protocol(msg.Hiker)
			if err != nil {
				fmt.Printf("Error in join protocol: %v", err)
			}
		case "ready":
			//ready status for the hikers
			err := r.ready_protocol(msg.Hiker)
			if err != nil {
				fmt.Printf("Error in ready protocol: %v", err)
			}
		case "updateConfig":
			//get timer config from msg.message.timerConfig
			timerConfig := msg.Message["timerConfig"]

			//get session config from msg.message.sessionConfig
			sessionConfig := msg.Message["sessionConfig"]

			err := r.updateConfig_protocol(msg.Hiker, timerConfig, sessionConfig)
			if err != nil {
				fmt.Printf("Error in updateTimerConfig protocol: %v", err)
			}
			// send ready responses
			err = r.responseFactory("updateConfig", msg.Hiker)
			if err != nil {
				log.Printf("error in ready_protocol: %v", err)

			}
		case "start":
			r.start_protocol()

			err := r.responseFactory("start", msg.Hiker)
			if err != nil {
				log.Printf("error in start_protocol: %v", err)
			}
		case "pause":
			err := r.pauseHiker_protocol(msg.Hiker)
			if err != nil {
				fmt.Printf("Error in pause protocol: %v", err)
			}
			err = r.responseFactory("pause", msg.Hiker)
			if err != nil {
				log.Printf("error in pause_protocol: %v", err)
			}
		case "resume":
			err := r.resumeHiker_protocol(msg.Hiker)
			if err != nil {
				fmt.Printf("Error in resume protocol: %v", err)
			}
			err = r.responseFactory("resume", msg.Hiker)
			if err != nil {
				log.Printf("error in resume_protocol: %v", err)
			}
		case "leave":
			err := r.leave_protocol(msg.Hiker)
			if err != nil {
				fmt.Printf("Error in leave protocol: %v", err)
			}
		case "end":
			err := r.end_protocol(msg.Hiker)
			if err != nil {
				fmt.Printf("Error in end protocol: %v", err)
			}
			err = r.responseFactory("end", msg.Hiker)
			if err != nil {
				log.Printf("error in end_protocol: %v", err)
			}
		case "extraSet":

			err := r.extraSet_protocol()
			if err != nil {
				fmt.Printf("Error in extraSet protocol: %v", err)
			}

			err = r.responseFactory("extraSet", msg.Hiker)
			if err != nil {
				log.Printf("error in extraet_protocol: %v", err)
			}
		case "extraSession":

			err := r.extraSession_protocol()
			if err != nil {
				fmt.Printf("Error in extraSession protocol: %v", err)
			}

			err = r.responseFactory("extraSession", msg.Hiker)
			if err != nil {
				log.Printf("error in extraSession_protocol: %v", err)
			}
		default:
			fmt.Printf("Received unknown protocol %s in room %s\n", msg.Header.Protocol, r.Id)

		}
	}
}

// SendMessage sends the encoded message to the specified hiker.
func (r *Room) sendMessage(h *Client, packet ServerPacket) {
	// Send message to hikers msg channel
	h.MsgCh <- packet

}

// Takes a &Message{protocol: "", Message: interface{}}
func (r *Room) responseFactory(protocol string, hiker *Client) error {

	// Make a snapshot of the hikers map
	r.HikersMux.RLock()
	hikersSnapshot := make(map[string]*Client, len(r.Hikers))
	for k, v := range r.Hikers {
		hikersSnapshot[k] = v
	}
	r.HikersMux.RUnlock()

	r.Session.SessionMux.RLock()
	sessionSnapshot := r.Session
	r.Session.SessionMux.RUnlock()

	r.Timer.TimerMux.RLock()
	timerSnapshot := r.Timer
	r.Timer.TimerMux.RUnlock()
	switch protocol {
	case "create":
		directMessage := map[string]interface{}{
			"status":  "success",
			"message": "",
			"hikers":  hikersSnapshot,
		}
		packet, err := r.packMessage("create", directMessage, hiker)
		if err != nil {
			fmt.Printf("Error in responseFactory: %v", err)
		}
		r.sendMessage(hiker, packet)
		return nil

	case "join":
		// Direct message to the joining hiker
		directMessage := map[string]interface{}{
			"type":    "direct",
			"status":  "success",
			"message": "",
			"hikers":  hikersSnapshot,
			"session": sessionSnapshot,
			"timer":   timerSnapshot,
		}
		packet, err := r.packMessage("join", directMessage, hiker)
		if err != nil {
			return fmt.Errorf("Error in responseFactory: %v", err)
		}
		r.sendMessage(hiker, packet)

		// Broadcast to all other hikers
		broadcastMessage := map[string]interface{}{
			"type":    "broadcast",
			"message": hiker.Username + " has joined the room",
			"hikers":  hikersSnapshot,
		}
		return r.broadcastExcept("join", broadcastMessage, hiker)

	case "kicked":
		broadcastMessage := map[string]interface{}{
			"type":    "broadcast",
			"message": hiker.Username + " has been kicked from the room",
			"hikers":  hikersSnapshot,
		}
		r.broadcast("kicked", broadcastMessage)

	case "ready":
		directMessage := map[string]interface{}{
			"type":    "direct",
			"status":  "success",
			"message": "",
			"hikers":  hikersSnapshot,
		}
		packet, err := r.packMessage("ready", directMessage, hiker)
		if err != nil {
			return fmt.Errorf("Error in responseFactory: %v", err)
		}
		r.sendMessage(hiker, packet)

		broadcastMessage := map[string]interface{}{
			"type":    "broadcast",
			"message": "",
			"hikers":  hikersSnapshot,
		}
		r.broadcastExcept("ready", broadcastMessage, hiker)
	case "updateConfig":

		directMessage := map[string]interface{}{
			"type":          "direct",
			"status":        "success",
			"message":       "Session Updated",
			"hikers":        hikersSnapshot,
			"sessionConfig": sessionSnapshot,
			"timerConfig":   timerSnapshot,
		}
		packet, err := r.packMessage("updateConfig", directMessage, hiker)
		if err != nil {
			return fmt.Errorf("Error in responseFactory: %v", err)
		}
		r.sendMessage(hiker, packet)

		broadcastMessage := map[string]interface{}{
			"type":          "broadcast",
			"message":       "Settings Updated",
			"hikers":        hikersSnapshot,
			"sessionConfig": sessionSnapshot,
			"timerConfig":   timerSnapshot,
		}
		fmt.Printf("responding with r.timer: %v\n", r.Timer)
		return r.broadcastExcept("updateConfig", broadcastMessage, hiker)
	case "start":
		// Start broadcast message
		err := r.broadcast("start", map[string]interface{}{
			"session": sessionSnapshot,
			"timer":   timerSnapshot,
			"message": "Starting Session",
		})
		if err != nil {
			return fmt.Errorf("Error in responseFactory: %v", err)
		}

	case "update":
		//send New hikers, session and timer states to hikers

		remainingTime := r.Timer.RemainingTime()

		message := map[string]interface{}{
			"type":          "broadcast",
			"hikers":        hikersSnapshot,
			"timer":         timerSnapshot,
			"session":       sessionSnapshot,
			"remainingTime": remainingTime.Seconds(), // Send remaining time in seconds
		}
		return r.broadcast("update", message)
	case "pause":

		// pause broadcast message
		broadcastMessage := fmt.Sprintf("Hiker %s has paused", hiker.Username)
		err := r.broadcastExcept("pause", map[string]interface{}{
			"type":          "broadcast",
			"pausedHikerId": hiker.Id,
			"message":       broadcastMessage,
			"session":       sessionSnapshot,
		}, hiker)
		if err != nil {
			return fmt.Errorf("Error in responseFactory: %v", err)
		}

		directMessage := map[string]interface{}{
			"type":    "direct",
			"status":  "success",
			"message": "",
			"session": sessionSnapshot,
		}
		packet, err := r.packMessage("pause", directMessage, hiker)
		if err != nil {
			fmt.Printf("Error in responseFactory: %v", err)
		}
		r.sendMessage(hiker, packet)
		return nil

	case "resume":
		// resume message

		remainingTime := r.Timer.RemainingTime()
		message := fmt.Sprintf("Hiker %s has resumed", hiker.Username)
		err := r.broadcastExcept("resume", map[string]interface{}{
			"resumeHikerId": hiker.Id,
			"remainingTime": remainingTime.Seconds(),
			"message":       message,
		}, hiker)
		if err != nil {
			return fmt.Errorf("Error in responseFactory: %v", err)
		}

		directMessage := map[string]interface{}{
			"type":          "direct",
			"status":        "success",
			"message":       "",
			"remainingTime": remainingTime.Seconds(),
		}
		packet, err := r.packMessage("resume", directMessage, hiker)
		if err != nil {
			fmt.Printf("Error in responseFactory: %v", err)
		}
		r.sendMessage(hiker, packet)
		return nil
	case "skipBreak":
		// skip break message
		broadcastMessage := map[string]interface{}{
			"type":    "broadcast",
			"message": "Skipping Break",
		}
		return r.broadcast("skipBreak", broadcastMessage)
	case "end":
		// end message
		return r.broadcast("end", map[string]interface{}{
			"type":    "broadcast",
			"message": "Session Ended",
		})

	case "extraSet":
		broadcastMessage := map[string]interface{}{
			"type":    "broadcast",
			"message": "Added full set, More Rewards!",
		}
		return r.broadcast("extraSet", broadcastMessage)
	case "extraSession":
		broadcastMessage := map[string]interface{}{
			"type":    "broadcast",
			"message": "Added extra session, More Rewards!",
		}
		return r.broadcast("extraSession", broadcastMessage)
	case "leave":
		message := fmt.Sprintf("Hiker %s has left", hiker.Username)
		return r.broadcastExcept("leave", map[string]interface{}{
			"type":    "broadcast",
			"message": message,
			"hikers":  hikersSnapshot,
		}, hiker)
	default:
		return fmt.Errorf("unknown protocol: %s", protocol)
	}

	return nil

}
func (r *Room) broadcastExcept(protocol string, message map[string]interface{}, h *Client) error {
	r.HikersMux.RLock() // Lock for reading
	defer r.HikersMux.RUnlock()
	fmt.Printf("Total hikers in room: %d\n", len(r.Hikers))
	fmt.Printf("Attempting broadcast except %v\n", h.Username)
	for id, hiker := range r.Hikers {

		// Send message to all hikers
		// Do not send message to sender
		if id == h.Id {
			fmt.Printf("Skipping broadcast to %v\n", hiker.Username)
			continue
		}

		fmt.Printf("Broadcasting to %v\n", hiker.Username)
		packet, err := r.packMessage(protocol, message, hiker)
		if err != nil {
			return fmt.Errorf("error in broadcastExcept: %v", err)

		}
		select {
		case hiker.MsgCh <- packet:
			fmt.Printf("Broadcast Sent to %v\n", hiker.Username)
		case <-time.After(100 * time.Millisecond):
			r.warnOrRemoveHiker(hiker)
			fmt.Printf("Message dropped for %v\n", hiker.Username)
		}
	}

	return nil
}

func (r *Room) broadcast(protocol string, message map[string]interface{}) error {

	for _, hiker := range r.Hikers {
		packet, err := r.packMessage(protocol, message, hiker)
		if err != nil {
			return fmt.Errorf("Error in broadcast: %v", err)
		}
		select {
		case hiker.MsgCh <- packet:
			fmt.Printf("Broadcast Sent to %v\n", hiker.Username)
		case <-time.After(100 * time.Millisecond):
			r.warnOrRemoveHiker(hiker)
			fmt.Printf("Message dropped for %v\n", hiker.Username)
		}
	}

	return nil
}

// packMessage add header to message creating ServerPacket struct
func (r *Room) packMessage(protocol string, message map[string]interface{}, h *Client) (ServerPacket, error) {
	newPacket := ServerPacket{
		Header: Header{
			Protocol: protocol,
			RoomId:   r.Id,
			UserId:   h.Id,
		},
		Response: message,
	}

	return newPacket, nil
}

func (r *Room) warnOrRemoveHiker(hiker *Client) {
	hiker.droppedMessages++
	if hiker.droppedMessages >= 3 {
		r.kickHiker(hiker)
		log.Printf("kicked hiker %s due to inactivity or slow connection\n", hiker.Id)
	}
}

func (r *Room) AddHiker(h *Client) error {

	r.HikersMux.Lock()
	defer r.HikersMux.Unlock()
	fmt.Println("in AddHiker")
	_, ok := r.Hikers[h.Id]
	if !ok {
		fmt.Println("setting hiker in room")
		r.Hikers[h.Id] = h
		fmt.Println("making hiker room Id")
		h.RoomId = r.Id
		fmt.Printf("Hiker %s added to room %s. Total hikers: %d\n", h.Username, r.Id, len(r.Hikers))
		r.Timer.TimerMux.RLock()
		defer r.Timer.TimerMux.RUnlock()
		if r.Timer.IsRunning {
			h.IsReady = true
		}
		return nil
	} else {

		return fmt.Errorf("Hiker %s already exists in room %s. Total hikers: %d\n", h.Username, r.Id, len(r.Hikers))
	}

}

func (r *Room) RemoveHiker(h *Client) string {
	r.HikersMux.Lock()
	defer r.HikersMux.Unlock()
	delete(r.Hikers, h.Id)
	fmt.Printf("Hiker %s removed from room %s. Total hikers: %d\n", h.Username, r.Id, len(r.Hikers))
	if len(r.Hikers) == 0 {
		close(r.IncomingMsgs)
		return "close room"
	} else {
		r.setNewHost()
		return "set new host"
	}

}

func (r *Room) setNewHost() error {
	var newHost *Client
	for _, hiker := range r.Hikers {
		newHost = hiker
		hiker.IsHost = true
		r.Host = hiker.Id
		fmt.Printf("Hiker %s is the new host\n", newHost.Username)
		break
	}
	hikersSnapshot := make(map[string]*Client, len(r.Hikers))
	for k, v := range r.Hikers {
		hikersSnapshot[k] = v
	}

	fmt.Printf("New Host is: %s\n", newHost.Username)
	message := map[string]interface{}{
		"type":    "broadcast",
		"hikers":  hikersSnapshot,
		"message": fmt.Sprintf("%s is the new host", newHost.Username),
	}

	return r.broadcast("newHost", message)

}

func (r *Room) kickHiker(h *Client) error {
	close(h.MsgCh)         //Close hikers msg channel
	delete(r.Hikers, h.Id) //Remove from room

	if len(r.Hikers) == 0 {
		close(r.IncomingMsgs)
	} else {
		err := r.setNewHost()
		if err != nil {
			return fmt.Errorf("error in kickHiker: %v\n", err)
		}
	}

	err := r.responseFactory("kicked", h) //Broadcast kicked message
	if err != nil {
		return fmt.Errorf("error in kickHiker: %v\n", err)
	}
	return nil
}
