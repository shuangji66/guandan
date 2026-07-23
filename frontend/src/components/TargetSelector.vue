<template>
  <div class="fixed inset-0 bg-black/70 flex items-center justify-center z-50">
    <div class="bg-[#252526] border border-[#3c3c3c] rounded-xl p-6 shadow-2xl min-w-[300px]">
      <h2 class="text-xl font-bold text-white mb-4 text-center">
        选择【{{ skillName }}】的目标
      </h2>
      <div class="flex flex-col gap-3">
        <template v-if="validTargets.length === 0">
          <p class="text-gray-400 text-center">没有可选择的目标</p>
        </template>
        <template v-else>
          <button
            v-for="player in validTargets"
            :key="player.seatIndex"
            :class="[
              'px-4 py-3 rounded-lg text-white font-medium transition-all flex justify-between items-center',
              player.seatIndex % 2 === mySeat % 2 ? 'bg-blue-700 hover:bg-blue-600' : 'bg-red-700 hover:bg-red-600',
            ]"
            @click="selectTarget(player.seatIndex)"
          >
            <span>{{ player.name }}</span>
            <span class="text-sm opacity-75">
              {{ player.seatIndex % 2 === mySeat % 2 ? '队友' : '对手' }} · {{ player.handCount }}张牌
            </span>
          </button>
        </template>
      </div>
      <button
        @click="cancel"
        class="mt-4 w-full px-4 py-2 bg-gray-600 hover:bg-gray-500 rounded-lg text-white font-medium transition-all"
      >
        取消
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { SkillCardType, SkillCardNames } from '@/types'

interface Player {
  name: string
  seatIndex: number
  handCount: number
}

const props = defineProps<{
  skillType: SkillCardType
  players: Player[]
  mySeat: number
}>()

const emit = defineEmits<{
  (e: 'select', targetSeat: number): void
  (e: 'cancel'): void
}>()

const validTargets = computed(() =>
  props.players.filter(p => p.seatIndex !== props.mySeat && p.handCount > 0)
)

const skillName = computed(() => SkillCardNames[props.skillType])

const selectTarget = (seat: number) => {
  emit('select', seat)
}

const cancel = () => {
  emit('cancel')
}
</script>