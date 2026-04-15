<template>
  <div class="skeleton-wrapper">
    <div v-if="type === 'card'" class="skeleton-card">
      <div class="skeleton-title"></div>
      <div class="skeleton-line"></div>
      <div class="skeleton-line short"></div>
    </div>
    
    <div v-else-if="type === 'dashboard'" class="skeleton-dashboard">
      <div class="skeleton-card">
        <div class="skeleton-title"></div>
        <div class="skeleton-stats">
          <div class="skeleton-stat"></div>
          <div class="skeleton-stat"></div>
          <div class="skeleton-stat"></div>
        </div>
      </div>
    </div>
    
    <div v-else-if="type === 'list'" class="skeleton-list">
      <div v-for="i in count" :key="i" class="skeleton-item">
        <div class="skeleton-circle"></div>
        <div class="skeleton-content">
          <div class="skeleton-line"></div>
          <div class="skeleton-line short"></div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  type?: 'card' | 'dashboard' | 'list'
  count?: number
}>()
</script>

<style scoped>
.skeleton-wrapper {
  animation: fadeIn 0.3s ease;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.skeleton-card,
.skeleton-item {
  background-color: #fafafa;
  border: 1px solid #e5e5e5;
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 15px;
}

.skeleton-title,
.skeleton-line,
.skeleton-circle,
.skeleton-stat {
  background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
  border-radius: 4px;
}

@keyframes shimmer {
  0% {
    background-position: 200% 0;
  }
  100% {
    background-position: -200% 0;
  }
}

.skeleton-title {
  height: 24px;
  width: 40%;
  margin-bottom: 16px;
}

.skeleton-line {
  height: 16px;
  width: 100%;
  margin-bottom: 12px;
}

.skeleton-line.short {
  width: 60%;
}

.skeleton-stats {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20px;
  margin-top: 20px;
}

.skeleton-stat {
  height: 80px;
  border-radius: 8px;
}

.skeleton-list {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.skeleton-item {
  display: flex;
  align-items: center;
  gap: 15px;
}

.skeleton-circle {
  width: 50px;
  height: 50px;
  border-radius: 50%;
  flex-shrink: 0;
}

.skeleton-content {
  flex: 1;
}
</style>
