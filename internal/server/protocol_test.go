package server

import (
	"fmt"
	"testing"
	"time"

	"log"
	"net/url"

	"github.com/gorilla/websocket"
)

func TestProtocol(t *testing.T) {
	// Create server
	testServer := NewServer("", 8080)
	testServer.Start()
	defer testServer.Stop()

	// Wait briefly to allow the server to start
	time.Sleep(100 * time.Millisecond) // Adjust this as needed for your environment

	wsUrl := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/groupsession"}
	log.Printf("connecting to %s", wsUrl.String())

	// Create client1
	client1, response1, err := websocket.DefaultDialer.Dial(wsUrl.String(), nil)
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer client1.Close()
	fmt.Println("Connected to server:", response1.Status)

	// Create client2
	client2, response2, err := websocket.DefaultDialer.Dial(wsUrl.String(), nil)
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer client2.Close()
	fmt.Println("Connected to server:", response2.Status)

	t.Run("Create Protocol", func(t *testing.T) {
		// Send a test message to the server
		testMessage := map[string]map[string]interface{}{"header": {"protocol": "create", "roomId": "", "userId": "1"}, "message": {"username": "testUser1"}}
		err = client1.WriteJSON(testMessage)
		if err != nil {
			t.Fatalf("Failed to write message: %v", err)
		}

		// Wait for a response from the server (if expected)
		response := &ServerPacket{}
		err := client1.ReadJSON(response)
		if err != nil {
			t.Fatalf("Failed to read message from server: %v", err)
		}

		// Validate the response (update this to match your server’s response format)
		expectedMessage := "room created"
		expectedStatus := "success"

		if response.Response["message"] != expectedMessage {
			t.Errorf("Expected response to be %s, but got %s", expectedMessage, response.Response["response"])
		}

		if response.Response["status"] != expectedStatus {
			t.Errorf("Expected status to be %s, but got %s", expectedStatus, response.Response["status"])
		}

	})
	t.Run("Join Protocol", func(t *testing.T) {
		// Send a test message to the server
		var roomIdToJoin string
		//get valid room
		for id, room := range testServer.Rooms {
			if room != nil {
				roomIdToJoin = id
				break
			}
		}
		testMessage := map[string]map[string]interface{}{"header": {"protocol": "join", "roomId": roomIdToJoin, "userId": "2"}, "message": {"username": "testUser2"}}
		err = client2.WriteJSON(testMessage)
		if err != nil {
			t.Fatalf("Failed to write message: %v", err)
		}

		// Wait for a response from the server (if expected
		directResponse := &ServerPacket{}
		err := client2.ReadJSON(directResponse)
		if err != nil {
			t.Fatalf("Failed to read message from server: %v", err)
		}

		broadcastResponse := &ServerPacket{}
		err = client1.ReadJSON(broadcastResponse)
		if err != nil {
			t.Fatalf("Failed to read message from server: %v", err)
		}

		// Validate the response (update this to match your server’s response format)
		expectedDirectMessage := "joined room"
		expectedDirectStatus := "success"
		expectedBroadcastMessage := "testUser1 joined the room"

		if directResponse.Response["message"] != expectedDirectMessage {
			t.Errorf("Expected response to be %s, but got %s", expectedDirectMessage, directResponse.Response["response"])
		}

		if directResponse.Response["status"] != expectedDirectStatus {
			t.Errorf("Expected status to be %s, but got %s", expectedDirectStatus, directResponse.Response["status"])
		}

		if broadcastResponse.Response["message"] != expectedBroadcastMessage {
			t.Errorf("Expected response to be %s, but got %s", expectedBroadcastMessage, broadcastResponse.Response["response"])
		}

	})
}
