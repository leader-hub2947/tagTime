<template>
  <div class="timer-page" :class="{ minimized: isMinimized }">
    <!-- 全屏计时页面 -->
    <div v-if="!isMinimized" class="fullscreen-timer">
      <div class="timer-header">
        <div class="header-spacer"></div>
        <button class="btn-icon" @click="minimizeTimer" title="最小化">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor">
            <line x1="5" y1="12" x2="19" y2="12" stroke-width="2"/>
          </svg>
        </button>
      </div>

      <div class="timer-content">
        <div class="task-info">
          <h1 class="task-name">{{ currentTask?.name }}</h1>
          <span class="task-tag" :style="{ backgroundColor: currentTask?.tag?.color }">
            {{ currentTask?.tag?.name }}
          </span>
        </div>

        <div class="timer-display">
          <div class="time-circle">
            <svg class="progress-ring" width="300" height="300">
              <circle
                class="progress-ring-bg"
                cx="150"
                cy="150"
                r="140"
              />
              <circle
                class="progress-ring-circle"
                cx="150"
                cy="150"
                r="140"
                :stroke-dasharray="circumference"
                :stroke-dashoffset="progressOffset"
              />
            </svg>
            <div class="time-text">
              <div class="time-main">{{ formattedTime }}</div>
              <div class="time-mode">{{ timerModeText }}</div>
            </div>
          </div>
        </div>

        <div class="timer-controls">
          <button 
            class="btn-control btn-pause" 
            @click="togglePause"
            :disabled="!currentEntry"
          >
            {{ isPaused ? '继续' : '暂停' }}
          </button>
          <button 
            class="btn-control btn-end" 
            @click="endTimer"
            :disabled="!currentEntry"
          >
            结束计时
          </button>
          <button 
            class="btn-control btn-complete" 
            @click="completeTask"
            :disabled="!currentEntry"
          >
            完成任务
          </button>
        </div>

        <div class="timer-mode-switch">
          <label>
            <input 
              type="radio" 
              value="free" 
              v-model="timerMode"
              :disabled="!!currentEntry"
            />
            自由计时
          </label>
          <label>
            <input 
              type="radio" 
              value="pomodoro" 
              v-model="timerMode"
              :disabled="!!currentEntry"
            />
            番茄钟 ({{ workMinutes }}分钟)
          </label>
        </div>

        <div v-if="timerMode === 'pomodoro'" class="pomodoro-info">
          <p>番茄钟: {{ pomodoroCount }} 个 | 休息时间: {{ breakCount }} 次</p>
          <p v-if="isBreakTime" class="break-notice">🎉 休息时间！({{ breakMinutes }}分钟)</p>
        </div>
      </div>
    </div>

    <!-- 悬浮计时条 -->
    <div v-else class="floating-timer">
      <div class="floating-content" @click="expandTimer">
        <div class="floating-task">
          <span class="floating-tag" :style="{ backgroundColor: currentTask?.tag?.color }">
            {{ currentTask?.tag?.name }}
          </span>
          <span class="floating-name">{{ currentTask?.name }}</span>
        </div>
        <div class="floating-time">{{ formattedTime }}</div>
      </div>
      <div class="floating-controls">
        <button class="btn-floating" @click.stop="togglePause">
          {{ isPaused ? '▶' : '⏸' }}
        </button>
        <button class="btn-floating" @click.stop="endTimer">⏹</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { timerAPI, taskAPI, type Task } from '../api/task'
import { toast, confirm } from '../utils/message'

const router = useRouter()
const route = useRoute()

const currentEntry = ref<any>(null)
const currentTask = ref<Task | null>(null)
const elapsedSeconds = ref(0)
const isPaused = ref(false)
const isMinimized = ref(false)
const timerMode = ref<'free' | 'pomodoro'>('free')
const pomodoroCount = ref(0)
const breakCount = ref(0)
const isBreakTime = ref(false)
const workMinutes = ref(25)
const breakMinutes = ref(5)

let intervalId: number | null = null

const circumference = 2 * Math.PI * 140

const progressOffset = computed(() => {
  if (timerMode.value === 'free') {
    return circumference
  }
  const targetTime = isBreakTime.value ? breakMinutes.value * 60 : workMinutes.value * 60
  const progress = elapsedSeconds.value / targetTime
  return circumference * (1 - Math.min(progress, 1))
})

const formattedTime = computed(() => {
  const hours = Math.floor(elapsedSeconds.value / 3600)
  const minutes = Math.floor((elapsedSeconds.value % 3600) / 60)
  const seconds = elapsedSeconds.value % 60
  
  if (hours > 0) {
    return `${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`
  }
  return `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`
})

const timerModeText = computed(() => {
  if (timerMode.value === 'free') return '自由计时'
  return isBreakTime.value ? '休息中' : `番茄钟 ${pomodoroCount.value + 1}`
})

const loadCurrentTimer = async () => {
  try {
    const res: any = await timerAPI.getCurrentTimer()
    if (res.entry) {
      currentEntry.value = res.entry
      currentTask.value = res.task
      isPaused.value = res.entry.is_paused
      timerMode.value = res.entry.timer_mode || 'free'
      pomodoroCount.value = res.entry.pomodoro_count || 0
      workMinutes.value = res.entry.work_minutes || 25
      breakMinutes.value = res.entry.break_minutes || 5
      
      if (!isPaused.value) {
        startInterval()
      }
      calculateElapsedTime()
    } else {
      router.push('/tasks')
    }
  } catch (err) {
    console.error('加载计时失败', err)
    router.push('/tasks')
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
      
      // 番茄钟模式检查
      if (timerMode.value === 'pomodoro') {
        const targetTime = isBreakTime.value ? breakMinutes.value * 60 : workMinutes.value * 60
        if (elapsedSeconds.value >= targetTime) {
          handlePomodoroComplete()
        }
      }
    }
  }, 1000)
}

const stopInterval = () => {
  if (intervalId) {
    clearInterval(intervalId)
    intervalId = null
  }
}

const handlePomodoroComplete = async () => {
  if (isBreakTime.value) {
    // 休息结束，开始新的番茄钟
    isBreakTime.value = false
    elapsedSeconds.value = 0
    breakCount.value++
    toast.info('休息结束！开始新的番茄钟', 5000)
  } else {
    // 工作时间结束，进入休息
    pomodoroCount.value++
    isBreakTime.value = true
    elapsedSeconds.value = 0
    toast.success(`完成第 ${pomodoroCount.value} 个番茄钟！休息 ${breakMinutes.value} 分钟`, 5000)
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
    }
    await loadCurrentTimer()
  } catch (err: any) {
    toast.error(err.response?.data?.error || '操作失败')
  }
}

const endTimer = async () => {
  if (!confirm('确定结束计时吗？')) return
  
  try {
    stopInterval()
    await timerAPI.endTimer()
    router.push('/tasks')
  } catch (err: any) {
    toast.error(err.response?.data?.error || '结束计时失败')
  }
}

const completeTask = async () => {
  if (!confirm('确定完成任务吗？这将结束计时并标记任务为已完成。')) return
  
  try {
    stopInterval()
    await taskAPI.completeTask(currentTask.value!.id)
    router.push('/tasks')
  } catch (err: any) {
    toast.error(err.response?.data?.error || '完成任务失败')
  }
}

const minimizeTimer = () => {
  isMinimized.value = true
  router.push('/tasks')
}

const expandTimer = () => {
  isMinimized.value = false
}

onMounted(() => {
  loadCurrentTimer()
})

onUnmounted(() => {
  stopInterval()
})
</script>

<style scoped>
.timer-page {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 9999;
}

.fullscreen-timer {
  width: 100%;
  height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  flex-direction: column;
  color: white;
}

.timer-header {
  padding: 20px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-spacer {
  width: 48px;
}

.btn-icon {
  background: rgba(255, 255, 255, 0.2);
  border: none;
  border-radius: 50%;
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: white;
  transition: all 0.3s;
}

.btn-icon:hover {
  background: rgba(255, 255, 255, 0.3);
  transform: scale(1.1);
}

.timer-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px;
}

.task-info {
  text-align: center;
  margin-bottom: 60px;
}

.task-name {
  font-size: 36px;
  font-weight: 600;
  margin-bottom: 16px;
}

.task-tag {
  padding: 8px 20px;
  border-radius: 20px;
  font-size: 16px;
  color: white;
}

.timer-display {
  margin-bottom: 60px;
}

.time-circle {
  position: relative;
  width: 300px;
  height: 300px;
}

.progress-ring {
  transform: rotate(-90deg);
}

.progress-ring-bg {
  fill: none;
  stroke: rgba(255, 255, 255, 0.2);
  stroke-width: 8;
}

.progress-ring-circle {
  fill: none;
  stroke: white;
  stroke-width: 8;
  stroke-linecap: round;
  transition: stroke-dashoffset 0.3s;
}

.time-text {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  text-align: center;
}

.time-main {
  font-size: 56px;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
}

.time-mode {
  font-size: 18px;
  opacity: 0.9;
  margin-top: 8px;
}

.timer-controls {
  display: flex;
  gap: 20px;
  margin-bottom: 40px;
}

.btn-control {
  padding: 16px 40px;
  border: 2px solid white;
  border-radius: 30px;
  background: transparent;
  color: white;
  font-size: 18px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s;
}

.btn-control:hover:not(:disabled) {
  background: white;
  color: #667eea;
  transform: translateY(-2px);
}

.btn-control:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-pause:hover:not(:disabled) {
  color: #f59e0b;
}

.btn-end:hover:not(:disabled) {
  color: #ef4444;
}

.btn-complete:hover:not(:disabled) {
  color: #10b981;
}

.timer-mode-switch {
  display: flex;
  gap: 30px;
  margin-bottom: 20px;
}

.timer-mode-switch label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  cursor: pointer;
}

.timer-mode-switch input[type="radio"] {
  width: 20px;
  height: 20px;
  cursor: pointer;
}

.pomodoro-info {
  text-align: center;
  font-size: 16px;
  opacity: 0.9;
}

.pomodoro-info p {
  margin: 8px 0;
}

.break-notice {
  font-size: 20px;
  font-weight: 600;
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.7; }
}

/* 悬浮计时条 */
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
  z-index: 9999;
  cursor: pointer;
  transition: all 0.3s;
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
}
</style>
