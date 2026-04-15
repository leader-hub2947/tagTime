<template>
  <div v-if="currentEntry && !isTimerPage" class="floating-timer" @click="goToTimer">
    <div class="floating-content">
      <div class="floating-task">
        <span class="floating-tag" :style="{ backgroundColor: currentTask?.tag?.color }">
          {{ currentTask?.tag?.name }}
        </span>
        <span class="floating-name">{{ currentTask?.name }}</span>
      </div>
      <div class="floating-time">{{ formattedTime }}</div>
    </div>
    <div class="floating-controls">
      <button class="btn-floating" @click.stop="togglePause" :title="isPaused ? '继续' : '暂停'">
        {{ isPaused ? '▶' : '⏸' }}
      </button>
      <button class="btn-floating" @click.stop="endTimer" title="结束">⏹</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { timerAPI, type Task } from '../api/task'

const router = useRouter()
const route = useRoute()

const currentEntry = ref<any>(null)
const currentTask = ref<Task | null>(null)
const elapsedSeconds = ref(0)
const isPaused = ref(false)

let intervalId: number | null = null
let checkIntervalId: number | null = null

const isTimerPage = computed(() => route.path === '/timer')

const formattedTime = computed(() => {
  const hours = Math.floor(elapsedSeconds.value / 3600)
  const minutes = Math.floor((elapsedSeconds.value % 3600) / 60)
  const seconds = elapsedSeconds.value % 60
  
  if (hours > 0) {
    return `${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`
  }
  return `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`
})

const loadCurrentTimer = async () => {
  // 检查是否已登录
  const token = localStorage.getItem('token')
  if (!token) {
    currentEntry.value = null
    currentTask.value = null
    stopInterval()
    return
  }
  
  try {
    const res: any = await timerAPI.getCurrentTimer()
    if (res.entry) {
      currentEntry.value = res.entry
      currentTask.value = res.task
      isPaused.value = res.entry.is_paused
      
      if (!isPaused.value && !intervalId) {
        startInterval()
      } else if (isPaused.value && intervalId) {
        stopInterval()
      }
      
      calculateElapsedTime()
    } else {
      currentEntry.value = null
      currentTask.value = null
      stopInterval()
    }
  } catch (err) {
    console.error('加载计时失败', err)
  }
}

const calculateElapsedTime = () => {
  if (!currentEntry.value) return
  
  const startTime = new Date(currentEntry.value.start_time).getTime()
  const now = Date.now()
  const pausedDuration = (currentEntry.value.paused_duration || 0) * 1000
  
  if (isPaused.value && currentEntry.value.last_pause_time) {
    const lastPauseTime = new Date(currentEntry.value.last_pause_time).getTime()
    elapsedSeconds.value = Math.floor((lastPauseTime - startTime - pausedDuration) / 1000)
  } else {
    elapsedSeconds.value = Math.floor((now - startTime - pausedDuration) / 1000)
  }
}

const startInterval = () => {
  if (intervalId) return
  
  intervalId = window.setInterval(() => {
    if (!isPaused.value) {
      elapsedSeconds.value++
    }
  }, 1000)
}

const stopInterval = () => {
  if (intervalId) {
    clearInterval(intervalId)
    intervalId = null
  }
}

const togglePause = async () => {
  try {
    if (isPaused.value) {
      await timerAPI.resumeTimer()
      isPaused.value = false
      startInterval()
    } else {
      await timerAPI.pauseTimer()
      isPaused.value = true
      stopInterval()
    }
    await loadCurrentTimer()
  } catch (err: any) {
    console.error('操作失败', err)
  }
}

const endTimer = async () => {
  if (!confirm('确定结束计时吗？')) return
  
  try {
    stopInterval()
    await timerAPI.endTimer()
    currentEntry.value = null
    currentTask.value = null
  } catch (err: any) {
    console.error('结束计时失败', err)
  }
}

const goToTimer = () => {
  router.push('/timer')
}

onMounted(() => {
  loadCurrentTimer()
  // 每10秒检查一次计时状态
  checkIntervalId = window.setInterval(loadCurrentTimer, 10000)
})

onUnmounted(() => {
  stopInterval()
  if (checkIntervalId) {
    clearInterval(checkIntervalId)
  }
})

// 监听路由变化
watch(() => route.path, () => {
  // 如果是登录页面，清理计时器
  if (route.path === '/login') {
    currentEntry.value = null
    currentTask.value = null
    stopInterval()
    if (checkIntervalId) {
      clearInterval(checkIntervalId)
    }
  } else if (route.path !== '/timer') {
    loadCurrentTimer()
  }
})
</script>

<style scoped>
.floating-timer {
  position: fixed;
  bottom: 20px;
  right: 20px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  color: white;
  z-index: 9998;
  cursor: pointer;
  transition: all 0.3s;
  animation: slideIn 0.3s ease-out;
}

@keyframes slideIn {
  from {
    transform: translateY(100px);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}

.floating-timer:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 40px rgba(0, 0, 0, 0.4);
}

.floating-content {
  display: flex;
  align-items: center;
  gap: 12px;
}

.floating-task {
  display: flex;
  align-items: center;
  gap: 8px;
}

.floating-tag {
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
  color: white;
  white-space: nowrap;
}

.floating-name {
  font-weight: 500;
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.floating-time {
  font-size: 20px;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
  min-width: 80px;
}

.floating-controls {
  display: flex;
  gap: 8px;
  border-left: 1px solid rgba(255, 255, 255, 0.3);
  padding-left: 12px;
}

.btn-floating {
  background: rgba(255, 255, 255, 0.2);
  border: none;
  border-radius: 6px;
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: white;
  font-size: 16px;
  transition: all 0.2s;
}

.btn-floating:hover {
  background: rgba(255, 255, 255, 0.3);
  transform: scale(1.1);
}
</style>
