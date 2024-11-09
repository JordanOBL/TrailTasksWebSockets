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
		fmt.Printf("Processing message for room %s\n", r.Id)

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
		//Start the session
		default:
			fmt.Printf("Received unknown protocol %s in room %s\n", msg.Header.Protocol, r.Id)

		}
	}
}

// SendMessage sends the encoded message to the specified hiker.
func (r *Room) sendMessage(h *Client, packet ServerPacket) {
	// Send message to hikers msg channel
	h.MsgCh <- packet
	return

}

// Takes a &Message{protocol: "", Message: interface{}}
func (r *Room) responseFactory(protocol string, hiker *Client) error {
	switch protocol {
	case "create":
		directMessage := map[string]interface{}{
			"status":  "success",
			"message": "",
			"hikers":  r.Hikers,
		}
		packet, err := r.packMessage("create", directMessage, hiker)
		if err != nil {
			fmt.Printf("Error in responseFactory: %v", err)
		}
		r.sendMessage(hiker, packet)
		return nil

	case "join":
		// Make a snapshot of the hikers map
		r.HikersMux.RLock()
		hikersSnapshot := make(map[string]*Client, len(r.Hikers))
		for k, v := range r.Hikers {
			hikersSnapshot[k] = v
		}
		r.HikersMux.RUnlock()

		// Direct message to the joining hiker
		directMessage := map[string]interface{}{
			"type":    "direct",
			"status":  "success",
			"message": "",
			"hikers":  hikersSnapshot,
		}
		packet, err := r.packMessage("join", directMessage, hiker)
		if err != nil {
			fmt.Printf("Error in responseFactory: %v", err)
		}
		r.sendMessage(hiker, packet)

		// Broadcast to all other hikers
		broadcastMessage := map[string]interface{}{
			"type":    "broadcast",
			"message": hiker.Username + " has joined the room",
			"hikers":  hikersSnapshot,
		}
		r.broadcastExcept("join", broadcastMessage, hiker)
		return nil

	case "kicked":
		// Similar approach with snapshot for kicked case
		r.HikersMux.RLock()
		hikersSnapshot := make(map[string]*Client, len(r.Hikers))
		for k, v := range r.Hikers {
			hikersSnapshot[k] = v
		}
		r.HikersMux.RUnlock()

		broadcastMessage := map[string]interface{}{
			"type":    "broadcast",
			"message": hiker.Username + " has been kicked from the room",
			"hikers":  hikersSnapshot,
		}
		r.broadcast("kicked", broadcastMessage)

	case "ready":
		// Snapshot for ready message
		r.HikersMux.RLock()
		hikersSnapshot := make(map[string]*Client, len(r.Hikers))
		for k, v := range r.Hikers {
			hikersSnapshot[k] = v
		}
		r.HikersMux.RUnlock()

		directMessage := map[string]interface{}{
			"type":    "direct",
			"status":  "success",
			"message": "",
			"hikers":  hikersSnapshot,
		}
		packet, err := r.packMessage("ready", directMessage, hiker)
		if err != nil {
			fmt.Printf("Error in responseFactory: %v", err)
		}
		r.sendMessage(hiker, packet)

		broadcastMessage := map[string]interface{}{
			"type":    "broadcast",
			"message": "",
			"hikers":  hikersSnapshot,
		}
		r.broadcastExcept("ready", broadcastMessage, hiker)

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
		packet, err := r.packMessage("broadcast", message, hiker)
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

func (r *Room) AddHiker(h *Client) {

	r.HikersMux.Lock()
	defer r.HikersMux.Unlock()
	_, ok := r.Hikers[h.Id]
	if !ok {
		r.Hikers[h.Id] = h
		fmt.Printf("Hiker %s added to room %s. Total hikers: %d\n", h.Username, r.Id, len(r.Hikers))

	} else {
		fmt.Printf("Hiker %s already exists in room %s. Total hikers: %d\n", h.Username, r.Id, len(r.Hikers))
	}

}

func (r *Room) kickHiker(h *Client) error {
	close(h.MsgCh) //Close hikers msg channel
	r.HikersMux.Lock()
	delete(r.Hikers, h.Id) //Remove from room
	r.HikersMux.Unlock()
	err := r.responseFactory("kicked", h) //Broadcast kicked message
	if err != nil {
		return fmt.Errorf("error in kickHiker: %v\n", err)
	}
	return nil
}
