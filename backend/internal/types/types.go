package types

import "encoding/json"

// Suit 花色
type Suit int

const (
	Spades Suit = iota
	Hearts
	Clubs
	Diamonds
	Joker
)

// Rank 牌面
type Rank int

const (
	Two Rank = iota + 2
	Three
	Four
	Five
	Six
	Seven
	Eight
	Nine
	Ten
	Jack
	Queen
	King
	Ace
	SmallJoker
	BigJoker
)

// Card 单张牌
type Card struct {
	Suit        Suit   `json:"suit"`
	Rank        Rank   `json:"rank"`
	ID          string `json:"id"`
	IsLevelCard bool   `json:"isLevelCard,omitempty"`
	IsWild      bool   `json:"isWild,omitempty"`
}

// HandType 牌型
type HandType string

const (
	HandTypeSingle       HandType = "Single"
	HandTypePair         HandType = "Pair"
	HandTypeTrips        HandType = "Trips"
	HandTypeTripsWithPair HandType = "TripsWithPair"
	HandTypeStraight     HandType = "Straight"
	HandTypeTube         HandType = "Tube"
	HandTypePlate        HandType = "Plate"
	HandTypeBomb         HandType = "Bomb"
	HandTypeStraightFlush HandType = "StraightFlush"
	HandTypeFourKings    HandType = "FourKings"
)

// Hand 一手牌
type Hand struct {
	Type     HandType `json:"type"`
	Cards    []Card   `json:"cards"`
	Value    int      `json:"value"`
	BombCount *int    `json:"bombCount,omitempty"`
}

// GameMode 游戏模式
type GameMode string

const (
	ModeNormal GameMode = "Normal"
	ModeSkill  GameMode = "Skill"
)

// SkillCardType 技能卡类型
type SkillCardType string

const (
	SkillDrawTwo  SkillCardType = "DrawTwo"
	SkillSteal    SkillCardType = "Steal"
	SkillDiscard  SkillCardType = "Discard"
	SkillSkip     SkillCardType = "Skip"
	SkillHarvest  SkillCardType = "Harvest"
)

// SkillCard 技能卡
type SkillCard struct {
	ID   string         `json:"id"`
	Type SkillCardType `json:"type"`
}

// Player 玩家信息
type Player struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	SeatIndex      int    `json:"seatIndex"`
	IsReady        bool   `json:"isReady"`
	IsBot          bool   `json:"isBot,omitempty"`
	IsDisconnected bool   `json:"isDisconnected,omitempty"`
}

// GamePhase 游戏阶段
type GamePhase string

const (
	PhaseWaiting      GamePhase = "Waiting"
	PhaseDealing      GamePhase = "Dealing"
	PhaseTribute      GamePhase = "Tribute"
	PhaseReturnTribute GamePhase = "ReturnTribute"
	PhasePlaying      GamePhase = "Playing"
	PhaseScore        GamePhase = "Score"
)

// TributeState 进贡状态
type TributeState struct {
	PendingTributes []struct {
		From int  `json:"from"`
		To   int  `json:"to"`
		Card *Card `json:"card,omitempty"`
	} `json:"pendingTributes"`
	PendingReturns []struct {
		From int  `json:"from"`
		To   int  `json:"to"`
		Card *Card `json:"card,omitempty"`
	} `json:"pendingReturns"`
	NextStartPlayer *int `json:"nextStartPlayer,omitempty"`
}

// RoundAction 回合动作
type RoundAction struct {
	Type  string `json:"type"` // "play" or "pass"
	Cards []Card `json:"cards,omitempty"`
	Hand  *Hand  `json:"hand,omitempty"`
}

// GameState 游戏状态（发送给客户端）
type GameState struct {
	Phase         GamePhase              `json:"phase"`
	Level         int                    `json:"level"`
	CurrentTurn   int                    `json:"currentTurn"`
	Hands         interface{}            `json:"hands"` // 对己方是 []Card，对其他是 int (牌数)
	LastHand      *HandInfo              `json:"lastHand,omitempty"`
	RoundActions  map[int]RoundAction    `json:"roundActions,omitempty"`
	Winners       []int                  `json:"winners"`
	TributeState  *TributeState          `json:"tributeState,omitempty"`
	TeamLevels    map[int]int            `json:"teamLevels"`
	ActiveTeam    int                    `json:"activeTeam"`
	GameMode      GameMode               `json:"gameMode"`
	MySkillCards  []SkillCard            `json:"mySkillCards,omitempty"`
	SkipNextTurn  []bool                 `json:"skipNextTurn"`
	NewCardIds    []string               `json:"newCardIds,omitempty"`
	History       []HistoryEntry         `json:"history"`
	CurrentRound  int                    `json:"currentRound"`
}

// HandInfo 上一次出牌信息
type HandInfo struct {
	PlayerIndex int  `json:"playerIndex"`
	Hand        Hand `json:"hand"`
}

// HistoryEntry 历史记录
type HistoryEntry struct {
	ID          string `json:"id"`
	Timestamp   int64  `json:"timestamp"`
	Type        string `json:"type"`
	PlayerIndex *int   `json:"playerIndex,omitempty"`
	PlayerName  string `json:"playerName,omitempty"`
	Message     string `json:"message"`
	Details     interface{} `json:"details,omitempty"`
}

// RoomState 房间状态
type RoomState struct {
	RoomID   string    `json:"roomId"`
	Players  []*Player `json:"players"`
	GameMode GameMode  `json:"gameMode"`
}

// RoomInfo 房间列表项
type RoomInfo struct {
	ID          string   `json:"id"`
	PlayerCount int      `json:"playerCount"`
	MaxPlayers  int      `json:"maxPlayers"`
	InGame      bool     `json:"inGame"`
	GameMode    GameMode `json:"gameMode"`
	HostName    string   `json:"hostName"`
}

// ChatMessage 聊天消息
type ChatMessage struct {
	Sender    string `json:"sender"`
	Text      string `json:"text"`
	Time      string `json:"time"`
	SeatIndex int    `json:"seatIndex"`
}

// ========== WebSocket 消息定义 ==========

// ClientMessage 客户端发来的消息
type ClientMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// ServerMessage 服务端发往客户端的消息
type ServerMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// 消息类型常量
const (
	// 客户端 -> 服务端
	MsgJoinRoom        = "joinRoom"
	MsgGetRoomList     = "getRoomList"
	MsgReady           = "ready"
	MsgStart           = "start"
	MsgPlayHand        = "playHand"
	MsgPass            = "pass"
	MsgTribute         = "tribute"
	MsgReturnTribute   = "returnTribute"
	MsgUseSkill        = "useSkill"
	MsgChat            = "chatMessage"
	MsgSwitchSeat      = "switchSeat"
	MsgSetGameMode     = "setGameMode"
	MsgForceEndGame    = "forceEndGame"

	// 服务端 -> 客户端
	MsgRoomState       = "roomState"
	MsgGameState       = "gameState"
	MsgChatMessage     = "chatMessage"
	MsgError           = "error"
	MsgGameOver        = "gameOver"
	MsgMatchOver       = "matchOver"
	MsgGameTerminated  = "gameTerminated"
	MsgRoomList        = "roomList"
	MsgHistoryUpdate   = "historyUpdate"
)