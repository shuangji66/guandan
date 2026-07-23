package game

import (
	"guandan/internal/types"
	"time"
)

type Room struct {
	id       string
	io       Hub
	players  []*types.Player
	match    *Match
	gameMode types.GameMode
}

func NewRoom(id string, io Hub) *Room {
	return &Room{
		id:       id,
		io:       io,
		players:  make([]*types.Player, 4),
		gameMode: types.ModeNormal,
	}
}

func (r *Room) AddPlayer(client Client, name string) {
	// 检查重连
	for _, p := range r.players {
		if p != nil && p.Name == name && p.IsDisconnected {
			p.IsDisconnected = false
			p.ID = client.ID()
			p.IsReady = false
			client.SetSeatIndex(p.SeatIndex)
			client.SetRoomID(r.id)
			client.SetPlayerName(p.Name)
			r.io.BroadcastToRoom(r.id, types.ServerMessage{Type: "error", Payload: name + " 重新连接"})
			r.broadcastState()
			return
		}
	}

	// 找空位
	seat := -1
	for i, p := range r.players {
		if p == nil {
			seat = i
			break
		}
	}
	if seat == -1 {
		client.Send(types.ServerMessage{Type: "error", Payload: "房间已满"})
		return
	}

	player := &types.Player{
		ID:        client.ID(),
		Name:      name,
		SeatIndex: seat,
		IsReady:   false,
	}
	r.players[seat] = player
	client.SetSeatIndex(seat)
	client.SetRoomID(r.id)
	client.SetPlayerName(name)

	r.broadcastState()
}

func (r *Room) HandleDisconnect(client Client) {
	for _, p := range r.players {
		if p != nil && p.ID == client.ID() {
			p.IsDisconnected = true
			p.IsReady = false
			if r.match != nil && r.match.currentGame != nil {
				r.io.BroadcastToRoom(r.id, types.ServerMessage{Type: "error", Payload: p.Name + " 断开连接"})
			} else {
				for i, pl := range r.players {
					if pl != nil && pl.ID == client.ID() {
						r.players[i] = nil
					}
				}
			}
			r.broadcastState()
			break
		}
	}
}

func (r *Room) SetReady(client Client) {
	for _, p := range r.players {
		if p != nil && p.ID == client.ID() {
			p.IsReady = true
			r.broadcastState()
			r.tryAutoStart()
			break
		}
	}
}

func (r *Room) StartGame(client Client) {
	if client.SeatIndex() != 0 {
		client.Send(types.ServerMessage{Type: "error", Payload: "只有房主可以开始游戏"})
		return
	}
	if r.match != nil && r.match.matchWinner == nil {
		client.Send(types.ServerMessage{Type: "error", Payload: "游戏正在进行"})
		return
	}
	r.startGame()
}

func (r *Room) startGame() {
	// 填充 Bot
	for i, p := range r.players {
		if p == nil {
			bot := &types.Player{
				ID:        "bot-" + string(rune(i)),
				Name:      "Bot " + string(rune(i+'0')),
				SeatIndex: i,
				IsReady:   true,
				IsBot:     true,
			}
			r.players[i] = bot
		}
	}
	r.broadcastState()

	r.match = NewMatch(r.id, convertPlayers(r.players), r.gameMode, r.io)
	r.match.StartMatch()
}

func (r *Room) SetGameMode(client Client, mode types.GameMode) {
	if client.SeatIndex() != 0 {
		client.Send(types.ServerMessage{Type: "error", Payload: "只有房主可以切换模式"})
		return
	}
	if r.match != nil && r.match.matchWinner == nil {
		client.Send(types.ServerMessage{Type: "error", Payload: "游戏进行中不能切换模式"})
		return
	}
	r.gameMode = mode
	r.io.BroadcastToRoom(r.id, types.ServerMessage{Type: "error", Payload: "游戏模式切换为: " + string(mode)})
	r.broadcastState()
}

func (r *Room) SwitchSeat(client Client, targetSeat int) {
	if targetSeat < 0 || targetSeat > 3 {
		return
	}
	if r.match != nil && r.match.matchWinner == nil {
		client.Send(types.ServerMessage{Type: "error", Payload: "游戏进行中不能换座"})
		return
	}
	currentSeat := client.SeatIndex()
	if currentSeat == targetSeat {
		return
	}
	if r.players[targetSeat] != nil {
		client.Send(types.ServerMessage{Type: "error", Payload: "该座位已被占用"})
		return
	}
	player := r.players[currentSeat]
	r.players[currentSeat] = nil
	player.SeatIndex = targetSeat
	r.players[targetSeat] = player
	client.SetSeatIndex(targetSeat)
	r.broadcastState()
}

func (r *Room) ForceEndGame(client Client) {
	if client.SeatIndex() != 0 {
		client.Send(types.ServerMessage{Type: "error", Payload: "只有房主可以强制结束"})
		return
	}
	if r.match == nil {
		client.Send(types.ServerMessage{Type: "error", Payload: "没有正在进行的对局"})
		return
	}
	r.match.ForceEnd()
	r.match = nil
	r.io.BroadcastToRoom(r.id, types.ServerMessage{Type: "error", Payload: "房主强制结束了对局"})
	r.broadcastState()
}

func (r *Room) HandleChat(client Client, text string) {
	for _, p := range r.players {
		if p != nil && p.ID == client.ID() {
			r.io.BroadcastToRoom(r.id, types.ServerMessage{
				Type: "chatMessage",
				Payload: types.ChatMessage{
					Sender:    p.Name,
					Text:      text,
					Time:      time.Now().Format("15:04:05"),
					SeatIndex: p.SeatIndex,
				},
			})
			break
		}
	}
}

func (r *Room) tryAutoStart() {
	readyCount := 0
	for _, p := range r.players {
		if p != nil && p.IsReady {
			readyCount++
		}
	}
	if readyCount == 4 && r.match == nil {
		r.startGame()
	}
}

func (r *Room) broadcastState() {
	playersState := make([]*types.Player, 4)
	for i, p := range r.players {
		if p != nil {
			playersState[i] = &types.Player{
				ID:             p.ID,
				Name:           p.Name,
				SeatIndex:      p.SeatIndex,
				IsReady:        p.IsReady,
				IsBot:          p.IsBot,
				IsDisconnected: p.IsDisconnected,
			}
		}
	}
	state := types.RoomState{
		RoomID:   r.id,
		Players:  playersState,
		GameMode: r.gameMode,
	}
	r.io.BroadcastToRoom(r.id, types.ServerMessage{Type: "roomState", Payload: state})
}

func (r *Room) Players() []*types.Player {
	return r.players
}

func (r *Room) Match() *Match {
	return r.match
}

func (r *Room) GameMode() types.GameMode {
	return r.gameMode
}

func convertPlayers(players []*types.Player) []types.Player {
	result := make([]types.Player, 0, len(players))
	for _, p := range players {
		if p != nil {
			result = append(result, *p)
		}
	}
	return result
}