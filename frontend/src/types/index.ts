// 花色
export enum Suit {
  Spades = 0,
  Hearts = 1,
  Clubs = 2,
  Diamonds = 3,
  Joker = 4,
}

// 牌面
export enum Rank {
  Two = 2,
  Three = 3,
  Four = 4,
  Five = 5,
  Six = 6,
  Seven = 7,
  Eight = 8,
  Nine = 9,
  Ten = 10,
  Jack = 11,
  Queen = 12,
  King = 13,
  Ace = 14,
  SmallJoker = 15,
  BigJoker = 16,
}

export interface Card {
  suit: Suit
  rank: Rank
  id: string
  isLevelCard?: boolean
  isWild?: boolean
}

export enum HandType {
  Single = 'Single',
  Pair = 'Pair',
  Trips = 'Trips',
  TripsWithPair = 'TripsWithPair',
  Straight = 'Straight',
  Tube = 'Tube',
  Plate = 'Plate',
  Bomb = 'Bomb',
  StraightFlush = 'StraightFlush',
  FourKings = 'FourKings',
}

export interface Hand {
  type: HandType
  cards: Card[]
  value: number
  bombCount?: number
}

export enum GameMode {
  Normal = 'Normal',
  Skill = 'Skill',
}

export enum SkillCardType {
  DrawTwo = 'DrawTwo',
  Steal = 'Steal',
  Discard = 'Discard',
  Skip = 'Skip',
  Harvest = 'Harvest',
}

export const SkillCardNames: Record<SkillCardType, string> = {
  [SkillCardType.DrawTwo]: '无中生有',
  [SkillCardType.Steal]: '顺手牵羊',
  [SkillCardType.Discard]: '过河拆桥',
  [SkillCardType.Skip]: '乐不思蜀',
  [SkillCardType.Harvest]: '五谷丰登',
}

export interface SkillCard {
  id: string
  type: SkillCardType
}

export enum GamePhase {
  Waiting = 'Waiting',
  Dealing = 'Dealing',
  Tribute = 'Tribute',
  ReturnTribute = 'ReturnTribute',
  Playing = 'Playing',
  Score = 'Score',
}

export interface TributeState {
  pendingTributes: { from: number; to: number; card?: Card }[]
  pendingReturns: { from: number; to: number; card?: Card }[]
  nextStartPlayer?: number
}

export interface RoundAction {
  type: 'play' | 'pass'
  cards?: Card[]
  hand?: Hand
}

export interface GameState {
  phase: GamePhase
  level: number
  currentTurn: number
  hands: (Card[] | number)[]    // 对自己是 Card[]，对其他是牌数
  lastHand?: { playerIndex: number; hand: Hand }
  roundActions?: Record<number, RoundAction>
  winners: number[]
  tributeState?: TributeState
  teamLevels: Record<number, number>
  activeTeam: number
  gameMode: GameMode
  mySkillCards?: SkillCard[]
  skipNextTurn: boolean[]
  newCardIds?: string[]
  history: HistoryEntry[]
  currentRound: number
}

export interface RoomState {
  roomId: string
  players: (Player | null)[]
  gameMode: GameMode
}

export interface Player {
  id: string
  name: string
  seatIndex: number
  isReady: boolean
  isBot?: boolean
  isDisconnected?: boolean
}

export interface RoomInfo {
  id: string
  playerCount: number
  maxPlayers: number
  inGame: boolean
  gameMode: GameMode
  hostName: string
}

export interface ChatMessage {
  sender: string
  text: string
  time: string
  seatIndex: number
}

export interface HistoryEntry {
  id: string
  timestamp: number
  type: string
  playerIndex?: number
  playerName?: string
  message: string
  details?: any
}

// WebSocket 消息
export interface ClientMessage {
  type: string
  payload: any
}

export interface ServerMessage {
  type: string
  payload: any
}