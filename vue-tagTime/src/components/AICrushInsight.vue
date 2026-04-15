<template>
  <div class="ai-crush-container">
    <div class="ai-crush-header">
      <h3 class="ai-crush-title">AI 洞察</h3>
      <span v-if="remainingCount !== null" class="remaining-count">
        今日剩余: {{ remainingCount }}/3
      </span>
    </div>

    <button
      class="crush-button"
      :disabled="isLoading || remainingCount === 0"
      @click="getCrushLine"
    >
      <span v-if="isLoading" class="loading-text">
        <span class="loading-spinner"></span>
        思考中...
      </span>
      <span v-else>
        {{ remainingCount === 0 ? '今日次数已用完' : '用一句话击溃我' }}
      </span>
    </button>

    <transition name="fade">
      <div v-if="crushLine" class="crush-card">
        <div class="crush-content">
          <p class="crush-text">{{ displayedText }}</p>
        </div>
        <div class="crush-actions">
          <button
            class="action-button"
            :disabled="remainingCount === 0"
            @click="getCrushLine"
          >
            再来一次
          </button>
          <button class="action-button" @click="copyToClipboard">
            复制
          </button>
        </div>
      </div>
    </transition>

    <div v-if="error" class="error-message">
      {{ error }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { toast } from '../utils/message'
import axios from '../api/axios'

const isLoading = ref(false)
const crushLine = ref('')
const displayedText = ref('')
const remainingCount = ref<number | null>(null)
const error = ref('')

// 打字机效果
const typeWriter = (text: string) => {
  displayedText.value = ''
  let index = 0
  
  const interval = setInterval(() => {
    if (index < text.length) {
      displayedText.value += text[index]
      index++
    } else {
      clearInterval(interval)
    }
  }, 50)
}

// 监听 crushLine 变化，触发打字机效果
watch(crushLine, (newValue) => {
  if (newValue) {
    typeWriter(newValue)
  }
})

// 获取击溃语
const getCrushLine = async () => {
  isLoading.value = true
  error.value = ''
  crushLine.value = ''
  displayedText.value = ''

  try {
    const data: any = await axios.post('/ai/crush')
    console.log('API 返回数据:', data) // 调试日志
    console.log('crush_line:', data.crush_line) // 调试日志
    console.log('remaining_count:', data.remaining_count) // 调试日志
    
    if (!data.crush_line) {
      console.error('crush_line 为空或未定义')
      error.value = '未获取到击溃语，请重试'
      return
    }
    
    crushLine.value = data.crush_line
    remainingCount.value = data.remaining_count
  } catch (err: any) {
    console.error('获取击溃语失败:', err)
    console.error('错误响应:', err.response)
    error.value = err.response?.data?.error || err.response?.data?.message || '获取失败，请稍后再试'
  } finally {
    isLoading.value = false
  }
}

// 复制到剪贴板
const copyToClipboard = async () => {
  try {
    await navigator.clipboard.writeText(crushLine.value)
    toast.success('已复制到剪贴板')
  } catch (err) {
    toast.error('复制失败')
  }
}

// 初始化时获取剩余次数
const fetchRemainingCount = async () => {
  try {
    const data: any = await axios.get('/ai/crush/remaining')
    remainingCount.value = data.remaining_count
  } catch (err) {
    console.error('获取剩余次数失败:', err)
  }
}

// 组件挂载时获取剩余次数
onMounted(() => {
  fetchRemainingCount()
})
</script>

<style scoped>
.ai-crush-container {
  background: white;
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.ai-crush-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.ai-crush-title {
  font-size: 20px;
  font-weight: 600;
  color: #333;
  margin: 0;
}

.remaining-count {
  font-size: 14px;
  color: #666;
  background: #f5f5f5;
  padding: 4px 12px;
  border-radius: 12px;
}

.crush-button {
  width: 100%;
  padding: 16px;
  font-size: 16px;
  font-weight: 600;
  color: white;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.crush-button:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.crush-button:disabled {
  background: #ccc;
  cursor: not-allowed;
  transform: none;
}

.loading-text {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.loading-spinner {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.crush-card {
  margin-top: 24px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 12px;
  padding: 32px;
  box-shadow: 0 8px 24px rgba(102, 126, 234, 0.3);
}

.crush-content {
  margin-bottom: 24px;
}

.crush-text {
  font-size: 24px;
  line-height: 1.6;
  color: white;
  margin: 0;
  text-align: center;
  font-weight: 500;
  min-height: 40px;
}

.crush-actions {
  display: flex;
  gap: 12px;
  justify-content: center;
}

.action-button {
  padding: 10px 24px;
  font-size: 14px;
  color: #667eea;
  background: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.3s ease;
  font-weight: 500;
}

.action-button:hover:not(:disabled) {
  background: #f5f5f5;
  transform: translateY(-1px);
}

.action-button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.error-message {
  margin-top: 16px;
  padding: 12px;
  background: #fee;
  color: #c33;
  border-radius: 6px;
  text-align: center;
  font-size: 14px;
}

.fade-enter-active, .fade-leave-active {
  transition: opacity 0.5s ease;
}

.fade-enter-from, .fade-leave-to {
  opacity: 0;
}
</style>
