package hub

import (
	"encoding/json"
	"guandan/internal/types"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	id         string
	conn       *websocket.Conn
	hub        *Hub
	sendChan   chan []byte // 改为小写，避免与 Send 方法冲突
	roomID     string
	playerName string
	seatIndex  int
	mu         sync.Mutex
}

func NewClient(id string, conn *websocket.Conn, hub *Hub) *Client {
	return &Client{
		id:        id,
		conn:      conn,
		hub:       hub,
		sendChan:  make(chan []byte, 256),
		seatIndex: -1,
	}
}

// 实现 game.Client 接口的方法
func (c *Client) ID() string {
	return c.id
}

func (c *Client) SeatIndex() int {
	return c.seatIndex
}

func (c *Client) RoomID() string {
	return c.roomID
}

func (c *Client) PlayerName() string {
	return c.playerName
}

func (c *Client) SetSeatIndex(idx int) {
	c.seatIndex = idx
}

func (c *Client) SetRoomID(roomID string) {
	c.roomID = roomID
}

func (c *Client) SetPlayerName(name string) {
	c.playerName = name
}

// Send 发送消息
func (c *Client) Send(msg types.ServerMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	select {
	case c.sendChan <- data:
	default:
		// 队列满，丢弃
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(4096)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		var msg types.ClientMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}
		c.hub.handleMessage(c, msg)
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.sendChan:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}