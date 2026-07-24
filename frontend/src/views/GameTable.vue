<template>
  <div class="relative w-full h-screen bg-[#1e1e1e] overflow-hidden flex items-center justify-center font-mono">
    <div class="absolute inset-20 border-2 border-[#333333] rounded-xl opacity-50 pointer-events-none"></div>

    <!-- 顶部玩家 -->
    <PlayerArea top="1rem" left="50%" transform="translateX(-50%)" :player="topPlayer" :seat="topSeat"
      :hand-count="topHandCount" :is-teammate="topIsTeammate" :is-opponent="topIsOpponent" :action="topAction"
      :winner-label="topWinnerLabel" :chat-bubble="topChatBubble" :is-my-turn="gameState?.currentTurn === topSeat"
      :game-exists="!!gameState" @click-seat="handleSwitchSeat" />

    <!-- 左侧玩家 -->
    <PlayerArea left="2rem" top="50%" transform="translateY(-50%)" :player="leftPlayer" :seat="leftSeat"
      :hand-count="leftHandCount" :is-teammate="leftIsTeammate" :is-opponent="leftIsOpponent" :action="leftAction"
      :winner-label="leftWinnerLabel" :chat-bubble="leftChatBubble" :is-my-turn="gameState?.currentTurn === leftSeat"
      :game-exists="!!gameState" @click-seat="handleSwitchSeat" />

    <!-- 右侧玩家 -->
    <PlayerArea right="2rem" top="50%" transform="translateY(-50%)" :player="rightPlayer" :seat="rightSeat"
      :hand-count="rightHandCount" :is-teammate="rightIsTeammate" :is-opponent="rightIsOpponent" :action="rightAction"
      :winner-label="rightWinnerLabel" :chat-bubble="rightChatBubble" :is-my-turn="gameState?.currentTurn === rightSeat"
      :game-exists="!!gameState" @click-seat="handleSwitchSeat" />

    <!-- 聊天框 -->
    <ChatBox :messages="chatMessages" @send="handleSendChat" />

    <!-- 等级显示 + 强制结束 -->
    <div v-if="gameState" class="absolute top-4 left-4 flex flex-col gap-2 items-start z-50">
      <div class="text-[#d4d4d4] font-bold text-xl bg-[#252526] border border-[#333333] px-4 py-2 rounded shadow-lg">
        <span class="text-[#569cd6]">const</span> <span class="text-[#9cdcfe]">Level</span> =
        <span class="text-[#b5cea8]">{{ gameState.level }}</span>;
      </div>
      <button
        v-if="mySeat === 0"
        @click="forceEnd"
        class="bg-red-900/80 hover:bg-red-600 text-white text-xs px-3 py-1 rounded border border-red-500/50 shadow-lg backdrop-blur-sm transition-all flex items-center gap-1"
      >
        <span>⛔</span> 强制结束
      </button>
    </div>

    <!-- 中间区域：上一手牌 + 等待/开始 -->
    <div class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 z-40">
      <div v-if="gameState?.lastHand" class="bg-green-700/50 p-4 rounded-lg flex flex-col items-center">
        <div class="text-white mb-2 font-bold">{{ getPlayerName(gameState.lastHand.playerIndex) }} 出牌:</div>
        <div class="flex -space-x-8">
          <CardComponent v-for="card in gameState.lastHand.hand.cards" :key="card.id" :card="card" />
        </div>
        <div class="text-yellow-300 font-bold mt-2">{{ gameState.lastHand.hand.type }}</div>
      </div>

      <!-- 等待/开始 -->
      <div v-if="!gameState" class="flex flex-col gap-4 mt-8 items-center">
        <div class="text-white text-xl">Waiting for players...</div>
        <div class="flex items-center gap-4 bg-[#252526] px-4 py-2 rounded-lg border border-[#333333]">
          <span class="text-[#9cdcfe] font-bold">模式:</span>
          <button
            @click="setGameMode(GameMode.Normal)"
            :disabled="mySeat !== 0"
            class="px-4 py-1 rounded font-bold transition-all"
            :class="[roomState?.gameMode !== GameMode.Skill ? 'bg-blue-600 text-white' : 'bg-gray-600 text-gray-300 hover:bg-gray-500', mySeat !== 0 ? 'cursor-not-allowed opacity-70' : '']"
          >
            普通
          </button>
          <button
            @click="setGameMode(GameMode.Skill)"
            :disabled="mySeat !== 0"
            class="px-4 py-1 rounded font-bold transition-all"
            :class="[roomState?.gameMode === GameMode.Skill ? 'bg-purple-600 text-white' : 'bg-gray-600 text-gray-300 hover:bg-gray-500', mySeat !== 0 ? 'cursor-not-allowed opacity-70' : '']"
          >
            技能
          </button>
        </div>
        <div v-if="roomState?.gameMode === GameMode.Skill" class="text-purple-400 text-sm">技能模式: 每人开局获得2张技能卡</div>
        <button v-if="myPlayer && !myPlayer.isReady" @click="sendMessage('ready', {})" class="bg-blue-500 text-white px-6 py-2 rounded font-bold">
          准备
        </button>
        <button v-if="mySeat === 0" @click="sendMessage('start', {})" class="bg-yellow-500 text-black px-6 py-2 rounded font-bold">
          开始游戏 (Host)
        </button>
      </div>
    </div>

    <!-- 底部：技能卡 + 操作按钮 + 手牌 -->
    <div class="absolute bottom-0 w-full flex flex-col items-center pb-4 z-20 pointer-events-none">
      <!-- 技能卡 -->
      <div v-if="gameState?.gameMode === GameMode.Skill && gameState.mySkillCards?.length"
        class="mb-4 pointer-events-auto flex flex-col items-center">
        <div class="text-purple-400 text-sm mb-2 font-bold">我的技能卡</div>
        <div class="flex gap-3">
          <SkillCardButton v-for="skill in gameState.mySkillCards" :key="skill.id" :skill="skill"
            @click="handleSkillClick(skill)"
            :disabled="gameState.currentTurn !== mySeat || gameState.phase !== 'Playing'" />
        </div>
        <div v-if="gameState.currentTurn === mySeat && gameState.phase === 'Playing'"
          class="text-xs text-gray-400 mt-1">
          点击技能卡使用（使用后仍可出牌）
        </div>
      </div>

      <!-- 操作按钮 -->
      <div class="mb-8 pointer-events-auto">
        <div v-if="gameState?.currentTurn === mySeat && gameState.phase === 'Playing'" class="flex gap-4">
          <button @click="toggleViewMode" class="bg-gray-600 hover:bg-gray-700 text-white px-4 py-2 rounded-full font-bold shadow-lg mr-4">
            {{ viewMode === 'normal' ? '切换同花顺视图' : '切换普通视图' }}
          </button>
          <button @click="handleHint" class="bg-yellow-500 hover:bg-yellow-600 text-black px-4 py-2 rounded-full font-bold shadow-lg mr-4">
            提示
          </button>
          <button @click="handlePlay" :disabled="selectedCards.length === 0" class="bg-blue-600 hover:bg-blue-700 text-white px-8 py-2 rounded-full font-bold shadow-lg disabled:opacity-50">
            出牌
          </button>
          <button @click="handlePass" class="bg-red-600 hover:bg-red-700 text-white px-8 py-2 rounded-full font-bold shadow-lg">
            过
          </button>
        </div>

        <!-- 进贡/还贡 -->
        <div v-if="isTributePhase && amIPaying" class="flex gap-4">
          <div class="text-yellow-400 font-bold text-xl animate-pulse">
            {{ gameState?.phase === 'Tribute' ? '请进贡最大牌' : '请还贡一张牌' }}
          </div>
          <button @click="handleTributeAction" class="bg-purple-600 hover:bg-purple-700 text-white px-8 py-2 rounded-full font-bold shadow-lg">
            确认
          </button>
        </div>
      </div>

      <!-- 手牌 -->
      <HandArea :cards="myCards" :level="gameState?.level || 2" :view-mode="viewMode" :selected-ids="selectedCardIds"
        :highlighted-ids="highlightedIds" :straight-flush-ids="straightFlushIds" @toggle-select="toggleSelect" />
      <div class="text-white font-bold mt-2">{{ myPlayer?.name }} (Me)</div>
    </div>

    <!-- 分数结算 -->
    <div v-if="gameState?.phase === 'Score'"
      class="absolute inset-0 bg-black/80 flex flex-col items-center justify-center text-white z-50">
      <h1 class="text-6xl font-bold mb-8 text-yellow-400">本局结束</h1>
      <div class="text-2xl mb-4">
        获胜顺序: {{ gameState.winners.map(w => getPlayerName(w)).join(' → ') }}
      </div>
      <div v-if="gameState.teamLevels" class="text-xl text-gray-300 mb-4">
        当前等级 - 队伍0: {{ gameState.teamLevels[0] }} | 队伍1: {{ gameState.teamLevels[1] }}
      </div>
      <div class="text-lg text-yellow-300 animate-pulse">⏳ 3秒后自动开始下一局...</div>
      <div class="text-sm text-gray-400 mt-4">(对局将持续到某队打到A并连胜两次)</div>
    </div>

    <!-- 牌型选择弹窗 -->
    <div v-if="uiStore.showHandSelector" class="absolute inset-0 bg-black/80 flex items-center justify-center z-50">
      <div class="bg-[#252526] border border-[#333333] rounded-lg p-6 max-w-md shadow-2xl">
        <h2 class="text-2xl font-bold text-[#9cdcfe] mb-4">选择牌型</h2>
        <p class="text-gray-400 mb-4">您的牌包含红心{{ gameState?.level }}（万能牌），可以组成以下牌型：</p>
        <div class="flex flex-col gap-3">
          <button
            v-for="(hand, idx) in uiStore.possibleHands"
            :key="idx"
            @click="selectHandType(hand)"
            class="bg-[#3c3c3c] hover:bg-[#4c4c4c] text-white px-6 py-3 rounded-lg font-bold transition-colors text-left"
          >
            <div class="text-lg">{{ getHandDescription(hand) }}</div>
            <div class="text-sm text-gray-400 mt-1">{{ hand.type }} - 值: {{ hand.value }}</div>
          </button>
        </div>
        <button @click="uiStore.closeHandSelector()" class="mt-4 w-full bg-gray-600 hover:bg-gray-700 text-white px-4 py-2 rounded-lg">
          取消
        </button>
      </div>
    </div>

    <!-- 目标选择弹窗 -->
    <TargetSelector v-if="uiStore.showTargetSelector" :skill-type="uiStore.targetSelectorSkill?.type"
      :players="targetPlayers" :my-seat="mySeat" @select="handleTargetSelect" @cancel="uiStore.closeTargetSelector()" />

    <!-- 历史记录 -->
    <GameHistory v-if="uiStore.showHistory" :history="gameState?.history || []"
      :current-round="gameState?.currentRound || 1" @close="uiStore.showHistory = false" />
    <button
      v-if="gameState"
      @click="uiStore.showHistory = true"
      class="fixed top-4 right-[280px] z-30 bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg shadow-lg font-medium transition flex items-center gap-2"
    >
      <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
      历史记录
      <span v-if="gameState.history?.length" class="bg-red-500 text-xs px-2 py-0.5 rounded-full">
        {{ gameState.history.length }}
      </span>
    </button>

    <!-- 调试信息（开发时使用） -->
    <div v-if="gameState" class="absolute bottom-0 left-0 text-xs text-gray-500 p-2 pointer-events-none">
      [Debug] mySeat={{ mySeat }}, currentTurn={{ gameState.currentTurn }}, phase={{ gameState.phase }}, isMyTurn={{
      gameState.currentTurn === mySeat ? 'true' : 'false' }}
    </div>
  </div>
</template>

<script setup lang="ts">
  import { ref, computed, onMounted, watch, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { useRoomStore } from '@/store/room'
import { useGameStore } from '@/store/game'
import { useUIStore } from '@/store/ui'
import { useWebSocket } from '@/composables/useWebSocket'
import { useGameEvents } from '@/composables/useGameEvents'
import { GameMode, GamePhase, HandType, SkillCardType, type Card, type Hand, type Player, Suit, Rank } from '@/types'
import CardComponent from '@/components/Card.vue'
import PlayerArea from '@/components/PlayerArea.vue'
import ChatBox from '@/components/ChatBox.vue'
import HandArea from '@/components/HandArea.vue'
import SkillCardButton from '@/components/SkillCardButton.vue'
import TargetSelector from '@/components/TargetSelector.vue'
import GameHistory from '@/components/GameHistory.vue'
import { getLogicValue, getAllPossibleHandTypes, getHandDescription } from '@/utils/rules'

const router = useRouter()
const roomStore = useRoomStore()
const gameStore = useGameStore()
const uiStore = useUIStore()
const { sendMessage } = useWebSocket('/ws')

const viewMode = ref<'normal' | 'stacked'>('normal')
const selectedCardIds = ref<string[]>([])
const highlightedIds = ref<Set<string>>(new Set())
const straightFlushIds = ref<Set<string>>(new Set())

// 从 store 获取数据
const roomState = computed(() => roomStore.roomState)
const gameState = computed(() => gameStore.gameState)
const mySeat = computed(() => roomStore.mySeat)
const chatMessages = computed(() => roomStore.chatMessages)
const myPlayer = computed(() => roomState.value?.players.find(p => p?.seatIndex === mySeat.value) || null)

// 玩家辅助函数
function getPlayer(seat: number) {
  return roomState.value?.players.find(p => p?.seatIndex === seat) || null
}
function getPlayerName(seat: number) {
  return getPlayer(seat)?.name || `Seat ${seat}`
}
function getHandCount(seat: number): number {
  if (!gameState.value) return 0
  const h = gameState.value.hands[seat]
  return typeof h === 'number' ? h : h.length
}
function getAction(seat: number) {
  return gameState.value?.roundActions?.[seat]
}
function getWinnerLabel(seat: number): string | null {
  if (!gameState.value) return null
  const idx = gameState.value.winners.indexOf(seat)
  if (idx === -1) return null
  return ['头游', '二游', '三游', '末游'][idx]
}
function getChatBubble(seat: number): string | null {
  const msgs = chatMessages.value.filter(m => m.seatIndex === seat)
  if (msgs.length === 0) return null
  return msgs[msgs.length - 1].text
}

// 四个玩家信息
const topSeat = computed(() => (mySeat.value + 2) % 4)
const leftSeat = computed(() => (mySeat.value + 3) % 4)
const rightSeat = computed(() => (mySeat.value + 1) % 4)

const topPlayer = computed(() => getPlayer(topSeat.value))
const leftPlayer = computed(() => getPlayer(leftSeat.value))
const rightPlayer = computed(() => getPlayer(rightSeat.value))

const topHandCount = computed(() => getHandCount(topSeat.value))
const leftHandCount = computed(() => getHandCount(leftSeat.value))
const rightHandCount = computed(() => getHandCount(rightSeat.value))

const topAction = computed(() => getAction(topSeat.value))
const leftAction = computed(() => getAction(leftSeat.value))
const rightAction = computed(() => getAction(rightSeat.value))

const topWinnerLabel = computed(() => getWinnerLabel(topSeat.value))
const leftWinnerLabel = computed(() => getWinnerLabel(leftSeat.value))
const rightWinnerLabel = computed(() => getWinnerLabel(rightSeat.value))

const topChatBubble = computed(() => getChatBubble(topSeat.value))
const leftChatBubble = computed(() => getChatBubble(leftSeat.value))
const rightChatBubble = computed(() => getChatBubble(rightSeat.value))

const topIsTeammate = computed(() => (topSeat.value % 2) === (mySeat.value % 2) && topSeat.value !== mySeat.value)
const leftIsTeammate = computed(() => (leftSeat.value % 2) === (mySeat.value % 2) && leftSeat.value !== mySeat.value)
const rightIsTeammate = computed(() => (rightSeat.value % 2) === (mySeat.value % 2) && rightSeat.value !== mySeat.value)

const topIsOpponent = computed(() => !topIsTeammate.value && topSeat.value !== mySeat.value)
const leftIsOpponent = computed(() => !leftIsTeammate.value && leftSeat.value !== mySeat.value)
// 修复 rightIsOpponent 计算错误
const rightIsOpponent = computed(() => !rightIsTeammate.value && rightSeat.value !== mySeat.value)

// 我的牌
const myCards = computed(() => {
  if (!gameState.value) return []
  const h = gameState.value.hands[mySeat.value]
  return Array.isArray(h) ? h : []
})

// 高亮新牌
watch(() => gameState.value?.newCardIds, (ids) => {
  if (ids && ids.length) {
    const newSet = new Set(highlightedIds.value)
    ids.forEach(id => newSet.add(id))
    highlightedIds.value = newSet
    setTimeout(() => {
      const updated = new Set(highlightedIds.value)
      ids.forEach(id => updated.delete(id))
      highlightedIds.value = updated
    }, 3000)
  }
})

// 检测同花顺（用于视图切换）
watch(myCards, (cards) => {
  if (!gameState.value) return
  const level = gameState.value.level
  const suits = [Suit.Spades, Suit.Hearts, Suit.Clubs, Suit.Diamonds]
  const sfSet = new Set<string>()
  suits.forEach(suit => {
    const suitCards = cards.filter(c => c.suit === suit && !c.isWild && c.rank <= Rank.Ace)
    suitCards.sort((a, b) => a.rank - b.rank)
    let seq: Card[] = []
    for (const c of suitCards) {
      if (seq.length === 0) {
        seq.push(c)
      } else {
        const last = seq[seq.length - 1]
        if (c.rank === last.rank + 1) {
          seq.push(c)
        } else if (c.rank !== last.rank) {
          if (seq.length >= 5) seq.forEach(c => sfSet.add(c.id))
          seq = [c]
        }
      }
    }
    if (seq.length >= 5) seq.forEach(c => sfSet.add(c.id))
  })
  straightFlushIds.value = sfSet
}, { deep: true })

// 选中操作
function toggleSelect(id: string) {
  const idx = selectedCardIds.value.indexOf(id)
  if (idx >= 0) selectedCardIds.value.splice(idx, 1)
  else selectedCardIds.value.push(id)
}

// 出牌逻辑
function handlePlay() {
  const cards = myCards.value.filter(c => selectedCardIds.value.includes(c.id))
  if (cards.length === 0) return
  const hasWild = cards.some(c => c.isWild)
  if (hasWild && gameState.value) {
    const possible = getAllPossibleHandTypes(cards, gameState.value.level)
    if (possible.length > 1) {
      uiStore.openHandSelector(possible)
      return
    } else if (possible.length === 1) {
      sendMessage('playHand', { cards, handType: possible[0] })
      selectedCardIds.value = []
      return
    }
  }
  sendMessage('playHand', { cards })
  selectedCardIds.value = []
}

function selectHandType(hand: Hand) {
  const cards = myCards.value.filter(c => selectedCardIds.value.includes(c.id))
  sendMessage('playHand', { cards, handType: hand })
  selectedCardIds.value = []
  uiStore.closeHandSelector()
}

function handlePass() {
  sendMessage('pass', {})
}

function handleHint() {
  if (myCards.value.length) {
    selectedCardIds.value = [myCards.value[myCards.value.length - 1].id]
  }
}

// 进贡/还贡
const isTributePhase = computed(() => gameState.value && (gameState.value.phase === 'Tribute' || gameState.value.phase === 'ReturnTribute'))
const amIPaying = computed(() => {
  if (!isTributePhase.value || !gameState.value || !gameState.value.tributeState) return false
  const state = gameState.value.tributeState
  if (gameState.value.phase === 'Tribute') {
    return state.pendingTributes.some(t => t.from === mySeat.value)
  } else {
    return state.pendingReturns.some(r => r.from === mySeat.value)
  }
})

function handleTributeAction() {
  const cards = myCards.value.filter(c => selectedCardIds.value.includes(c.id))
  if (cards.length !== 1) {
    alert('请选择一张牌')
    return
  }
  if (gameState.value?.phase === 'Tribute') {
    sendMessage('tribute', { cards })
  } else {
    sendMessage('returnTribute', { cards })
  }
  selectedCardIds.value = []
}

// 技能卡
function handleSkillClick(skill: SkillCard) {
  const needsTarget = [SkillCardType.Steal, SkillCardType.Discard, SkillCardType.Skip]
  if (needsTarget.includes(skill.type)) {
    uiStore.openTargetSelector(skill)
  } else {
    sendMessage('useSkill', { skillId: skill.id })
  }
}

function handleTargetSelect(targetSeat: number) {
  const skill = uiStore.targetSelectorSkill
  if (skill) {
    sendMessage('useSkill', { skillId: skill.id, targetSeat })
    uiStore.closeTargetSelector()
  }
}

// 目标玩家列表
const targetPlayers = computed(() => {
  if (!roomState.value) return []
  return roomState.value.players
    .filter((p): p is Player => p !== null)
    .map(p => ({ name: p.name, seatIndex: p.seatIndex, handCount: getHandCount(p.seatIndex) }))
})

// 切换视图
function toggleViewMode() {
  viewMode.value = viewMode.value === 'normal' ? 'stacked' : 'normal'
}

// 换座
function handleSwitchSeat(seat: number) {
  if (gameState.value) return // 游戏中不能换
  sendMessage('switchSeat', { seatIndex: seat })
}

// 模式切换
function setGameMode(mode: GameMode) {
  sendMessage('setGameMode', { mode })
}

// 强制结束
function forceEnd() {
  if (confirm('⚠️ 确定要强制结束当前游戏吗？所有进度将丢失。')) {
    sendMessage('forceEndGame', {})
  }
}

// 聊天
function handleSendChat(text: string) {
  sendMessage('chatMessage', { text })
}

// 离开房间时重置
onMounted(() => {
  // 注册消息处理已由 useGameEvents 完成
})
</script>