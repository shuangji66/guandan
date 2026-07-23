package game

import (
	"guandan/internal/rules"
	"guandan/internal/types"
	"sort"
)

type Bot struct {
	cards []types.Card
	level int
}

func NewBot(cards []types.Card, level int) *Bot {
	return &Bot{
		cards: rules.SortCards(cards, level),
		level: level,
	}
}

// DecideMove 决定出牌或过牌
func (b *Bot) DecideMove(target *types.Hand) []types.Card {
	if len(b.cards) == 0 {
		return nil
	}
	if target == nil {
		// 自由出牌：尝试出三带二、三张、对子、单张等
		// 优先出三带二
		trips := b.getGroups(3)
		if len(trips) > 0 {
			pair := b.findPairExcluding(trips[0])
			if pair != nil {
				return append(trips[0], pair...)
			}
			return trips[0]
		}
		pairs := b.getGroups(2)
		if len(pairs) > 0 {
			return pairs[0]
		}
		// 出最小单张
		return []types.Card{b.cards[len(b.cards)-1]}
	}

	// 需要压牌
	move := b.findBeat(target)
	if move != nil {
		return move
	}

	// 尝试炸弹
	bomb := b.findBomb(target)
	if bomb != nil {
		return bomb
	}

	return nil
}

func (b *Bot) findBeat(target *types.Hand) []types.Card {
	switch target.Type {
	case types.HandTypeSingle:
		for i := len(b.cards) - 1; i >= 0; i-- {
			if rules.GetLogicValue(b.cards[i].Rank, b.level) > target.Value {
				return []types.Card{b.cards[i]}
			}
		}
	case types.HandTypePair:
		pairs := b.getGroups(2) // 从小到大
		for _, p := range pairs {
			if rules.GetLogicValue(p[0].Rank, b.level) > target.Value {
				return p
			}
		}
	case types.HandTypeTrips:
		trips := b.getGroups(3)
		for _, t := range trips {
			if rules.GetLogicValue(t[0].Rank, b.level) > target.Value {
				return t
			}
		}
	case types.HandTypeTripsWithPair:
		trips := b.getGroups(3)
		for _, t := range trips {
			if rules.GetLogicValue(t[0].Rank, b.level) > target.Value {
				pair := b.findPairExcluding(t)
				if pair != nil {
					return append(t, pair...)
				}
			}
		}
		// 其他牌型暂不实现，可扩展
	}
	return nil
}

func (b *Bot) findPairExcluding(exclude []types.Card) []types.Card {
	excludeMap := make(map[string]bool)
	for _, c := range exclude {
		excludeMap[c.ID] = true
	}
	var available []types.Card
	for _, c := range b.cards {
		if !excludeMap[c.ID] {
			available = append(available, c)
		}
	}
	pairs := b.getGroupsFrom(available, 2)
	if len(pairs) > 0 {
		return pairs[0]
	}
	return nil
}

func (b *Bot) getGroups(size int) [][]types.Card {
	return b.getGroupsFrom(b.cards, size)
}

func (b *Bot) getGroupsFrom(cards []types.Card, size int) [][]types.Card {
	if len(cards) == 0 {
		return nil
	}
	sorted := rules.SortCards(cards, b.level)
	var groups [][]types.Card
	var current []types.Card
	for _, c := range sorted {
		if len(current) == 0 || rules.GetLogicValue(c.Rank, b.level) == rules.GetLogicValue(current[0].Rank, b.level) {
			current = append(current, c)
		} else {
			if len(current) >= size {
				groups = append(groups, current[:size])
			}
			current = []types.Card{c}
		}
	}
	if len(current) >= size {
		groups = append(groups, current[:size])
	}
	// 从小到大排序
	sort.Slice(groups, func(i, j int) bool {
		return rules.GetLogicValue(groups[i][0].Rank, b.level) < rules.GetLogicValue(groups[j][0].Rank, b.level)
	})
	return groups
}

func (b *Bot) findBomb(target *types.Hand) []types.Card {
	// 1. 获取4张以上炸弹
	bombs := b.getBombs()
	// 2. 获取同花顺（简化：仅自然5张同花顺）
	sfs := b.findStraightFlushes()
	// 3. 四大天王
	kings := b.findFourKings()

	if target == nil {
		if len(bombs) > 0 {
			return bombs[0].Cards
		}
		if len(sfs) > 0 {
			return sfs[0]
		}
		if kings != nil {
			return kings
		}
		return nil
	}

	// 目标是否为炸弹家族
	targetIsBomb := target.Type == types.HandTypeBomb || target.Type == types.HandTypeStraightFlush || target.Type == types.HandTypeFourKings
	if !targetIsBomb {
		if len(bombs) > 0 {
			return bombs[0].Cards
		}
		if len(sfs) > 0 {
			return sfs[0]
		}
		if kings != nil {
			return kings
		}
		return nil
	}

	if target.Type == types.HandTypeFourKings {
		return nil
	}

	if target.Type == types.HandTypeStraightFlush {
		// 找更大的同花顺
		//for _, sf := range sfs {
		//	// 需要计算SF的值，这里简单用value
		//	// 实际我们需要计算SF的值，但简化：假设sfs按值排序
		//}
		// 或四大天王
		if kings != nil {
			return kings
		}
		// 或6张以上炸弹
		for _, b := range bombs {
			if len(b.Cards) >= 6 {
				return b.Cards
			}
		}
		return nil
	}

	if target.Type == types.HandTypeBomb {
		tCount := 4
		if target.BombCount != nil {
			tCount = *target.BombCount
		}
		tVal := target.Value

		for _, b := range bombs {
			bCount := len(b.Cards)
			bVal := b.Value
			if bCount > tCount || (bCount == tCount && bVal > tVal) {
				return b.Cards
			}
		}

		if tCount <= 5 && len(sfs) > 0 {
			return sfs[0]
		}
		if kings != nil {
			return kings
		}
	}
	return nil
}

func (b *Bot) getBombs() []struct{ Cards []types.Card; Value int } {
	groups := b.getGroups(4)
	var result []struct{ Cards []types.Card; Value int }
	for _, g := range groups {
		result = append(result, struct {
			Cards []types.Card
			Value int
		}{Cards: g, Value: rules.GetLogicValue(g[0].Rank, b.level)})
	}
	return result
}

func (b *Bot) findStraightFlushes() [][]types.Card {
	// 简化：仅检测自然同花顺（非万能）
	var result [][]types.Card
	suits := []types.Suit{types.Spades, types.Hearts, types.Clubs, types.Diamonds}
	for _, suit := range suits {
		var suitCards []types.Card
		for _, c := range b.cards {
			if c.Suit == suit && c.Rank <= types.Ace && !c.IsWild {
				suitCards = append(suitCards, c)
			}
		}
		if len(suitCards) < 5 {
			continue
		}
		sort.Slice(suitCards, func(i, j int) bool {
			return suitCards[i].Rank < suitCards[j].Rank
		})
		for i := 0; i <= len(suitCards)-5; i++ {
			window := suitCards[i : i+5]
			ranks := make([]int, 5)
			for j, c := range window {
				ranks[j] = int(c.Rank)
			}
			if rules.IsConsecutive(ranks) {
				result = append(result, window)
			}
		}
	}
	// 按值排序（升序）
	sort.Slice(result, func(i, j int) bool {
		// 取最大值
		maxI := result[i][4].Rank
		maxJ := result[j][4].Rank
		return maxI < maxJ
	})
	return result
}

func (b *Bot) findFourKings() []types.Card {
	var small, big []types.Card
	for _, c := range b.cards {
		if c.Rank == types.SmallJoker {
			small = append(small, c)
		} else if c.Rank == types.BigJoker {
			big = append(big, c)
		}
	}
	if len(small) >= 2 && len(big) >= 2 {
		return append(small[:2], big[:2]...)
	}
	return nil
}