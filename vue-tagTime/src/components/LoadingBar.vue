<template>
  <transition name="loading-fade">
    <div v-if="isLoading" class="loading-bar">
      <div class="loading-progress" :style="{ width: progress + '%' }"></div>
    </div>
  </transition>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const isLoading = ref(false)
const progress = ref(0)
let timer: number | null = null

const startLoading = () => {
  isLoading.value = true
  progress.value = 0
  
  if (timer) clearInterval(timer)
  
  timer = window.setInterval(() => {
    if (progress.value < 90) {
      progress.value += Math.random() * 10
    }
  }, 200)
}

const finishLoading = () => {
  progress.value = 100
  
  if (timer) {
    clearInterval(timer)
    timer = null
  }
  
  setTimeout(() => {
    isLoading.value = false
    progress.value = 0
  }, 300)
}

router.beforeEach(() => {
  startLoading()
})

router.afterEach(() => {
  finishLoading()
})
</script>

<style scoped>
.loading-bar {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  z-index: 99999;
  background-color: transparent;
}

.loading-progress {
  height: 100%;
  background: linear-gradient(90deg, #667eea 0%, #764ba2 100%);
  transition: width 0.2s ease;
  box-shadow: 0 0 10px rgba(102, 126, 234, 0.5);
}

.loading-fade-enter-active,
.loading-fade-leave-active {
  transition: opacity 0.3s ease;
}

.loading-fade-enter-from,
.loading-fade-leave-to {
  opacity: 0;
}
</style>
