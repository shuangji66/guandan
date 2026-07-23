import { createPinia } from 'pinia'
export const pinia = createPinia()
export { useRoomStore } from './room'
export { useGameStore } from './game'
export { useUIStore } from './ui'