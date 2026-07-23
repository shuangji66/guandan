package rules

import (
	"guandan/internal/types"
	"sort"
)

// GetLogicValue 获取牌的逻辑值（用于排序和比较）
func GetLogicValue(rank types.Rank, level int) int {
	if rank == types.SmallJoker {
		return 20
	}
	if rank == types.BigJoker {
		return 21
	}
	if int(rank) == level {
		return 19 // 级牌最大（除大小王外）
	}
	if rank == types.Ace {
		return 14
	}
	return int(rank)
}

// SortCards 按逻辑值降序排序
func SortCards(cards []types.Card, level int) []types.Card {
	sorted := make([]types.Card, len(cards))
	copy(sorted, cards)
	sort.Slice(sorted, func(i, j int) bool {
		vi := GetLogicValue(sorted[i].Rank, level)
		vj := GetLogicValue(sorted[j].Rank, level)
		if vi != vj {
			return vi > vj
		}
		return sorted[i].Suit < sorted[j].Suit
	})
	return sorted
}

// IsConsecutive 检查一组值是否连续（处理A-2-3-4-5）
func IsConsecutive(values []int) bool {
	if len(values) < 2 {
		return false
	}
	sorted := make([]int, len(values))
	copy(sorted, values)
	sort.Ints(sorted)

	// 特殊：A-2-3-4-5 (即 14,2,3,4,5 -> 转化为1,2,3,4,5)
	if sorted[len(sorted)-1] == 14 && sorted[0] == 2 && len(sorted) == 5 {
		// 检查是否2,3,4,5
		if sorted[0] == 2 && sorted[1] == 3 && sorted[2] == 4 && sorted[3] == 5 {
			return true
		}
	}

	for i := 0; i < len(sorted)-1; i++ {
		if sorted[i+1] != sorted[i]+1 {
			return false
		}
	}
	return true
}

// GetHandType 分析手牌类型
func GetHandType(cards []types.Card, level int) *types.Hand {
	if len(cards) == 0 {
		return nil
	}

	lenCards := len(cards)

	// 统计非万能牌
	var wildCount int
	nonWilds := make([]types.Card, 0, len(cards))
	for _, c := range cards {
		if c.IsWild {
			wildCount++
		} else {
			nonWilds = append(nonWilds, c)
		}
	}

	// 构建 value -> count 映射 (非万能)
	counts := make(map[int]int)
	for _, c := range nonWilds {
		v := GetLogicValue(c.Rank, level)
		counts[v]++
	}
	uniqueVals := make([]int, 0, len(counts))
	for v := range counts {
		uniqueVals = append(uniqueVals, v)
	}
	sort.Slice(uniqueVals, func(i, j int) bool { return uniqueVals[i] > uniqueVals[j] })

	maxCount := 0
	for _, cnt := range counts {
		if cnt > maxCount {
			maxCount = cnt
		}
	}

	// 1. Four Kings
	if lenCards == 4 {
		smallJokers := 0
		bigJokers := 0
		for _, c := range cards {
			if c.Rank == types.SmallJoker {
				smallJokers++
			} else if c.Rank == types.BigJoker {
				bigJokers++
			}
		}
		if smallJokers == 2 && bigJokers == 2 {
			return &types.Hand{Type: types.HandTypeFourKings, Cards: cards, Value: 999}
		}
	}

	// 2. Single
	if lenCards == 1 {
		val := 19
		if !cards[0].IsWild {
			val = GetLogicValue(cards[0].Rank, level)
		}
		return &types.Hand{Type: types.HandTypeSingle, Cards: cards, Value: val}
	}

	// 3. Pair
	if lenCards == 2 {
		if wildCount == 2 {
			return &types.Hand{Type: types.HandTypePair, Cards: cards, Value: 19}
		}
		if wildCount == 1 {
			if nonWilds[0].Rank > types.Ace {
				return nil
			}
			return &types.Hand{Type: types.HandTypePair, Cards: cards, Value: GetLogicValue(nonWilds[0].Rank, level)}
		}
		if len(uniqueVals) == 1 {
			return &types.Hand{Type: types.HandTypePair, Cards: cards, Value: uniqueVals[0]}
		}
	}

	// 4. Trips
	if lenCards == 3 {
		if wildCount == 3 {
			return &types.Hand{Type: types.HandTypeTrips, Cards: cards, Value: 19}
		}
		if len(uniqueVals) == 1 && counts[uniqueVals[0]]+wildCount == 3 {
			if uniqueVals[0] > 19 {
				return nil
			}
			return &types.Hand{Type: types.HandTypeTrips, Cards: cards, Value: uniqueVals[0]}
		}
	}

	// 5. Trips with Pair (Full House)
	if lenCards == 5 {
		if maxCount+wildCount == 5 {
			bombCount := 5
			return &types.Hand{Type: types.HandTypeBomb, Cards: cards, Value: uniqueVals[0], BombCount: &bombCount}
		}
		for _, tVal := range uniqueVals {
			if tVal > 19 {
				continue
			}
			tCount := counts[tVal]
			wildsForTrips := 3 - tCount
			if wildCount >= wildsForTrips {
				remWilds := wildCount - wildsForTrips
				otherVals := []int{}
				for _, v := range uniqueVals {
					if v != tVal {
						otherVals = append(otherVals, v)
					}
				}
				if len(otherVals) == 0 {
					continue
				}
				if len(otherVals) == 1 {
					pVal := otherVals[0]
					if pVal > 19 {
						continue
					}
					if counts[pVal]+remWilds >= 2 {
						return &types.Hand{Type: types.HandTypeTripsWithPair, Cards: cards, Value: tVal}
					}
				}
			}
		}
	}

	// 6. Straight (5 cards)
	if lenCards == 5 {
		// 只考虑非万能且非Joker的牌
		ranks := []int{}
		for _, c := range nonWilds {
			if c.Rank <= types.Ace {
				ranks = append(ranks, int(c.Rank))
			}
		}
		// 去重
		rankSet := make(map[int]bool)
		for _, r := range ranks {
			rankSet[r] = true
		}
		uniqueRanks := make([]int, 0, len(rankSet))
		for r := range rankSet {
			uniqueRanks = append(uniqueRanks, r)
		}
		if len(uniqueRanks) == 5 {
			// 检查是否连续（包括A-2-3-4-5）
			if IsConsecutive(uniqueRanks) {
				// 确定顺子的最大值
				sortedRanks := make([]int, len(uniqueRanks))
				copy(sortedRanks, uniqueRanks)
				sort.Ints(sortedRanks)
				// 处理A-2-3-4-5
				top := sortedRanks[4]
				if top == 14 && sortedRanks[0] == 2 {
					top = 5
				}
				// 检查是否同花（同花顺）
				suits := make(map[types.Suit]bool)
				for _, c := range cards {
					suits[c.Suit] = true
				}
				if len(suits) == 1 {
					bombCount := 5
					return &types.Hand{Type: types.HandTypeStraightFlush, Cards: cards, Value: top, BombCount: &bombCount}
				}
				return &types.Hand{Type: types.HandTypeStraight, Cards: cards, Value: top}
			}
		}
	}

	// 7. Bomb (4+ cards)
	if lenCards >= 4 {
		if maxCount+wildCount == lenCards {
			if len(uniqueVals) <= 1 {
				val := 19
				if len(uniqueVals) > 0 {
					val = uniqueVals[0]
				}
				if val <= 19 {
					bombCount := lenCards
					return &types.Hand{Type: types.HandTypeBomb, Cards: cards, Value: val, BombCount: &bombCount}
				}
			}
		}
	}

	// 8. Tube (6 cards) 和 Plate (6 cards) - 简化处理，仅自然牌
	if lenCards == 6 && wildCount == 0 {
		sortedByRank := make([]types.Card, len(cards))
		copy(sortedByRank, cards)
		sort.Slice(sortedByRank, func(i, j int) bool {
			return sortedByRank[i].Rank < sortedByRank[j].Rank
		})
		// Tube: 三对连续
		if sortedByRank[0].Rank == sortedByRank[1].Rank &&
			sortedByRank[2].Rank == sortedByRank[3].Rank &&
			sortedByRank[4].Rank == sortedByRank[5].Rank {
			if sortedByRank[2].Rank == sortedByRank[0].Rank+1 &&
				sortedByRank[4].Rank == sortedByRank[2].Rank+1 {
				return &types.Hand{Type: types.HandTypeTube, Cards: cards, Value: int(sortedByRank[4].Rank)}
			}
			// A-2-3 特殊情况
			if sortedByRank[4].Rank == types.Ace && sortedByRank[0].Rank == types.Two && sortedByRank[2].Rank == types.Three {
				return &types.Hand{Type: types.HandTypeTube, Cards: cards, Value: 3}
			}
		}
		// Plate: 两个三连
		if sortedByRank[0].Rank == sortedByRank[1].Rank && sortedByRank[1].Rank == sortedByRank[2].Rank &&
			sortedByRank[3].Rank == sortedByRank[4].Rank && sortedByRank[4].Rank == sortedByRank[5].Rank {
			if sortedByRank[3].Rank == sortedByRank[0].Rank+1 {
				return &types.Hand{Type: types.HandTypePlate, Cards: cards, Value: int(sortedByRank[3].Rank)}
			}
		}
	}

	return nil
}

// CompareHands 比较两手牌，返回 >0 表示 a 大于 b
func CompareHands(a, b *types.Hand) int {
	if a.Type == types.HandTypeFourKings {
		return 1
	}
	if b.Type == types.HandTypeFourKings {
		return -1
	}

	isBombA := a.Type == types.HandTypeBomb || a.Type == types.HandTypeStraightFlush
	isBombB := b.Type == types.HandTypeBomb || b.Type == types.HandTypeStraightFlush

	if isBombA && !isBombB {
		return 1
	}
	if !isBombA && isBombB {
		return -1
	}

	if isBombA && isBombB {
		score := func(h *types.Hand) int {
			if h.Type == types.HandTypeStraightFlush {
				return 55 // SF 视为 5.5 但用整数放大
			}
			if h.BombCount != nil {
				return *h.BombCount * 10
			}
			return 40
		}
		sA := score(a)
		sB := score(b)
		if sA != sB {
			return sA - sB
		}
		return a.Value - b.Value
	}

	// 普通牌型
	if a.Type != b.Type {
		return 0
	}
	if len(a.Cards) != len(b.Cards) {
		return 0
	}
	return a.Value - b.Value
}

// GetLargestCard 获取手牌中最大的一张（按逻辑值）
func GetLargestCard(cards []types.Card, level int) types.Card {
	sorted := SortCards(cards, level)
	if len(sorted) == 0 {
		return types.Card{}
	}
	return sorted[0]
}

// GetAllPossibleHandTypes 获取所有可能的牌型解释（用于万能牌）
func GetAllPossibleHandTypes(cards []types.Card, level int) []types.Hand {
	// 简化实现：只返回一个解释
	h := GetHandType(cards, level)
	if h == nil {
		return nil
	}
	return []types.Hand{*h}
}