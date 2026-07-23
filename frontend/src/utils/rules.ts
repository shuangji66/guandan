import { Card, Hand, HandType, Rank, Suit } from '@/types'

export function getLogicValue(rank: Rank, level: number): number {
  if (rank === Rank.SmallJoker) return 20
  if (rank === Rank.BigJoker) return 21
  if (rank === level) return 19
  if (rank === Rank.Ace) return 14
  return rank
}

export function getAllPossibleHandTypes(cards: Card[], level: number): Hand[] {
  // 简化实现：直接调用 getHandType，但该函数需移植
  // 这里仅作示意
  return []
}

export function getHandDescription(hand: Hand): string {
  // 返回描述
  return hand.type
}