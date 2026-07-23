import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { GameState, HistoryEntry } from '@/types'

export const useGameStore = defineStore('game', () => {
  const gameState = ref<GameState | null>(null)
  const isGameOver = ref(false)

  function updateGameState(state: GameState) {
    gameState.value = state
    isGameOver.value = state.phase === 'Score'
  }

  function setGameOver(winners: number[]) {
    isGameOver.value = true
    if (gameState.value) {
      gameState.value.winners = winners
    }
  }

  function addHistoryEntry(entry: HistoryEntry) {
    if (gameState.value) {
      gameState.value.history.push(entry)
    }
  }

  function resetGame() {
    gameState.value = null
    isGameOver.value = false
  }

  return { gameState, isGameOver, updateGameState, setGameOver, addHistoryEntry, resetGame }
})