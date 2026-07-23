<template>
  <div class="bg-[#1e1e1e] min-h-screen text-gray-300">
    <!-- 错误提示 -->
    <div v-if="uiStore.error" class="fixed top-4 left-1/2 -translate-x-1/2 bg-red-500 text-white px-6 py-2 rounded-full shadow-lg z-50 font-bold animate-pulse">
      {{ uiStore.error }}
    </div>

    <!-- 路由视图 -->
    <router-view />
  </div>
</template>

<script setup lang="ts">
import { useUIStore } from './store'
const uiStore = useUIStore()
</script>