<template>
  <div class="ai-crush-page">
    <AppHeader />
    <div class="page-container">
      <div class="page-header">
        <h1 class="page-title">AI 洞悉</h1>
        <p class="page-description">让 AI 用一句话击溃你的拖延症</p>
      </div>
      <div class="content-wrapper">
        <AICrushInsight />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import AppHeader from '../components/AppHeader.vue'
import AICrushInsight from '../components/AICrushInsight.vue'
</script>

<style scoped>
.ai-crush-page {
  min-height: 100vh;
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
}

.page-container {
  max-width: 800px;
  margin: 0 auto;
  padding: 40px 20px;
}

.page-header {
  text-align: center;
  margin-bottom: 40px;
}

.page-title {
  font-size: 36px;
  font-weight: 700;
  color: #333;
  margin: 0 0 12px 0;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.page-description {
  font-size: 16px;
  color: #666;
  margin: 0;
}

.content-wrapper {
  animation: fadeInUp 0.6s ease;
}

@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
