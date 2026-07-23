package game

import (
	"fmt"
	"guandan/internal/deck"
	"guandan/internal/rules"
	"guandan/internal/types"
	"math/rand"
	"sync"
	"time"
)

// Hub 定义游戏需要的广播接口（由 hub 包实现）
type Hub interface {
	BroadcastToRoom(roomID string, msg types.ServerMessage)
	SendToClient(clientID string, msg types.ServerMessage)
}

// Client 定义客户端接口（由 hub 包中的 Client 实现）
type Client interface {
	Send(msg types.ServerMessage)
	ID() string
	SeatIndex() int
	RoomID() string
	PlayerName() string
	SetSeatIndex(int)
	SetRoomID(string)
	SetPlayerName(string)
}

// Game 单局游戏
type Game struct {
	mu           sync.RWMutex
	io           Hub
	roomID       string
	players      []types.Player
	level        int
	phase        types.GamePhase
	hands        [][]types.Card
	currentTurn  int
	lastHand     *types.HandInfo
	passCount    int // 已连续过牌数（用于判断循环）
	roundActions map[int]types.RoundAction
	winners      []int
	tributeState *types.TributeState
	teamLevels   map[int]int
	activeTeam   int
	gameMode     types.GameMode
	skillCards   [][]types.SkillCard
	skipNextTurn []bool
	newCardIds   map[int][]string // 跟踪新增牌，用于高亮
	history      []types.HistoryEntry
	historyID    int
	currentRound int
	isActive     bool
	timeouts     []*time.Timer
	onGameEnd    func(winners []int) // 回调，由Match设置
}

func NewGame(roomID string, players []types.Player, mode types.GameMode, io Hub) *Game {
	g := &Game{
		roomID:       roomID,
		players:      players,
		gameMode:     mode,
		io:           io,
		phase:        types.PhaseWaiting,
		level:        2,
		hands:        make([][]types.Card, 4),
		roundActions: make(map[int]types.RoundAction),
		winners:      []int{},
		teamLevels:   map[int]int{0: 2, 1: 2},
		activeTeam:   0,
		skipNextTurn: []bool{false, false, false, false},
		newCardIds:   make(map[int][]string),
		history:      []types.HistoryEntry{},
		isActive:     true,
	}
	g.currentRound = 1
	return g
}

// Start 开始游戏
func (g *Game) Start() {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.phase != types.PhaseWaiting {
		return
	}
	g.startGame()
}

// ResetAndStart 重置并开始新一局（用于下一局）
func (g *Game) ResetAndStart(prevWinners []int) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if len(prevWinners) == 4 {
		g.handleLevelUp(prevWinners)
	}
	g.winners = []int{}
	g.tributeState = nil
	g.startGame()
}

func (g *Game) startGame() {
	g.phase = types.PhaseDealing
	if g.currentRound == 1 && len(g.winners) == 0 {
		g.activeTeam = 0
		g.teamLevels = map[int]int{0: 2, 1: 2}
		g.currentRound = 1
		g.history = []types.HistoryEntry{}
		g.historyID = 0
	} else {
		g.currentRound++
	}

	g.level = g.teamLevels[g.activeTeam]

	// 发牌
	deckCards := deck.CreateDeck()
	deck.ShuffleDeck(deckCards)
	g.hands = make([][]types.Card, 4)
	for i, c := range deckCards {
		g.hands[i%4] = append(g.hands[i%4], c)
	}
	// 更新属性并排序
	for i := 0; i < 4; i++ {
		g.hands[i] = deck.UpdateCardProperties(g.hands[i], g.level)
		g.hands[i] = rules.SortCards(g.hands[i], g.level)
	}

	g.skipNextTurn = []bool{false, false, false, false}

	if g.gameMode == types.ModeSkill {
		g.dealSkillCards()
	} else {
		g.skillCards = make([][]types.SkillCard, 4)
	}

	// 进贡逻辑（非第一局）
	if len(g.winners) > 0 {
		g.initTributePhase()
	} else {
		g.currentTurn = 0
		g.phase = types.PhasePlaying
		g.passCount = 0
		g.lastHand = nil
	}

	g.addHistoryEntry("GameStart", fmt.Sprintf("第%d局开始 - 等级:%d", g.currentRound, g.level), nil, nil)

	g.broadcastState()
	g.runBotTurn()
}

// 处理上一局结果升级
func (g *Game) handleLevelUp(prevWinners []int) {
	if len(prevWinners) < 4 {
		return
	}
	p1 := prevWinners[0]
	p2 := prevWinners[1]
	winningTeam := p1 % 2

	step := 1
	if (p1%2) == (p2%2) {
		step = 3
	} else if (p1%2) == (prevWinners[2]%2) {
		step = 2
	}

	if winningTeam != g.activeTeam {
		g.activeTeam = winningTeam
	}
	g.teamLevels[g.activeTeam] += step
	if g.teamLevels[g.activeTeam] > 14 {
		g.teamLevels[g.activeTeam] = 14
	}
}

func (g *Game) dealSkillCards() {
	// 从5种技能卡各2张，共10张，随机分配给4个玩家，每人2张
	pool := []types.SkillCard{}
	for _, st := range []types.SkillCardType{types.SkillDrawTwo, types.SkillSteal, types.SkillDiscard, types.SkillSkip, types.SkillHarvest} {
		pool = append(pool, types.SkillCard{ID: fmt.Sprintf("skill-%s-1", st), Type: st})
		pool = append(pool, types.SkillCard{ID: fmt.Sprintf("skill-%s-2", st), Type: st})
	}
	rand.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })
	g.skillCards = make([][]types.SkillCard, 4)
	for i := 0; i < 4; i++ {
		g.skillCards[i] = []types.SkillCard{pool[i*2], pool[i*2+1]}
	}
}

func (g *Game) initTributePhase() {
	// 根据上局结果决定进贡
	if len(g.winners) == 0 {
		g.phase = types.PhasePlaying
		g.currentTurn = g.activeTeam
		return
	}
	// 简化：只实现单进贡（最后一名进贡给第一名）
	p1 := g.winners[0]
	p4 := g.winners[3]
	if (p1%2) == (p4%2) {
		// 同队，无进贡
		g.phase = types.PhasePlaying
		g.currentTurn = p1
		return
	}
	// 检查抗贡：最后一名是否有两张大王？
	bigJokers := 0
	for _, c := range g.hands[p4] {
		if c.Rank == types.BigJoker {
			bigJokers++
		}
	}
	if bigJokers >= 2 {
		// 抗贡成功
		g.phase = types.PhasePlaying
		g.currentTurn = p1
		return
	}
	// 需要进贡
	g.tributeState = &types.TributeState{
		PendingTributes: []struct {
			From int  `json:"from"`
			To   int  `json:"to"`
			Card *types.Card `json:"card,omitempty"`
		}{{From: p4, To: p1}},
	}
	g.phase = types.PhaseTribute
	// 如果是Bot自动进贡
	g.processAutoTribute()
}

func (g *Game) processAutoTribute() {
	if g.tributeState == nil {
		return
	}
	for i := range g.tributeState.PendingTributes {
		t := &g.tributeState.PendingTributes[i]
		if g.players[t.From].IsBot && t.Card == nil {
			largest := rules.GetLargestCard(g.hands[t.From], g.level)
			t.Card = &largest
			// 从原手牌移除
			g.hands[t.From] = removeCard(g.hands[t.From], largest.ID)
			g.hands[t.To] = append(g.hands[t.To], largest)
			g.hands[t.To] = rules.SortCards(g.hands[t.To], g.level)
		}
	}
	allDone := true
	for _, t := range g.tributeState.PendingTributes {
		if t.Card == nil {
			allDone = false
			break
		}
	}
	if allDone {
		g.phase = types.PhaseReturnTribute
		// 构建还贡
		g.tributeState.PendingReturns = make([]struct {
			From int  `json:"from"`
			To   int  `json:"to"`
			Card *types.Card `json:"card,omitempty"`
		}, len(g.tributeState.PendingTributes))
		for i, t := range g.tributeState.PendingTributes {
			g.tributeState.PendingReturns[i] = struct {
				From int  `json:"from"`
				To   int  `json:"to"`
				Card *types.Card `json:"card,omitempty"`
			}{From: t.To, To: t.From}
		}
		g.tributeState.PendingTributes = nil
		g.processAutoReturn()
	}
}

func (g *Game) processAutoReturn() {
	if g.tributeState == nil {
		return
	}
	for i := range g.tributeState.PendingReturns {
		r := &g.tributeState.PendingReturns[i]
		if g.players[r.From].IsBot && r.Card == nil {
			// 还最小的一张
			if len(g.hands[r.From]) == 0 {
				continue
			}
			smallest := g.hands[r.From][len(g.hands[r.From])-1]
			r.Card = &smallest
			g.hands[r.From] = removeCard(g.hands[r.From], smallest.ID)
			g.hands[r.To] = append(g.hands[r.To], smallest)
			g.hands[r.To] = rules.SortCards(g.hands[r.To], g.level)
		}
	}
	g.checkReturnDone()
}

func (g *Game) checkReturnDone() {
	if g.tributeState == nil {
		return
	}
	allDone := true
	for _, r := range g.tributeState.PendingReturns {
		if r.Card == nil {
			allDone = false
			break
		}
	}
	if allDone {
		g.phase = types.PhasePlaying
		// 确定起始玩家
		if g.tributeState.NextStartPlayer != nil {
			g.currentTurn = *g.tributeState.NextStartPlayer
		} else {
			g.currentTurn = g.winners[0] // fallback
		}
		g.tributeState = nil
		g.broadcastState()
	}
}

func (g *Game) HandlePlayHand(seat int, cards []types.Card, handType *types.Hand) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.phase != types.PhasePlaying || g.currentTurn != seat || len(g.winners) >= 3 {
		return
	}

	// 验证牌是否在手牌中
	handMap := make(map[string]bool)
	for _, c := range g.hands[seat] {
		handMap[c.ID] = true
	}
	for _, c := range cards {
		if !handMap[c.ID] {
			g.emitError(seat, "你没有这些牌")
			return
		}
	}

	// 确定牌型
	var h *types.Hand
	if handType != nil {
		// 验证牌型是否有效
		// 这里可以调用 GetHandType 验证，简化：直接使用传入的
		h = handType
	} else {
		h = rules.GetHandType(cards, g.level)
		if h == nil {
			g.emitError(seat, "无效牌型")
			return
		}
	}

	// 如果是自由出牌（自己先出或上一轮自己出的），直接出
	if g.lastHand == nil || g.lastHand.PlayerIndex == seat {
		// 可以出
	} else {
		// 必须大于上一手
		cmp := rules.CompareHands(h, &g.lastHand.Hand)
		if cmp <= 0 {
			g.emitError(seat, "牌不够大")
			return
		}
	}

	// 移除出的牌
	for _, c := range cards {
		g.hands[seat] = removeCard(g.hands[seat], c.ID)
	}

	g.lastHand = &types.HandInfo{PlayerIndex: seat, Hand: *h}
	g.roundActions[seat] = types.RoundAction{Type: "play", Cards: cards, Hand: h}
	g.addHistoryEntry("Play", fmt.Sprintf("%s 出牌: %s", g.players[seat].Name, h.Type), &seat, nil)

	if len(g.hands[seat]) == 0 {
		g.winners = append(g.winners, seat)
		pos := []string{"第一名", "第二名", "第三名", "第四名"}[len(g.winners)-1]
		g.addHistoryEntry("PlayerFinish", fmt.Sprintf("%s %s", g.players[seat].Name, pos), &seat, nil)

		if len(g.winners) == 2 {
			p1 := g.winners[0]
			p2 := g.winners[1]
			if (p1%2) == (p2%2) {
				// 双扣
				losers := []int{}
				for i := 0; i < 4; i++ {
					if !contains(g.winners, i) {
						losers = append(losers, i)
					}
				}
				g.winners = append(g.winners, losers...)
				g.endGame()
				return
			}
		}
		if len(g.winners) == 3 {
			last := 0
			for i := 0; i < 4; i++ {
				if !contains(g.winners, i) {
					last = i
					break
				}
			}
			g.winners = append(g.winners, last)
			g.endGame()
			return
		}
	}

	g.advanceTurn()
	g.broadcastState()
	g.runBotTurn()
}

func (g *Game) HandlePass(seat int) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.phase != types.PhasePlaying || g.currentTurn != seat {
		return
	}
	if g.lastHand == nil || g.lastHand.PlayerIndex == seat {
		g.emitError(seat, "不能过牌")
		return
	}

	g.roundActions[seat] = types.RoundAction{Type: "pass"}
	g.addHistoryEntry("Pass", fmt.Sprintf("%s 过牌", g.players[seat].Name), &seat, nil)

	g.advanceTurn()
	g.broadcastState()
	g.runBotTurn()
}

func (g *Game) advanceTurn() {
	if len(g.winners) >= 3 {
		return
	}
	next := (g.currentTurn + 1) % 4
	// 跳过已出完和乐不思蜀
	for i := 0; i < 4; i++ {
		if len(g.hands[next]) == 0 {
			next = (next + 1) % 4
			continue
		}
		if g.skipNextTurn[next] {
			g.skipNextTurn[next] = false
			g.io.BroadcastToRoom(g.roomID, types.ServerMessage{Type: "error", Payload: fmt.Sprintf("%s 被跳过回合", g.players[next].Name)})
			next = (next + 1) % 4
			continue
		}
		break
	}

	// 检查是否循环到出牌人
	if g.lastHand != nil && next == g.lastHand.PlayerIndex {
		g.endRoundAndFindNext(g.lastHand.PlayerIndex)
		return
	}

	g.currentTurn = next
}

func (g *Game) endRoundAndFindNext(winner int) {
	// 接风：winner的队友先出
	if len(g.hands[winner]) == 0 {
		next := (winner + 2) % 4
		// 如果队友也出完了，继续找
		for i := 0; i < 4; i++ {
			if len(g.hands[next]) > 0 {
				g.currentTurn = next
				g.lastHand = nil
				g.roundActions = make(map[int]types.RoundAction)
				return
			}
			next = (next + 1) % 4
		}
		// 全出完了，游戏结束
		g.endGame()
		return
	}
	// 否则winner出
	g.currentTurn = winner
	g.lastHand = nil
	g.roundActions = make(map[int]types.RoundAction)
}

func (g *Game) endGame() {
	g.phase = types.PhaseScore
	g.addHistoryEntry("GameEnd", fmt.Sprintf("游戏结束，排名: %v", g.winners), nil, nil)
	g.broadcastState()
	if g.onGameEnd != nil {
		g.onGameEnd(g.winners)
	}
}

func (g *Game) HandleTribute(seat int, cards []types.Card) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.phase != types.PhaseTribute || g.tributeState == nil {
		return
	}
	if len(cards) != 1 {
		g.emitError(seat, "请选择一张牌")
		return
	}
	// 查找对应进贡
	for i, t := range g.tributeState.PendingTributes {
		if t.From == seat && t.Card == nil {
			// 验证是否是最大牌
			largest := rules.GetLargestCard(g.hands[seat], g.level)
			if rules.GetLogicValue(cards[0].Rank, g.level) < rules.GetLogicValue(largest.Rank, g.level) {
				g.emitError(seat, "必须进贡最大牌")
				return
			}
			g.tributeState.PendingTributes[i].Card = &cards[0]
			g.hands[seat] = removeCard(g.hands[seat], cards[0].ID)
			g.hands[t.To] = append(g.hands[t.To], cards[0])
			g.hands[t.To] = rules.SortCards(g.hands[t.To], g.level)
			break
		}
	}
	g.checkTributeDone()
}

func (g *Game) checkTributeDone() {
	allDone := true
	for _, t := range g.tributeState.PendingTributes {
		if t.Card == nil {
			allDone = false
			break
		}
	}
	if allDone {
		// 进入还贡
		g.phase = types.PhaseReturnTribute
		g.tributeState.PendingReturns = make([]struct {
			From int  `json:"from"`
			To   int  `json:"to"`
			Card *types.Card `json:"card,omitempty"`
		}, len(g.tributeState.PendingTributes))
		for i, t := range g.tributeState.PendingTributes {
			g.tributeState.PendingReturns[i] = struct {
				From int  `json:"from"`
				To   int  `json:"to"`
				Card *types.Card `json:"card,omitempty"`
			}{From: t.To, To: t.From}
		}
		g.tributeState.PendingTributes = nil
		g.processAutoReturn()
		g.broadcastState()
	}
}

func (g *Game) HandleReturnTribute(seat int, cards []types.Card) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.phase != types.PhaseReturnTribute || g.tributeState == nil {
		return
	}
	if len(cards) != 1 {
		g.emitError(seat, "请选择一张牌")
		return
	}
	for i, r := range g.tributeState.PendingReturns {
		if r.From == seat && r.Card == nil {
			g.tributeState.PendingReturns[i].Card = &cards[0]
			g.hands[seat] = removeCard(g.hands[seat], cards[0].ID)
			g.hands[r.To] = append(g.hands[r.To], cards[0])
			g.hands[r.To] = rules.SortCards(g.hands[r.To], g.level)
			break
		}
	}
	g.checkReturnDone()
}

// HandleUseSkill 使用技能卡
func (g *Game) HandleUseSkill(seat int, skillID string, targetSeat *int) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.gameMode != types.ModeSkill || g.phase != types.PhasePlaying || g.currentTurn != seat {
		g.emitError(seat, "无法使用技能")
		return
	}
	if len(g.winners) >= 3 {
		g.emitError(seat, "游戏已结束")
		return
	}

	// 查找技能卡
	var skill *types.SkillCard
	var idx int
	for i, s := range g.skillCards[seat] {
		if s.ID == skillID {
			skill = &s
			idx = i
			break
		}
	}
	if skill == nil {
		g.emitError(seat, "没有这张技能卡")
		return
	}

	// 验证目标
	needTarget := skill.Type == types.SkillSteal || skill.Type == types.SkillDiscard || skill.Type == types.SkillSkip
	if needTarget && (targetSeat == nil || *targetSeat == seat || len(g.hands[*targetSeat]) == 0) {
		g.emitError(seat, "无效目标")
		return
	}

	// 应用效果
	success := g.applySkill(skill.Type, seat, targetSeat)
	if success {
		g.skillCards[seat] = append(g.skillCards[seat][:idx], g.skillCards[seat][idx+1:]...)
		g.addHistoryEntry("SkillUse", fmt.Sprintf("%s 使用 %s", g.players[seat].Name, skill.Type), &seat, nil)
		g.broadcastState()
	}
}

func (g *Game) applySkill(st types.SkillCardType, user int, target *int) bool {
	switch st {
	case types.SkillDrawTwo:
		// 无中生有：获得两张随机牌
		for i := 0; i < 2; i++ {
			card := generateRandomCard(g.level)
			g.hands[user] = append(g.hands[user], card)
			g.hands[user] = rules.SortCards(g.hands[user], g.level)
			g.newCardIds[user] = append(g.newCardIds[user], card.ID)
		}
		return true
	case types.SkillSteal:
		if target == nil {
			return false
		}
		if len(g.hands[*target]) == 0 {
			return false
		}
		idx := rand.Intn(len(g.hands[*target]))
		stolen := g.hands[*target][idx]
		g.hands[*target] = append(g.hands[*target][:idx], g.hands[*target][idx+1:]...)
		g.hands[user] = append(g.hands[user], stolen)
		g.hands[user] = rules.SortCards(g.hands[user], g.level)
		g.newCardIds[user] = append(g.newCardIds[user], stolen.ID)
		return true
	case types.SkillDiscard:
		if target == nil {
			return false
		}
		if len(g.hands[*target]) == 0 {
			return false
		}
		idx := rand.Intn(len(g.hands[*target]))
		g.hands[*target] = append(g.hands[*target][:idx], g.hands[*target][idx+1:]...)
		return true
	case types.SkillSkip:
		if target == nil {
			return false
		}
		g.skipNextTurn[*target] = true
		return true
	case types.SkillHarvest:
		// 五谷丰登：所有活着的玩家获得一张随机牌
		for i := 0; i < 4; i++ {
			if len(g.hands[i]) > 0 && !contains(g.winners, i) {
				card := generateRandomCard(g.level)
				g.hands[i] = append(g.hands[i], card)
				g.hands[i] = rules.SortCards(g.hands[i], g.level)
				g.newCardIds[i] = append(g.newCardIds[i], card.ID)
			}
		}
		return true
	default:
		return false
	}
}

func (g *Game) emitError(seat int, msg string) {
	g.io.SendToClient(g.players[seat].ID, types.ServerMessage{Type: "error", Payload: msg})
}

func (g *Game) broadcastState() {
	g.mu.RLock()
	defer g.mu.RUnlock()

	// 构建GameState
	state := types.GameState{
		Phase:        g.phase,
		Level:        g.level,
		CurrentTurn:  g.currentTurn,
		RoundActions: g.roundActions,
		Winners:      g.winners,
		TeamLevels:   g.teamLevels,
		ActiveTeam:   g.activeTeam,
		GameMode:     g.gameMode,
		SkipNextTurn: g.skipNextTurn,
		History:      g.history,
		CurrentRound: g.currentRound,
	}
	if g.lastHand != nil {
		state.LastHand = g.lastHand
	}
	if g.tributeState != nil {
		state.TributeState = g.tributeState
	}
	// 处理每个玩家的手牌
	handsForClient := make([]interface{}, 4)
	for i := 0; i < 4; i++ {
		if i == g.currentTurn {
			// 当前玩家看到自己的牌（完整）
			handsForClient[i] = g.hands[i]
		} else {
			// 其他玩家只看到牌数
			handsForClient[i] = len(g.hands[i])
		}
	}
	state.Hands = handsForClient

	// 每个玩家单独发送
	for i, p := range g.players {
		if p.IsBot || p.ID == "" {
			continue
		}
		// 复制状态，修改hands为自己能看到的样子
		stateCopy := state
		// 重新构造hands
		handsCopy := make([]interface{}, 4)
		for j := 0; j < 4; j++ {
			if j == i {
				handsCopy[j] = g.hands[j]
			} else {
				handsCopy[j] = len(g.hands[j])
			}
		}
		stateCopy.Hands = handsCopy
		// 技能卡只发送给自己
		if g.gameMode == types.ModeSkill {
			stateCopy.MySkillCards = g.skillCards[i]
		}
		// 新牌高亮
		if ids, ok := g.newCardIds[i]; ok && len(ids) > 0 {
			stateCopy.NewCardIds = ids
			// 清除，只发送一次
			delete(g.newCardIds, i)
		}
		g.io.SendToClient(p.ID, types.ServerMessage{Type: "gameState", Payload: stateCopy})
	}
}

func (g *Game) runBotTurn() {
	if len(g.winners) >= 3 || g.phase != types.PhasePlaying {
		return
	}
	// 检查当前回合是不是Bot
	seat := g.currentTurn
	if seat >= 0 && seat < 4 && g.players[seat].IsBot {
		// 使用定时器延迟执行
		time.AfterFunc(1500*time.Millisecond, func() {
			g.mu.Lock()
			defer g.mu.Unlock()
			if g.isActive && g.phase == types.PhasePlaying && g.currentTurn == seat && len(g.winners) < 3 {
				g.handleBotTurn(seat)
			}
		})
	}
}

func (g *Game) handleBotTurn(seat int) {
	// 已经在锁内
	hand := g.hands[seat]
	if len(hand) == 0 {
		g.advanceTurn()
		g.broadcastState()
		g.runBotTurn()
		return
	}

	// 技能模式，Bot可能使用技能
	if g.gameMode == types.ModeSkill && len(g.skillCards[seat]) > 0 {
		// 简单决策：随机使用
		if rand.Float64() < 0.2 {
			skill := g.skillCards[seat][0]
			var target *int
			if skill.Type == types.SkillSteal || skill.Type == types.SkillDiscard || skill.Type == types.SkillSkip {
				// 找一个对手
				for i := 0; i < 4; i++ {
					if i != seat && (i%2) != (seat%2) && len(g.hands[i]) > 0 {
						target = &i
						break
					}
				}
			}
			g.applySkill(skill.Type, seat, target)
			g.skillCards[seat] = g.skillCards[seat][1:]
			g.broadcastState()
			// 继续出牌
			time.AfterFunc(500*time.Millisecond, func() {
				g.mu.Lock()
				defer g.mu.Unlock()
				if g.isActive && g.phase == types.PhasePlaying && g.currentTurn == seat {
					g.handleBotTurn(seat)
				}
			})
			return
		}
	}

	// 普通出牌
	bot := NewBot(hand, g.level)
	target := g.lastHand
	if target != nil && target.PlayerIndex == seat {
		target = nil // 自由出牌
	}
	var targetHand *types.Hand
	if target != nil {
		targetHand = &target.Hand
	}
	move := bot.DecideMove(targetHand)

	if move != nil && len(move) > 0 {
		// 获取牌型
		h := rules.GetHandType(move, g.level)
		if h != nil {
			// 检查是否合法
			if target == nil || rules.CompareHands(h, &target.Hand) > 0 {
				// 出牌
				for _, c := range move {
					g.hands[seat] = removeCard(g.hands[seat], c.ID)
				}
				g.lastHand = &types.HandInfo{PlayerIndex: seat, Hand: *h}
				g.roundActions[seat] = types.RoundAction{Type: "play", Cards: move, Hand: h}
				if len(g.hands[seat]) == 0 {
					g.winners = append(g.winners, seat)
					if len(g.winners) == 2 {
						p1 := g.winners[0]
						p2 := g.winners[1]
						if (p1%2) == (p2%2) {
							losers := []int{}
							for i := 0; i < 4; i++ {
								if !contains(g.winners, i) {
									losers = append(losers, i)
								}
							}
							g.winners = append(g.winners, losers...)
							g.endGame()
							g.broadcastState()
							return
						}
					}
					if len(g.winners) == 3 {
						last := 0
						for i := 0; i < 4; i++ {
							if !contains(g.winners, i) {
								last = i
								break
							}
						}
						g.winners = append(g.winners, last)
						g.endGame()
						g.broadcastState()
						return
					}
				}
				g.advanceTurn()
				g.broadcastState()
				g.runBotTurn()
				return
			}
		}
	}

	// 无法出牌，过
	g.roundActions[seat] = types.RoundAction{Type: "pass"}
	g.advanceTurn()
	g.broadcastState()
	g.runBotTurn()
}

func (g *Game) addHistoryEntry(eventType string, message string, playerIndex *int, details interface{}) {
	g.historyID++
	entry := types.HistoryEntry{
		ID:          fmt.Sprintf("%d", g.historyID),
		Timestamp:   time.Now().UnixMilli(),
		Type:        eventType,
		Message:     message,
		Details:     details,
	}
	if playerIndex != nil {
		entry.PlayerIndex = playerIndex
		entry.PlayerName = g.players[*playerIndex].Name
	}
	g.history = append(g.history, entry)
	// 广播给所有玩家
	g.io.BroadcastToRoom(g.roomID, types.ServerMessage{Type: "historyUpdate", Payload: entry})
}

// 辅助函数
func removeCard(slice []types.Card, id string) []types.Card {
	result := make([]types.Card, 0, len(slice))
	for _, c := range slice {
		if c.ID != id {
			result = append(result, c)
		}
	}
	return result
}

func contains(slice []int, val int) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}

func generateRandomCard(level int) types.Card {
	suits := []types.Suit{types.Spades, types.Hearts, types.Clubs, types.Diamonds}
	ranks := []types.Rank{types.Two, types.Three, types.Four, types.Five, types.Six, types.Seven,
		types.Eight, types.Nine, types.Ten, types.Jack, types.Queen, types.King, types.Ace}
	if rand.Float64() < 0.05 {
		// 大小王
		if rand.Float64() < 0.5 {
			return types.Card{Suit: types.Joker, Rank: types.SmallJoker, ID: fmt.Sprintf("gen-%d", rand.Int63())}
		}
		return types.Card{Suit: types.Joker, Rank: types.BigJoker, ID: fmt.Sprintf("gen-%d", rand.Int63())}
	}
	suit := suits[rand.Intn(len(suits))]
	rank := ranks[rand.Intn(len(ranks))]
	id := fmt.Sprintf("gen-%d", rand.Int63())
	card := types.Card{Suit: suit, Rank: rank, ID: id}
	if int(rank) == level {
		card.IsLevelCard = true
		if suit == types.Hearts {
			card.IsWild = true
		}
	}
	return card
}