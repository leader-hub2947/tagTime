<template>
  <div class="container">
    <AppHeader />

    <div class="page-header">
      <h1 class="page-title">数据统计</h1>
    </div>

    <!-- 加载状态 -->
    <div v-if="isLoading" class="dashboard-grid">
      <SkeletonLoader type="dashboard" />
      <SkeletonLoader type="dashboard" />
      <div class="dashboard-card full-width">
        <div class="skeleton-title"></div>
        <div class="skeleton-line"></div>
      </div>
    </div>

    <!-- 数据内容 -->
    <div v-else class="dashboard-grid">
      <!-- AI 洞察 -->
      <div class="dashboard-card full-width">
        <AICrushInsight />
      </div>

      <!-- 任务完成度统计 -->
      <div class="dashboard-card">
        <h3>任务完成度</h3>
        <div class="period-tabs">
          <button 
            v-for="p in ['week', 'month', 'year']" 
            :key="p"
            :class="{ active: period === p }"
            @click="period = p; loadStatistics()"
          >
            {{ { week: '本周', month: '本月', year: '本年' }[p] }}
          </button>
        </div>
        <div class="stats-display">
          <div class="stat-item">
            <div class="stat-value">{{ statistics.total_tasks }}</div>
            <div class="stat-label">总任务数</div>
          </div>
          <div class="stat-item">
            <div class="stat-value">{{ statistics.completed_tasks }}</div>
            <div class="stat-label">已完成</div>
          </div>
          <div class="stat-item">
            <div class="stat-value">{{ statistics.completion_rate?.toFixed(1) }}%</div>
            <div class="stat-label">完成率</div>
          </div>
        </div>
      </div>

      <!-- 标签时长排行 -->
      <div class="dashboard-card">
        <h3>标签时长排行</h3>
        <div class="ranking-list">
          <div v-if="rankings.length === 0" class="empty-hint">暂无数据</div>
          <div v-for="(rank, index) in rankings" :key="rank.tag_id" class="ranking-item">
            <div class="rank-number">{{ index + 1 }}</div>
            <div class="rank-tag" :style="{ backgroundColor: rank.tag_color }">
              {{ rank.tag_name }}
            </div>
            <div class="rank-duration">{{ formatDuration(rank.total_duration) }}</div>
          </div>
        </div>
      </div>

      <!-- 今日时间轴 -->
      <div class="dashboard-card full-width">
        <h3>今日时间轴</h3>
        <div class="timeline-container">
          <div v-if="timeline.length === 0" class="empty-hint">今日暂无计时记录</div>
          <div v-else class="timeline-chart">
            <div class="timeline-hours">
              <span v-for="h in 24" :key="h">{{ h - 1 }}</span>
            </div>
            <div class="timeline-bars">
              <div 
                v-for="entry in timeline" 
                :key="entry.task_id"
                class="timeline-bar"
                :style="getTimelineStyle(entry)"
                :title="`${entry.task_name} - ${formatDuration(entry.duration)}`"
              >
                <span>{{ entry.task_name }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { dashboardAPI } from '../api/task'
import AppHeader from '../components/AppHeader.vue'
import SkeletonLoader from '../components/SkeletonLoader.vue'
import AICrushInsight from '../components/AICrushInsight.vue'

const period = ref('month')
const isLoading = ref(true)
const statistics = ref({
  total_tasks: 0,
  completed_tasks: 0,
  completion_rate: 0
})
const rankings = ref<any[]>([])
const timeline = ref<any[]>([])

const formatDuration = (seconds: number) => {
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  return `${hours}小时${minutes}分钟`
}

const getTimelineStyle = (entry: any) => {
  const start = new Date(entry.start_time)
  const end = entry.end_time ? new Date(entry.end_time) : new Date()
  
  const startHour = start.getHours() + start.getMinutes() / 60
  const endHour = end.getHours() + end.getMinutes() / 60
  
  const left = (startHour / 24) * 100
  const width = ((endHour - startHour) / 24) * 100
  
  return {
    left: `${left}%`,
    width: `${width}%`,
    backgroundColor: entry.tag_color
  }
}

const loadStatistics = async () => {
  try {
    const res: any = await dashboardAPI.getTaskStatistics(period.value)
    statistics.value = res
  } catch (err) {
    console.error('加载统计数据失败', err)
  }
}

const loadRankings = async () => {
  try {
    const res: any = await dashboardAPI.getTagRanking()
    rankings.value = res.rankings || []
  } catch (err) {
    console.error('加载排行数据失败', err)
  }
}

const loadTimeline = async () => {
  try {
    const res: any = await dashboardAPI.getTimeline()
    timeline.value = res.timeline || []
  } catch (err) {
    console.error('加载时间轴数据失败', err)
  }
}

const loadAllData = async () => {
  isLoading.value = true
  await Promise.all([
    loadStatistics(),
    loadRankings(),
    loadTimeline()
  ])
  isLoading.value = false
}

onMounted(() => {
  loadAllData()
})
</script>

<style scoped>
.skeleton-title,
.skeleton-line {
  background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
  border-radius: 4px;
  margin-bottom: 12px;
}

.skeleton-title {
  height: 24px;
  width: 40%;
}

.skeleton-line {
  height: 60px;
  width: 100%;
}

@keyframes shimmer {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

.dashboard-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20px;
}

.dashboard-card {
  background-color: #fafafa;
  border: 1px solid #e5e5e5;
  border-radius: 8px;
  padding: 25px;
  transition: all 0.3s ease;
}

.dashboard-card:hover {
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.08);
  transform: translateY(-2px);
}

.dashboard-card.full-width {
  grid-column: 1 / -1;
}

.dashboard-card h3 {
  font-size: 18px;
  margin-bottom: 20px;
}

.period-tabs {
  display: flex;
  gap: 10px;
  margin-bottom: 20px;
}

.period-tabs button {
  padding: 8px 16px;
  border: 1px solid #e5e5e5;
  border-radius: 6px;
  background: white;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.3s ease;
}

.period-tabs button:hover {
  border-color: #333;
  transform: translateY(-2px);
}

.period-tabs button.active {
  background-color: #333;
  color: white;
  border-color: #333;
}

.stats-display {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20px;
}

.stat-item {
  text-align: center;
  transition: all 0.3s ease;
}

.stat-item:hover {
  transform: scale(1.05);
}

.stat-value {
  font-size: 32px;
  font-weight: 600;
  color: #333;
  margin-bottom: 8px;
}

.stat-label {
  font-size: 14px;
  color: #666;
}

.ranking-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.ranking-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: white;
  border-radius: 6px;
  transition: all 0.3s ease;
}

.ranking-item:hover {
  transform: translateX(5px);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.rank-number {
  width: 30px;
  height: 30px;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #333;
  color: white;
  border-radius: 50%;
  font-weight: 600;
}

.rank-tag {
  flex: 1;
  padding: 6px 12px;
  border-radius: 6px;
  color: white;
  font-weight: 500;
}

.rank-duration {
  font-size: 14px;
  color: #666;
}

.timeline-container {
  min-height: 200px;
}

.timeline-chart {
  position: relative;
}

.timeline-hours {
  display: grid;
  grid-template-columns: repeat(24, 1fr);
  gap: 2px;
  margin-bottom: 10px;
  font-size: 12px;
  color: #999;
}

.timeline-bars {
  position: relative;
  height: 60px;
  background: white;
  border-radius: 6px;
}

.timeline-bar {
  position: absolute;
  height: 100%;
  border-radius: 4px;
  display: flex;
  align-items: center;
  padding: 0 8px;
  color: white;
  font-size: 12px;
  overflow: hidden;
  white-space: nowrap;
  transition: all 0.3s ease;
}

.timeline-bar:hover {
  transform: scaleY(1.1);
  z-index: 10;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}

.empty-hint {
  text-align: center;
  color: #999;
  padding: 40px 0;
}

@media (max-width: 968px) {
  .dashboard-grid {
    grid-template-columns: 1fr;
  }
}
</style>
