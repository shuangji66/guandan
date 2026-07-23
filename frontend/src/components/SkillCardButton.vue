<template>
  <button
    :class="[
      'w-16 h-20 rounded-lg bg-gradient-to-br border-2 flex flex-col items-center justify-center transition-all duration-200 shadow-lg',
      colorClass,
      disabled ? 'opacity-50 cursor-not-allowed' : 'hover:scale-110 hover:shadow-xl cursor-pointer active:scale-95',
    ]"
    :title="name"
    :disabled="disabled"
    @click="onClick"
  >
    <span class="text-2xl font-bold text-white drop-shadow-md">{{ icon }}</span>
    <span class="text-[10px] text-white/90 mt-1 font-medium">{{ name }}</span>
  </button>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { SkillCard, SkillCardType, SkillCardNames } from '@/types'

const props = defineProps<{
  skill: SkillCard
  disabled?: boolean
}>()

const emit = defineEmits<{
  (e: 'click'): void
}>()

const skillColors: Record<SkillCardType, string> = {
  [SkillCardType.DrawTwo]: 'from-green-600 to-green-800 border-green-400',
  [SkillCardType.Steal]: 'from-yellow-600 to-yellow-800 border-yellow-400',
  [SkillCardType.Discard]: 'from-red-600 to-red-800 border-red-400',
  [SkillCardType.Skip]: 'from-blue-600 to-blue-800 border-blue-400',
  [SkillCardType.Harvest]: 'from-amber-500 to-amber-700 border-amber-300',
}

const skillIcons: Record<SkillCardType, string> = {
  [SkillCardType.DrawTwo]: '+2',
  [SkillCardType.Steal]: '牵',
  [SkillCardType.Discard]: '拆',
  [SkillCardType.Skip]: '跳',
  [SkillCardType.Harvest]: '丰',
}

const colorClass = computed(() => skillColors[props.skill.type])
const icon = computed(() => skillIcons[props.skill.type])
const name = computed(() => SkillCardNames[props.skill.type])

const onClick = () => {
  emit('click')
}
</script>