package deck

import (
	"guandan/internal/types"
	"math/rand"
	"time"
)

// CreateDeck 生成两副牌（共108张）
func CreateDeck() []types.Card {
	var cards []types.Card
	suits := []types.Suit{types.Spades, types.Hearts, types.Clubs, types.Diamonds}
	ranks := []types.Rank{
		types.Two, types.Three, types.Four, types.Five, types.Six, types.Seven,
		types.Eight, types.Nine, types.Ten, types.Jack, types.Queen, types.King, types.Ace,
	}

	for deck := 0; deck < 2; deck++ {
		for _, suit := range suits {
			for _, rank := range ranks {
				cards = append(cards, types.Card{
					Suit: suit,
					Rank: rank,
					ID:   generateID(),
				})
			}
		}
		// 大小王
		cards = append(cards, types.Card{Suit: types.Joker, Rank: types.SmallJoker, ID: generateID()})
		cards = append(cards, types.Card{Suit: types.Joker, Rank: types.BigJoker, ID: generateID()})
	}
	return cards
}

// ShuffleDeck 洗牌
func ShuffleDeck(cards []types.Card) []types.Card {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(cards), func(i, j int) {
		cards[i], cards[j] = cards[j], cards[i]
	})
	return cards
}

// UpdateCardProperties 根据当前级牌更新牌的属性（级牌、红心级牌万能）
func UpdateCardProperties(cards []types.Card, level int) []types.Card {
	updated := make([]types.Card, len(cards))
	for i, c := range cards {
		c.IsLevelCard = int(c.Rank) == level
		c.IsWild = c.IsLevelCard && c.Suit == types.Hearts
		updated[i] = c
	}
	return updated
}

func generateID() string {
	// 简单生成唯一ID，实际可使用uuid或更复杂的方法
	return randString(8)
}

func randString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}