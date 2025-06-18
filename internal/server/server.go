package server

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	ws "github.com/gorilla/websocket"
)

type ServerInterface interface {
	Start() error
	Stop() error
	CreateRoom() *Room
	DeleteRoom(id string)
}

type Server struct {
	mux     sync.RWMutex
	Addr    string
	Rooms   map[string]*Room
	Clients map[*ws.Conn]bool
}

type Header struct {
	Protocol string `json:"protocol"`
	RoomId   string `json:"roomId"`
	UserId   string `json:"userId"`
}

type ServerPacket struct {
	Header   Header                 `json:"header"`
	Response map[string]interface{} `json:"response"`
}

func NewServer(host string, port int) *Server {
	return &Server{
		Addr:    host + ":" + fmt.Sprint(port),
		Rooms:   make(map[string]*Room),
		Clients: make(map[*ws.Conn]bool),
	}
}

func (s *Server) createWSConn(w http.ResponseWriter, r *http.Request) (*ws.Conn, error) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			//Allow all connectios for now
			return true
		},
	}

	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, fmt.Errorf("createWSConn error: %v", err)

	}

	return wsConn, nil
}

func (s *Server) addClient(c *ws.Conn) {

	s.mux.Lock()
	defer s.mux.Unlock()
	s.Clients[c] = true
}

func (s *Server) removeClient(c *Client) {
	//close(c.MsgCh)
	close(c.MsgCh)
	fmt.Println("Removing from room")
	if c.RoomId != "" {

		str := s.Rooms[c.RoomId].RemoveHiker(c)
		if str == "close room" {
			delete(s.Rooms, c.RoomId)
			fmt.Println("Room closed")
		}

	}

	c.Conn.Close()
	s.mux.Lock()
	defer s.mux.Unlock()
	delete(s.Clients, c.Conn)
	fmt.Println("Client Disconnected")
}

func (s *Server) handleNewConnection(w http.ResponseWriter, r *http.Request) {
	// Establish WebSocket connection
	wsConn, err := s.createWSConn(w, r)
	if err != nil {
		errString := fmt.Sprintf(`{"protocol": "Error", "error": %v}`, err)
		http.Error(w, errString, http.StatusInternalServerError)
		return
	}

	fmt.Println("New Connection From:", r.RemoteAddr)

	// Create and add new client
	client := &Client{
		Conn:  wsConn,
		MsgCh: make(chan ServerPacket, 2048), // Buffered channel for outgoing messages
	}
	s.addClient(wsConn)

	// Start client write pump
	go client.writePump()

	// Start reading messages from client
	go s.readLoop(client)
}

func (s *Server) readLoop(c *Client) {
	for {
		//read the json message *Message
		clientPacket := &ClientPacket{
			Header:  Header{},
			Message: make(map[string]interface{}), // Initialize to prevent nil map errors
			Hiker:   c,                            // Make sure `c` is a valid reference to the client
		}

		if err := c.Conn.ReadJSON(clientPacket); err != nil {
			fmt.Println("error reading client Header:", err)
			s.removeClient(c)
			return
		}

		//check incoming client message protocol
		switch clientPacket.Header.Protocol {
		case "create":
			//create a new room
			newRoom := &Room{
				Id:      uuid.New().String(),
				Hikers:  make(map[string]*Client, 1024),
				Session: &Session{Level: 1, HighestCompletedLevel: 0},
				Timer: &Timer{
					FocusTime:      1500,
					ShortBreakTime: 300,
					LongBreakTime:  900,
					Sets:           3,
					CompletedSets:  0,
					Pace:           2.0,
				},
				IncomingMsgs: make(chan *ClientPacket, 2048),
				Host:         c.Id,
			}
			//add room to Servers rooms
			s.mux.Lock()
			s.Rooms[newRoom.Id] = newRoom
			s.mux.Unlock()

			//start new thread to handle new rooms messages
			go newRoom.handleRoomMessages()

			//add username to client
			c.Username = clientPacket.Message["username"].(string)

			//add id to client
			c.Id = clientPacket.Header.UserId

			//make client host of the new room
			newRoom.Host = c.Id
			c.IsHost = true

			//add room id to packet
			clientPacket.Header.RoomId = newRoom.Id

			//send message to room Channel
			newRoom.IncomingMsgs <- clientPacket

		case "join":
			// Set username for the client
			c.Username = clientPacket.Message["username"].(string)

			// Set id for the client
			c.Id = clientPacket.Header.UserId

			// Retrieve the room by RoomId
			roomRef, ok := s.Rooms[clientPacket.Header.RoomId]
			if !ok {
				// Room does not exist, send error message to client
				errMsg := map[string]interface{}{"message": "Room ID Does Not Exist"}

				newPacket := ServerPacket{
					Header: Header{
						Protocol: "Error",
						RoomId:   "",
					},
					Response: errMsg,
				}

				select {
				case c.MsgCh <- newPacket:
				default:
					s.removeClient(c)
				}
			} else {
				// Room exists, log room size and add client to room
				fmt.Println("Amount of hikers in room:", len(roomRef.Hikers))
				roomRef.IncomingMsgs <- clientPacket
			}
		default:
			//check if room exiists on server
			//if room doenst exist respond with error
			//close connection
			roomRef, ok := s.Rooms[clientPacket.Header.RoomId]
			if !ok {
				errMsg := map[string]interface{}{"message": "Room ID Does Not Exist"}

				newPacket := ServerPacket{
					Header: Header{
						Protocol: "Error",
						RoomId:   "",
					},
					Response: errMsg,
				}

				select {
				case c.MsgCh <- newPacket:
					//s.removeClient(c)
				default:
					s.removeClient(c)
				}
			}
			//if room exists
			fmt.Println("Room Found! Sending Message to room: ", clientPacket.Header.RoomId)
			roomRef.IncomingMsgs <- clientPacket

		}
	}
}

func (s *Server) Stop() error {
	//Stop listene
	fmt.Println("Server stopped listening")

	//Shutdown Rooms
	for roomId, _ := range s.Rooms {
		delete(s.Rooms, roomId)
	}
	fmt.Println("Server stopped rooms")

	//Shutdown Clients Connections
	for conn := range s.Clients {
		conn.Close()
		delete(s.Clients, conn)
	}
	fmt.Println("Server stopped clients")
	fmt.Println("Server stopped Successfully")

	return nil
}
func (s *Server) Start() error {
	// Print the starting message before attempting to serve
	fmt.Println("Starting server on", s.Addr)

	// Set up the handler for new connections
	http.HandleFunc("/groupsession", s.handleNewConnection)

	// Run ListenAndServe in a separate goroutine to prevent blocking
	go func() {
		if err := http.ListenAndServe(s.Addr, nil); err != nil {
			fmt.Println("Server failed to start:", err)
		}
	}()

	return nil
}
