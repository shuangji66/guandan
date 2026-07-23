import { createRouter, createWebHistory } from 'vue-router'
import Lobby from '@/views/Lobby.vue'
import GameTable from '@/views/GameTable.vue'

const routes = [
  { path: '/', component: Lobby },
  { path: '/game', component: GameTable, meta: { requiresRoom: true } },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to, from, next) => {
  // 简单检查是否有房间状态，实际可从 store 获取
  next()
})

export default router