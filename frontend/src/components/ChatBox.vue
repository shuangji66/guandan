<template>
  <div class="absolute top-4 right-4 w-72 h-56 bg-[#252526] border border-[#333333] rounded flex flex-col pointer-events-auto z-10 shadow-lg">
    <!-- 消息列表 -->
    <div ref="chatEnd" class="flex-1 overflow-y-auto p-2 text-sm text-[#d4d4d4] scrollbar-thin">
      <div v-for="(msg, idx) in messages" :key="idx" class="mb-1">
        <span class="text-[#858585] text-xs">[{{ msg.time }}] </span>
        <span class="font-bold text-[#569cd6]">{{ msg.sender }}: </span>
        <span class="break-words">{{ msg.text }}</span>
      </div>
      <div ref="chatEndRef" />
    </div>

    <!-- 表情选择 -->
    <div v-if="showEmojiPicker" class="p-2 border-t border-[#333333] bg-[#1e1e1e] grid grid-cols-10 gap-1">
      <button
        v-for="(emoji, idx) in quickEmojis"
        :key="idx"
        type="button"
        @click="insertEmoji(emoji)"
        class="text-lg hover:bg-[#3c3c3c] rounded p-1 transition-colors"
      >
        {{ emoji }}
      </button>
    </div>

    <!-- 输入 -->
    <form @submit.prevent="send" class="p-2 border-t border-[#333333] flex items-center gap-1">
      <button
        type="button"
        @click="toggleEmojiPicker"
        class="text-lg hover:bg-[#3c3c3c] rounded p-1"
        title="表情"
      >
        😊
      </button>
      <input
        v-model="inputText"
        class="flex-1 bg-[#3c3c3c] border-none text-white text-sm focus:outline-none rounded px-2 py-1"
        placeholder="输入消息..."
        @keyup.enter="send"
      />
      <button type="submit" class="text-[#0e639c] font-bold text-sm hover:text-[#1177bb]">发送</button>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'
import type { ChatMessage } from '@/types'

const props = defineProps<{
  messages: ChatMessage[]
}>()

const emit = defineEmits<{
  (e: 'send', text: string): void
}>()

const inputText = ref('')
const showEmojiPicker = ref(false)
const chatEndRef = ref<HTMLElement | null>(null)

const quickEmojis = ['😀', '😂', '🤣', '😎', '🥳', '😭', '😡', '🤔', '👍', '👎', '❤️', '🔥', '💯', '🎉', '🤝', '✌️', '💪', '🙏', '😱', '🤯']

const send = () => {
  if (inputText.value.trim()) {
    emit('send', inputText.value.trim())
    inputText.value = ''
  }
}

const insertEmoji = (emoji: string) => {
  inputText.value += emoji
  showEmojiPicker.value = false
}

const toggleEmojiPicker = () => {
  showEmojiPicker.value = !showEmojiPicker.value
}

// 自动滚动到底部
watch(() => props.messages.length, () => {
  nextTick(() => {
    chatEndRef.value?.scrollIntoView({ behavior: 'smooth' })
  })
})
</script>