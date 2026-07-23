import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useUIStore = defineStore('ui', () => {
  const error = ref<string | null>(null)
  const showTargetSelector = ref(false)
  const targetSelectorSkill = ref<any>(null)
  const showHistory = ref(false)
  const showHandSelector = ref(false)
  const possibleHands = ref<any[]>([])

  function showError(msg: string) {
    error.value = msg
    setTimeout(() => { error.value = null }, 3000)
  }

  function openTargetSelector(skill: any) {
    targetSelectorSkill.value = skill
    showTargetSelector.value = true
  }

  function closeTargetSelector() {
    showTargetSelector.value = false
    targetSelectorSkill.value = null
  }

  function openHandSelector(hands: any[]) {
    possibleHands.value = hands
    showHandSelector.value = true
  }

  function closeHandSelector() {
    showHandSelector.value = false
    possibleHands.value = []
  }

  function showMatchOver(payload: any) {
    // 可以使用 alert 或自定义弹窗
    alert(`🎉 对局结束！\n获胜队伍：${payload.winningTeam === 0 ? '0号和2号' : '1号和3号'}\n最终等级：${JSON.stringify(payload.finalLevels)}`)
  }

  return {
    error,
    showTargetSelector,
    targetSelectorSkill,
    showHistory,
    showHandSelector,
    possibleHands,
    showError,
    openTargetSelector,
    closeTargetSelector,
    openHandSelector,
    closeHandSelector,
    showMatchOver,
  }
})