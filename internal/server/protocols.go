package server

import (
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

//func (r *Room) start_protocol(h *Client) error {
//	for _, hiker := range r.Hikers {
//		if !hiker.IsReady {
//			// Create payload as a map for more flexibility
//			payload := map[string]interface{}{
//				"status":  "error",
//				"message": "All Hikers Must Be Ready To Start"}
//
//			// Encode response message with room header for client
//			message, err := r.packMessage("error", payload, hiker)
//			if err != nil {
//				return fmt.Errorf("error encoding message in ready_protocol: %v", err)
//			}
//
//			// Send message to all hikers
//			r.broadcast(message)
//		}
//		break
//	}
//
//	// Create payload as a map for more flexibility
//	payload := map[string]interface{}{
//		"status": "success", "message": "starting session"}
//
//	// Encode response message with room header for client
//	message, err := r.packMessage("start", payload, h)
//	if err != nil {
//		return fmt.Errorf("error encoding message in ready_protocol: %v", err)
//	}
//
//	// Send message to specific hiker
//	r.broadcast(message)
//
//	return nil
//}
