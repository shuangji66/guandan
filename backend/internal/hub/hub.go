package hub

import (
	"encoding/json"
	"guandan/internal/game"
	"guandan/internal/types"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Hub struct {
	clients    map[*Client]bool
	rooms      map[string]*game.Room
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

// 确保 Hub 实现了 game.Hub 接口
var _ game.Hub = (*Hub)(nil)

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		rooms:      make(map[string]*game.Room),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.sendChan)
				if client.RoomID() != "" {
					if room, ok := h.rooms[client.RoomID()]; ok {
						room.HandleDisconnect(client)
					}
				}
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := NewClient(conn.RemoteAddr().String(), conn, h)
	h.register <- client

	go client.WritePump()
	go client.ReadPump()
}

func (h *Hub) handleMessage(client *Client, msg types.ClientMessage) {
	switch msg.Type {
	case types.MsgJoinRoom:
		var payload struct {
			PlayerName string `json:"playerName"`
			RoomID     string `json:"roomId"`
		}
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return
		}
		room := h.getOrCreateRoom(payload.RoomID)
		room.AddPlayer(client, payload.PlayerName)

	case types.MsgGetRoomList:
		roomList := h.getRoomList()
		client.Send(types.ServerMessage{Type: "roomList", Payload: roomList})

	case types.MsgReady:
		if room, ok := h.rooms[client.RoomID()]; ok {
			room.SetReady(client)
		}
	case types.MsgStart:
		if room, ok := h.rooms[client.RoomID()]; ok {
			room.StartGame(client)
		}
	case types.MsgPlayHand:
		var payload struct {
			Cards    []types.Card `json:"cards"`
			HandType *types.Hand  `json:"handType,omitempty"`
		}
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return
		}
		if room, ok := h.rooms[client.RoomID()]; ok && room.Match() != nil && room.Match().GetCurrentGame() != nil {
			room.Match().GetCurrentGame().HandlePlayHand(client.SeatIndex(), payload.Cards, payload.HandType)
		}
	case types.MsgPass:
		if room, ok := h.rooms[client.RoomID()]; ok && room.Match() != nil && room.Match().GetCurrentGame() != nil {
			room.Match().GetCurrentGame().HandlePass(client.SeatIndex())
		}
	case types.MsgTribute:
		var payload struct {
			Cards []types.Card `json:"cards"`
		}
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return
		}
		if room, ok := h.rooms[client.RoomID()]; ok && room.Match() != nil && room.Match().GetCurrentGame() != nil {
			room.Match().GetCurrentGame().HandleTribute(client.SeatIndex(), payload.Cards)
		}
	case types.MsgReturnTribute:
		var payload struct {
			Cards []types.Card `json:"cards"`
		}
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return
		}
		if room, ok := h.rooms[client.RoomID()]; ok && room.Match() != nil && room.Match().GetCurrentGame() != nil {
			room.Match().GetCurrentGame().HandleReturnTribute(client.SeatIndex(), payload.Cards)
		}
	case types.MsgUseSkill:
		var payload struct {
			SkillID    string `json:"skillId"`
			TargetSeat *int   `json:"targetSeat,omitempty"`
		}
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return
		}
		if room, ok := h.rooms[client.RoomID()]; ok && room.Match() != nil && room.Match().GetCurrentGame() != nil {
			room.Match().GetCurrentGame().HandleUseSkill(client.SeatIndex(), payload.SkillID, payload.TargetSeat)
		}
	case types.MsgChat:
		var payload struct {
			Text string `json:"text"`
		}
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return
		}
		if room, ok := h.rooms[client.RoomID()]; ok {
			room.HandleChat(client, payload.Text)
		}
	case types.MsgSwitchSeat:
		var payload struct {
			SeatIndex int `json:"seatIndex"`
		}
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return
		}
		if room, ok := h.rooms[client.RoomID()]; ok {
			room.SwitchSeat(client, payload.SeatIndex)
		}
	case types.MsgSetGameMode:
		var payload struct {
			Mode types.GameMode `json:"mode"`
		}
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return
		}
		if room, ok := h.rooms[client.RoomID()]; ok {
			room.SetGameMode(client, payload.Mode)
		}
	case types.MsgForceEndGame:
		if room, ok := h.rooms[client.RoomID()]; ok {
			room.ForceEndGame(client)
		}
	}
}

func (h *Hub) getOrCreateRoom(roomID string) *game.Room {
	h.mu.Lock()
	defer h.mu.Unlock()
	if room, ok := h.rooms[roomID]; ok {
		return room
	}
	room := game.NewRoom(roomID, h)
	h.rooms[roomID] = room
	return room
}

func (h *Hub) getRoomList() []types.RoomInfo {
	h.mu.RLock()
	defer h.mu.RUnlock()
	list := []types.RoomInfo{}
	for id, room := range h.rooms {
		playerCount := 0
		for _, p := range room.Players() {
			if p != nil && !p.IsDisconnected {
				playerCount++
			}
		}
		inGame := room.Match() != nil && room.Match().GetCurrentGame() != nil
		hostName := ""
		if room.Players()[0] != nil {
			hostName = room.Players()[0].Name
		}
		list = append(list, types.RoomInfo{
			ID:          id,
			PlayerCount: playerCount,
			MaxPlayers:  4,
			InGame:      inGame,
			GameMode:    room.GameMode(),
			HostName:    hostName,
		})
	}
	return list
}

// BroadcastToRoom 实现 game.Hub 接口
func (h *Hub) BroadcastToRoom(roomID string, msg types.ServerMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.clients {
		if client.RoomID() == roomID {
			client.Send(msg)
		}
	}
}

// SendToClient 实现 game.Hub 接口
func (h *Hub) SendToClient(clientID string, msg types.ServerMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.clients {
		if client.ID() == clientID {
			client.Send(msg)
			return
		}
	}
}