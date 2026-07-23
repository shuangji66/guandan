package game

import (
	"guandan/internal/types"
	"time"
)

type Match struct {
	io             Hub
	roomID         string
	players        []types.Player
	gameMode       types.GameMode
	currentGame    *Game
	teamLevels     map[int]int
	activeTeam     int
	consecutiveWins map[int]int
	matchWinner    *int
	lastWinners    []int
}

func NewMatch(roomID string, players []types.Player, mode types.GameMode, io Hub) *Match {
	return &Match{
		io:              io,
		roomID:          roomID,
		players:         players,
		gameMode:        mode,
		teamLevels:      map[int]int{0: 2, 1: 2},
		activeTeam:      0,
		consecutiveWins: map[int]int{0: 0, 1: 0},
	}
}

func (m *Match) StartMatch() {
	m.startNextGame()
}

func (m *Match) startNextGame() {
	if m.matchWinner != nil {
		return
	}
	if m.currentGame != nil {
		m.currentGame.isActive = false // 停止旧游戏
	}
	// 创建新游戏
	game := NewGame(m.roomID, m.players, m.gameMode, m.io)
	game.teamLevels = m.teamLevels
	game.activeTeam = m.activeTeam
	game.currentRound = 1
	if len(m.lastWinners) > 0 {
		game.winners = m.lastWinners // 用于进贡
	}
	game.onGameEnd = func(winners []int) {
		m.handleGameEnd(winners)
	}
	m.currentGame = game
	game.Start()
}

func (m *Match) handleGameEnd(winners []int) {
	if len(winners) != 4 {
		return
	}
	p1 := winners[0]
	p2 := winners[1]
	winningTeam := p1 % 2

	// 计算升级步数
	step := 1
	if (p1%2) == (p2%2) {
		step = 3
	} else if (p1%2) == (winners[2]%2) {
		step = 2
	}

	// 更新等级
	//oldLevel := m.teamLevels[winningTeam]
	m.teamLevels[winningTeam] += step
	if m.teamLevels[winningTeam] > 14 {
		m.teamLevels[winningTeam] = 14
	}

	// 更新庄家
	if winningTeam != m.activeTeam {
		m.activeTeam = winningTeam
	}

	// 判断是否达到A
	if m.teamLevels[winningTeam] == 14 {
		m.consecutiveWins[winningTeam]++
		other := 1 - winningTeam
		m.consecutiveWins[other] = 0
		if m.consecutiveWins[winningTeam] >= 2 {
			// 比赛结束
			m.matchWinner = &winningTeam
			m.io.BroadcastToRoom(m.roomID, types.ServerMessage{
				Type: "matchOver",
				Payload: map[string]interface{}{
					"winningTeam":  winningTeam,
					"winners":      m.players,
					"finalLevels":  m.teamLevels,
				},
			})
			return
		}
	} else {
		m.consecutiveWins[0] = 0
		m.consecutiveWins[1] = 0
	}

	// 保存赢家用于下局进贡
	m.lastWinners = winners

	// 3秒后开始下一局
	time.AfterFunc(3*time.Second, func() {
		m.startNextGame()
	})
}

func (m *Match) ForceEnd() {
	if m.currentGame != nil {
		m.currentGame.isActive = false
	}
	m.matchWinner = nil
	m.consecutiveWins = map[int]int{0: 0, 1: 0}
	m.io.BroadcastToRoom(m.roomID, types.ServerMessage{Type: "gameTerminated", Payload: nil})
}

func (m *Match) GetCurrentGame() *Game {
	return m.currentGame
}