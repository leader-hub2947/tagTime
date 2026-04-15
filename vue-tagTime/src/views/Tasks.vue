<template>
  <div class="container">
    <AppHeader />

    <div class="page-header">
      <h1 class="page-title">任务管理</h1>
      <div class="header-actions">
        <button class="btn btn-icon" @click="showSettingsModal = true" title="自动归档设置">
          ⚙️
        </button>
        <button class="btn btn-primary" @click="showCreateModal = true">+ 新建任务</button>
      </div>
    </div>

    <div class="filter-bar">
      <div class="filter-left">
        <div 
          class="tag-filter"
          :class="{ active: !showArchived && selectedStatus === null }"
          @click="selectStatus(null)"
        >
          我的任务
        </div>
        <div 
          class="tag-filter"
          :class="{ active: !showArchived && selectedStatus === 2 }"
          @click="selectStatus(2)"
        >
          已完成
        </div>
      </div>
      <div class="filter-right">
        <div 
          class="tag-filter archive-filter"
          :class="{ active: showArchived }"
          @click="toggleArchived"
        >
          📦 归档
        </div>
      </div>
    </div>

    <div class="tasks-list">
      <div v-if="!showArchived">
        <div v-if="filteredTasks.length === 0" class="empty-state">
          <div class="empty-state-icon">✅</div>
          <div class="empty-state-text">
            {{ selectedStatus === 0 ? '暂无未开始的任务' : selectedStatus === 1 ? '暂无进行中的任务' : '暂无已完成的任务' }}
          </div>
        </div>

        <div v-for="(tasks, date) in groupTasksByDate(filteredTasks)" :key="date" class="task-group">
          <div class="task-date-header">
            <span class="date-icon">📅</span>
            <span class="date-text">{{ formatArchiveDate(date) }}</span>
            <span class="task-count">{{ tasks.length }} 个任务</span>
          </div>
          
          <div 
            v-for="task in tasks" 
            :key="task.id" 
            class="task-card"
            :class="{ 'task-completing': completingTaskIds.has(task.id) }"
          >
            <div class="task-header">
              <div>
                <span class="task-status" :class="'status-' + task.status">
                  {{ ['未开始', '进行中', '已完成'][task.status] }}
                </span>
                <span class="task-date">创建于 {{ formatDate(task.created_at) }}</span>
              </div>
              <div class="task-actions">
                <button v-if="task.status !== 2" @click="startTask(task)">开始计时</button>
                <button v-if="task.status === 1" @click="showCompleteModal(task)" class="btn-complete">
                  <span class="check-icon">✓</span>
                </button>
                <button v-if="task.status === 2" @click="archiveTask(task.id)" class="btn-archive">归档</button>
                <button @click="editTask(task)">编辑</button>
                <button @click="deleteTask(task.id)">删除</button>
              </div>
            </div>
            <div class="task-content">
              <h3>{{ task.name }}</h3>
              <p v-if="task.description">{{ task.description }}</p>
            </div>
            <div class="task-footer">
              <span v-if="task.tag" class="task-tag" :style="{ backgroundColor: task.tag.color }">
                {{ task.tag.name }}
              </span>
              <span class="task-duration">累计时长: {{ formatDuration(task.total_duration) }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 归档任务视图 -->
      <div v-else class="archived-tasks">
        <div v-if="Object.keys(groupedArchivedTasks).length === 0" class="empty-state">
          <div class="empty-state-icon">📦</div>
          <div class="empty-state-text">暂无归档任务</div>
        </div>

        <div v-for="(tasks, date) in groupedArchivedTasks" :key="date" class="archived-group">
          <div class="archived-date-header">
            <span class="date-icon">📅</span>
            <span class="date-text">{{ formatArchiveDate(date) }}</span>
            <span class="task-count">{{ tasks.length }} 个任务</span>
          </div>
          
          <div v-for="task in tasks" :key="task.id" class="task-card archived-task-card">
            <div class="task-header">
              <div>
                <span class="task-status status-archived">已归档</span>
                <span class="task-date">创建于 {{ formatDate(task.created_at) }}</span>
              </div>
              <div class="task-actions">
                <button @click="unarchiveTask(task.id)" class="btn-unarchive">取消归档</button>
                <button @click="deleteTask(task.id)">删除</button>
              </div>
            </div>
            <div class="task-content">
              <h3>{{ task.name }}</h3>
              <p v-if="task.description">{{ task.description }}</p>
            </div>
            <div class="task-footer">
              <span v-if="task.tag" class="task-tag" :style="{ backgroundColor: task.tag.color }">
                {{ task.tag.name }}
              </span>
              <span class="task-duration">累计时长: {{ formatDuration(task.total_duration) }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 创建/编辑任务弹窗 -->
    <div v-if="showCreateModal" class="modal" @click.self="showCreateModal = false">
      <div class="modal-content">
        <h2>{{ editingTask ? '编辑任务' : '新建任务' }}</h2>
        <div class="form-group">
          <label>任务名称</label>
          <input v-model="taskForm.name" type="text" placeholder="输入任务名称" />
        </div>
        <div class="form-group">
          <label>任务描述</label>
          <textarea v-model="taskForm.description" placeholder="输入任务描述（可选）" rows="3"></textarea>
        </div>
        <div class="form-group">
          <label>选择标签（可选）</label>
          <select v-model="taskForm.tag_id">
            <option :value="null">无标签</option>
            <option v-for="tag in tags" :key="tag.id" :value="tag.id">
              {{ tag.name }}
            </option>
          </select>
        </div>
        <div class="modal-actions">
          <button class="btn btn-secondary" @click="showCreateModal = false">取消</button>
          <button class="btn btn-primary" @click="saveTask">保存</button>
        </div>
      </div>
    </div>

    <!-- 选择计时方式弹窗 -->
    <div v-if="showTimerModeModal" class="modal" @click.self="showTimerModeModal = false">
      <div class="modal-content timer-mode-modal">
        <h2>选择计时方式</h2>
        <div class="timer-mode-options">
          <div 
            class="timer-mode-card"
            :class="{ selected: selectedTimerMode === 'free' }"
            @click="selectedTimerMode = 'free'"
          >
            <div class="mode-icon">⏱️</div>
            <h3>自由计时</h3>
            <p>不限时长，自由控制计时节奏</p>
          </div>
          <div 
            class="timer-mode-card"
            :class="{ selected: selectedTimerMode === 'pomodoro' }"
            @click="selectedTimerMode = 'pomodoro'"
          >
            <div class="mode-icon">🍅</div>
            <h3>番茄钟</h3>
            <p>自定义工作和休息时间</p>
          </div>
        </div>

        <!-- 番茄钟设置 -->
        <div v-if="selectedTimerMode === 'pomodoro'" class="pomodoro-settings">
          <div class="setting-group">
            <label>工作时间（分钟）</label>
            <div class="time-input-group">
              <button class="btn-adjust" @click="adjustWorkTime(-5)">-5</button>
              <button class="btn-adjust" @click="adjustWorkTime(-1)">-1</button>
              <input 
                type="number" 
                v-model.number="pomodoroSettings.workMinutes" 
                min="1" 
                max="120"
                @input="validateWorkTime"
              />
              <button class="btn-adjust" @click="adjustWorkTime(1)">+1</button>
              <button class="btn-adjust" @click="adjustWorkTime(5)">+5</button>
            </div>
          </div>
          <div class="setting-group">
            <label>休息时间（分钟）</label>
            <div class="time-input-group">
              <button class="btn-adjust" @click="adjustBreakTime(-5)">-5</button>
              <button class="btn-adjust" @click="adjustBreakTime(-1)">-1</button>
              <input 
                type="number" 
                v-model.number="pomodoroSettings.breakMinutes" 
                min="1" 
                max="60"
                @input="validateBreakTime"
              />
              <button class="btn-adjust" @click="adjustBreakTime(1)">+1</button>
              <button class="btn-adjust" @click="adjustBreakTime(5)">+5</button>
            </div>
          </div>
          <div class="settings-preview">
            <span class="preview-text">
              🍅 工作 {{ pomodoroSettings.workMinutes }} 分钟 → 
              ☕ 休息 {{ pomodoroSettings.breakMinutes }} 分钟
            </span>
          </div>
        </div>

        <div class="modal-actions">
          <button class="btn btn-secondary" @click="showTimerModeModal = false">取消</button>
          <button class="btn btn-primary" @click="confirmStartTimer">开始计时</button>
        </div>
      </div>
    </div>

    <!-- 完成任务弹窗 -->
    <div v-if="showCompleteTaskModal" class="modal" @click.self="showCompleteTaskModal = false">
      <div class="modal-content">
        <h2>完成任务</h2>
        <div class="complete-task-info">
          <p class="task-name-display">{{ completingTask?.name }}</p>
          <p v-if="completingTask?.tag" class="task-tag-display">
            <span class="task-tag" :style="{ backgroundColor: completingTask.tag.color }">
              {{ completingTask.tag.name }}
            </span>
          </p>
        </div>
        <div class="form-group">
          <label>完成时间</label>
          <input 
            type="datetime-local" 
            v-model="completeTime" 
            :max="maxDateTime"
          />
          <p class="help-text">默认为当前时间，可以修改为实际完成时间</p>
        </div>
        <div class="modal-actions">
          <button class="btn btn-secondary" @click="showCompleteTaskModal = false">取消</button>
          <button class="btn btn-primary" @click="confirmCompleteTask">确认完成</button>
        </div>
      </div>
    </div>

    <!-- 自动归档设置弹窗 -->
    <div v-if="showSettingsModal" class="modal" @click.self="showSettingsModal = false">
      <div class="modal-content settings-modal">
        <h2>⚙️ 自动归档设置</h2>
        <div class="settings-description">
          <p>系统会在指定时间自动归档所有任务</p>
        </div>
        
        <div class="form-group">
          <label class="switch-label">
            <span>启用自动归档</span>
            <label class="switch">
              <input type="checkbox" v-model="settingsForm.auto_archive_enabled" />
              <span class="slider"></span>
            </label>
          </label>
        </div>

        <div class="form-group" v-if="settingsForm.auto_archive_enabled">
          <label>归档时间</label>
          <input 
            type="time" 
            v-model="settingsForm.auto_archive_time"
            class="time-input"
          />
          <p class="help-text">每天在此时间自动归档所有已完成的任务</p>
        </div>

        <div class="settings-preview" v-if="settingsForm.auto_archive_enabled">
          <div class="preview-icon">📦</div>
          <div class="preview-text">
            系统将在每天 <strong>{{ settingsForm.auto_archive_time }}</strong> 自动归档所有已完成的任务
          </div>
        </div>

        <div class="modal-actions">
          <button class="btn btn-secondary" @click="showSettingsModal = false">取消</button>
          <button class="btn btn-primary" @click="saveSettings">保存设置</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onActivated, watch } from 'vue'
import { useRouter } from 'vue-router'
import { taskAPI, timerAPI, type Task, type TimerConfig } from '../api/task'
import { tagAPI, type Tag } from '../api/tag'
import { settingsAPI, type UserSettings } from '../api/settings'
import AppHeader from '../components/AppHeader.vue'
import confetti from 'canvas-confetti'
import { toast, confirm } from '../utils/message'

// 定义组件名称以支持 keep-alive
defineOptions({
  name: 'Tasks'
})

const router = useRouter()
const tasks = ref<Task[]>([])
const tags = ref<Tag[]>([])
const selectedStatus = ref<number | null>(null) // null表示显示全部任务（未开始+进行中）
const showArchived = ref(false)
const groupedArchivedTasks = ref<Record<string, Task[]>>({})
const showCreateModal = ref(false)
const editingTask = ref<Task | null>(null)
const showTimerModeModal = ref(false)
const selectedTimerMode = ref<'free' | 'pomodoro'>('free')
const pendingTask = ref<Task | null>(null)
const showCompleteTaskModal = ref(false)
const completingTask = ref<Task | null>(null)
const completeTime = ref('')
const completingTaskIds = ref<Set<number>>(new Set())
const showSettingsModal = ref(false)

const taskForm = ref({
  name: '',
  description: '',
  tag_id: null as number | null
})

const settingsForm = ref({
  auto_archive_time: '00:00',
  auto_archive_enabled: true
})

const pomodoroSettings = ref({
  workMinutes: 25,
  breakMinutes: 5
})

const maxDateTime = computed(() => {
  const now = new Date()
  const year = now.getFullYear()
  const month = String(now.getMonth() + 1).padStart(2, '0')
  const day = String(now.getDate()).padStart(2, '0')
  const hours = String(now.getHours()).padStart(2, '0')
  const minutes = String(now.getMinutes()).padStart(2, '0')
  return `${year}-${month}-${day}T${hours}:${minutes}`
})

const selectStatus = (status: number | null) => {
  selectedStatus.value = status
  showArchived.value = false // 切换到状态筛选时，关闭归档视图
}

const filteredTasks = computed(() => {
  if (selectedStatus.value === null) {
    // 显示全部任务：未开始(0) + 进行中(1)
    return tasks.value.filter(task => task.status === 0 || task.status === 1)
  }
  return tasks.value.filter(task => task.status === selectedStatus.value)
})

const formatDate = (dateStr: string | undefined) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

const formatArchiveDate = (dateStr: string | undefined) => {
  if (!dateStr) return ''
  
  const date = new Date(dateStr)
  const today = new Date()
  const yesterday = new Date(today)
  yesterday.setDate(yesterday.getDate() - 1)
  
  const dateOnly = formatDate(dateStr)
  const todayStr = formatDate(today.toISOString())
  const yesterdayStr = formatDate(yesterday.toISOString())
  
  if (dateOnly === todayStr) return '今天'
  if (dateOnly === yesterdayStr) return '昨天'
  
  const year = date.getFullYear()
  const month = date.getMonth() + 1
  const day = date.getDate()
  const weekdays = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
  const weekday = weekdays[date.getDay()]
  
  return `${year}年${month}月${day}日 ${weekday}`
}

const formatDuration = (seconds: number) => {
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  return `${hours}小时${minutes}分钟`
}

const groupTasksByDate = (taskList: Task[]) => {
  const grouped: Record<string, Task[]> = {}
  
  taskList.forEach(task => {
    // 根据任务状态选择不同的日期字段
    let dateKey: string
    if (task.status === 2 && task.completed_at) {
      // 已完成任务按完成日期分组
      dateKey = task.completed_at?.split('T')[0] || ''
    } else {
      // 未开始和进行中任务按创建日期分组
      dateKey = task.created_at?.split('T')[0] || ''
    }
    
    if (dateKey) {
      if (!grouped[dateKey]) {
        grouped[dateKey] = []
      }
      grouped[dateKey]!.push(task)
    }
  })
  
  // 按日期倒序排序
  const sortedKeys = Object.keys(grouped).sort((a, b) => b.localeCompare(a))
  const result: Record<string, Task[]> = {}
  sortedKeys.forEach(key => {
    result[key] = grouped[key]!
  })
  
  return result
}

const loadTasks = async () => {
  try {
    const res: any = await taskAPI.getTasks()
    tasks.value = res.tasks || []
  } catch (err) {
    console.error('加载任务失败', err)
  }
}

const loadArchivedTasks = async () => {
  try {
    const res: any = await taskAPI.getArchivedTasks()
    groupedArchivedTasks.value = res.grouped_tasks || {}
  } catch (err) {
    console.error('加载归档任务失败', err)
  }
}

const toggleArchived = () => {
  showArchived.value = !showArchived.value
  if (showArchived.value) {
    loadArchivedTasks()
  }
}

const loadTags = async () => {
  try {
    const res: any = await tagAPI.getTags()
    tags.value = res.tags || []
  } catch (err) {
    console.error('加载标签失败', err)
  }
}

const editTask = (task: Task) => {
  editingTask.value = task
  taskForm.value = {
    name: task.name,
    description: task.description,
    tag_id: task.tag_id
  }
  showCreateModal.value = true
}

const saveTask = async () => {
  if (!taskForm.value.name) {
    toast.warning('请填写任务名称')
    return
  }

  try {
    if (editingTask.value) {
      await taskAPI.updateTask(editingTask.value.id, taskForm.value)
    } else {
      await taskAPI.createTask(taskForm.value as any)
    }
    showCreateModal.value = false
    editingTask.value = null
    taskForm.value = { name: '', description: '', tag_id: null }
    loadTasks()
  } catch (err) {
    console.error('保存任务失败', err)
  }
}

const deleteTask = async (id: number) => {
  if (!confirm('确定删除这个任务吗？')) return
  
  try {
    await taskAPI.deleteTask(id)
    loadTasks()
  } catch (err) {
    console.error('删除任务失败', err)
  }
}

const adjustWorkTime = (delta: number) => {
  const newValue = pomodoroSettings.value.workMinutes + delta
  pomodoroSettings.value.workMinutes = Math.max(1, Math.min(120, newValue))
}

const adjustBreakTime = (delta: number) => {
  const newValue = pomodoroSettings.value.breakMinutes + delta
  pomodoroSettings.value.breakMinutes = Math.max(1, Math.min(60, newValue))
}

const validateWorkTime = () => {
  if (pomodoroSettings.value.workMinutes < 1) pomodoroSettings.value.workMinutes = 1
  if (pomodoroSettings.value.workMinutes > 120) pomodoroSettings.value.workMinutes = 120
}

const validateBreakTime = () => {
  if (pomodoroSettings.value.breakMinutes < 1) pomodoroSettings.value.breakMinutes = 1
  if (pomodoroSettings.value.breakMinutes > 60) pomodoroSettings.value.breakMinutes = 60
}

const startTask = (task: Task) => {
  pendingTask.value = task
  selectedTimerMode.value = 'free' // 默认选择自由计时
  // 重置番茄钟设置为默认值
  pomodoroSettings.value = {
    workMinutes: 25,
    breakMinutes: 5
  }
  showTimerModeModal.value = true
}

const confirmStartTimer = async () => {
  if (!pendingTask.value) return
  
  try {
    const timerConfig: TimerConfig = selectedTimerMode.value === 'pomodoro' 
      ? {
          mode: 'pomodoro' as const,
          workMinutes: pomodoroSettings.value.workMinutes,
          breakMinutes: pomodoroSettings.value.breakMinutes
        }
      : { mode: 'free' as const }
    
    const res: any = await taskAPI.startTimer(pendingTask.value.id, timerConfig)
    showTimerModeModal.value = false
    router.push('/timer')
  } catch (err: any) {
    if (err.response?.data?.active_entry) {
      // 已有进行中的计时，询问是否切换
      const confirmed = await confirm('已有正在进行的计时任务，是否切换到当前任务？', '切换任务', { type: 'warning' })
      if (confirmed) {
        try {
          const timerConfig: TimerConfig = selectedTimerMode.value === 'pomodoro' 
            ? {
                mode: 'pomodoro' as const,
                workMinutes: pomodoroSettings.value.workMinutes,
                breakMinutes: pomodoroSettings.value.breakMinutes
              }
            : { mode: 'free' as const }
          
          await timerAPI.switchTimer(pendingTask.value.id, timerConfig)
          showTimerModeModal.value = false
          router.push('/timer')
        } catch (switchErr: any) {
          toast.error(switchErr.response?.data?.error || '切换任务失败')
        }
      }
    } else {
      toast.error(err.response?.data?.error || '开始计时失败')
    }
  }
}

const showCompleteModal = (task: Task) => {
  completingTask.value = task
  // 设置默认完成时间为当前时间
  const now = new Date()
  const year = now.getFullYear()
  const month = String(now.getMonth() + 1).padStart(2, '0')
  const day = String(now.getDate()).padStart(2, '0')
  const hours = String(now.getHours()).padStart(2, '0')
  const minutes = String(now.getMinutes()).padStart(2, '0')
  completeTime.value = `${year}-${month}-${day}T${hours}:${minutes}`
  showCompleteTaskModal.value = true
}

const confirmCompleteTask = async () => {
  if (!completingTask.value || !completeTime.value) return
  
  const taskId = completingTask.value.id
  const completedDate = completeTime.value.split('T')[0] // 获取完成日期
  
  try {
    // 添加到正在完成的任务集合，触发动画
    completingTaskIds.value.add(taskId)
    
    // 将本地时间转换为ISO格式
    const completedAt = new Date(completeTime.value).toISOString()
    await taskAPI.completeTaskWithTime(taskId, completedAt)
    
    showCompleteTaskModal.value = false
    completingTask.value = null
    
    // 等待动画完成后再重新加载任务列表
    setTimeout(async () => {
      await loadTasks()
      completingTaskIds.value.delete(taskId)
      
      // 检查是否是当天最后一个任务
      if (completedDate) {
        checkAndCelebrate(completedDate)
      }
    }, 600)
  } catch (err: any) {
    completingTaskIds.value.delete(taskId)
    toast.error(err.response?.data?.error || '完成任务失败')
  }
}

// 检查是否是当天最后一个任务，如果是则触发礼花动画
const checkAndCelebrate = (completedDate: string) => {
  // 获取当天所有未完成的任务（未开始或进行中）
  const todayIncompleteTasks = tasks.value.filter(task => {
    const taskDate = task.created_at?.split('T')[0] || ''
    return taskDate === completedDate && (task.status === 0 || task.status === 1)
  })
  
  // 如果当天没有未完成的任务了，触发礼花
  if (todayIncompleteTasks.length === 0) {
    triggerConfetti()
  }
}

// 触发礼花动画
const triggerConfetti = () => {
  const duration = 3000
  const animationEnd = Date.now() + duration
  const defaults = { startVelocity: 30, spread: 360, ticks: 60, zIndex: 2000 }

  function randomInRange(min: number, max: number) {
    return Math.random() * (max - min) + min
  }

  const interval = setInterval(function() {
    const timeLeft = animationEnd - Date.now()

    if (timeLeft <= 0) {
      return clearInterval(interval)
    }

    const particleCount = 50 * (timeLeft / duration)
    
    // 从左侧发射
    confetti({
      ...defaults,
      particleCount,
      origin: { x: randomInRange(0.1, 0.3), y: Math.random() - 0.2 }
    })
    
    // 从右侧发射
    confetti({
      ...defaults,
      particleCount,
      origin: { x: randomInRange(0.7, 0.9), y: Math.random() - 0.2 }
    })
  }, 250)
}

const archiveTask = async (id: number) => {
  if (!confirm('确定要归档这个任务吗？归档后可以在归档列表中查看。')) return
  
  try {
    await taskAPI.archiveTask(id)
    loadTasks()
  } catch (err: any) {
    toast.error(err.response?.data?.error || '归档任务失败')
  }
}

const unarchiveTask = async (id: number) => {
  if (!confirm('确定要取消归档吗？任务将恢复为已完成状态。')) return
  
  try {
    await taskAPI.unarchiveTask(id)
    loadArchivedTasks()
  } catch (err: any) {
    toast.error(err.response?.data?.error || '取消归档失败')
  }
}

const loadSettings = async () => {
  try {
    const res: any = await settingsAPI.getSettings()
    if (res.settings) {
      settingsForm.value = {
        auto_archive_time: res.settings.auto_archive_time,
        auto_archive_enabled: res.settings.auto_archive_enabled
      }
    }
  } catch (err) {
    console.error('加载设置失败', err)
  }
}

const saveSettings = async () => {
  try {
    await settingsAPI.updateSettings(settingsForm.value)
    showSettingsModal.value = false
    toast.success('设置已保存')
  } catch (err: any) {
    toast.error(err.response?.data?.error || '保存设置失败')
  }
}

onMounted(() => {
  loadTasks()
  loadTags()
  loadSettings()
})

// 当页面被激活时（从其他页面切换回来），重新加载标签列表
onActivated(() => {
  loadTags()
})
</script>

<style scoped>
.task-status {
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
  margin-right: 10px;
}

.status-0 {
  background-color: #f0f0f0;
  color: #666;
}

.status-1 {
  background-color: #fff3cd;
  color: #856404;
}

.status-2 {
  background-color: #d4edda;
  color: #155724;
}

.status-archived {
  background-color: #e2e8f0;
  color: #475569;
}

.archive-filter {
  font-weight: 500;
}

.archived-tasks {
  width: 100%;
}

.archived-group,
.task-group {
  margin-bottom: 32px;
}

.archived-date-header,
.task-date-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px 20px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 12px;
  margin-bottom: 16px;
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.2);
}

.date-icon {
  font-size: 24px;
}

.date-text {
  font-size: 18px;
  font-weight: 600;
  color: white;
  flex: 1;
}

.task-count {
  font-size: 14px;
  color: rgba(255, 255, 255, 0.9);
  background: rgba(255, 255, 255, 0.2);
  padding: 4px 12px;
  border-radius: 12px;
}

.archived-task-card {
  opacity: 0.95;
  border-left: 4px solid #94a3b8;
}

.archived-task-card:hover {
  opacity: 1;
}

.btn-archive {
  background-color: #64748b !important;
  color: white !important;
}

.btn-archive:hover {
  background-color: #475569 !important;
}

.btn-unarchive {
  background-color: #3b82f6 !important;
  color: white !important;
}

.btn-unarchive:hover {
  background-color: #2563eb !important;
}

.task-content h3 {
  font-size: 18px;
  margin-bottom: 8px;
}

.task-content p {
  color: #666;
  font-size: 14px;
}

.task-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 12px;
}

.task-duration {
  font-size: 13px;
  color: #999;
}

.modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0,0,0,0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: white;
  padding: 30px;
  border-radius: 8px;
  width: 90%;
  max-width: 600px;
}

.modal-content h2 {
  margin-bottom: 20px;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-weight: 500;
}

.form-group input,
.form-group textarea,
.form-group select {
  width: 100%;
  padding: 12px;
  border: 1px solid #e5e5e5;
  border-radius: 6px;
  font-size: 14px;
  font-family: inherit;
}

.form-group textarea {
  resize: vertical;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

.timer-mode-modal {
  max-width: 700px;
}

.timer-mode-options {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
  margin: 30px 0;
}

.timer-mode-card {
  border: 2px solid #e5e5e5;
  border-radius: 12px;
  padding: 30px 20px;
  text-align: center;
  cursor: pointer;
  transition: all 0.3s;
}

.timer-mode-card:hover {
  border-color: #667eea;
  transform: translateY(-4px);
  box-shadow: 0 8px 20px rgba(102, 126, 234, 0.2);
}

.timer-mode-card.selected {
  border-color: #667eea;
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
}

.mode-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.timer-mode-card h3 {
  font-size: 20px;
  margin-bottom: 12px;
  color: #333;
}

.timer-mode-card p {
  font-size: 14px;
  color: #666;
  line-height: 1.5;
}

.pomodoro-settings {
  background: #f8f9fa;
  border-radius: 8px;
  padding: 24px;
  margin-top: 20px;
}

.setting-group {
  margin-bottom: 20px;
}

.setting-group:last-child {
  margin-bottom: 0;
}

.setting-group label {
  display: block;
  font-weight: 500;
  margin-bottom: 12px;
  color: #333;
  font-size: 14px;
}

.time-input-group {
  display: flex;
  align-items: center;
  gap: 8px;
  justify-content: center;
}

.time-input-group input {
  width: 80px;
  padding: 10px;
  border: 2px solid #e5e5e5;
  border-radius: 6px;
  font-size: 18px;
  font-weight: 600;
  text-align: center;
  transition: all 0.3s;
}

.time-input-group input:focus {
  outline: none;
  border-color: #667eea;
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}

.btn-adjust {
  padding: 8px 12px;
  border: 1px solid #e5e5e5;
  border-radius: 6px;
  background: white;
  color: #666;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  min-width: 45px;
}

.btn-adjust:hover {
  border-color: #667eea;
  color: #667eea;
  background: rgba(102, 126, 234, 0.05);
}

.btn-adjust:active {
  transform: scale(0.95);
}

.settings-preview {
  margin-top: 20px;
  padding: 16px;
  background: white;
  border-radius: 8px;
  text-align: center;
  border: 2px dashed #667eea;
}

.preview-text {
  font-size: 15px;
  color: #667eea;
  font-weight: 500;
  line-height: 1.6;
}

.btn-complete {
  background-color: #10b981 !important;
  color: white !important;
  min-width: 44px;
  padding: 8px 16px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.btn-complete:hover {
  background-color: #059669 !important;
  transform: scale(1.05);
}

.check-icon {
  font-size: 20px;
  font-weight: bold;
  line-height: 1;
}

/* 任务完成动画 */
.task-completing {
  animation: slideOutRight 0.6s ease-in-out forwards;
}

@keyframes slideOutRight {
  0% {
    transform: translateX(0);
    opacity: 1;
  }
  50% {
    transform: translateX(30px);
    opacity: 0.5;
  }
  100% {
    transform: translateX(100%);
    opacity: 0;
  }
}

.complete-task-info {
  background: #f8f9fa;
  padding: 20px;
  border-radius: 8px;
  margin-bottom: 20px;
}

.task-name-display {
  font-size: 18px;
  font-weight: 600;
  color: #333;
  margin-bottom: 12px;
}

.task-tag-display {
  display: flex;
  align-items: center;
}

.help-text {
  font-size: 12px;
  color: #999;
  margin-top: 8px;
  margin-bottom: 0;
}

input[type="datetime-local"] {
  width: 100%;
  padding: 12px;
  border: 1px solid #e5e5e5;
  border-radius: 6px;
  font-size: 14px;
  font-family: inherit;
}

input[type="datetime-local"]:focus {
  outline: none;
  border-color: #667eea;
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}

.header-actions {
  display: flex;
  gap: 12px;
  align-items: center;
}

.btn-icon {
  background: white;
  border: 2px solid #e5e5e5;
  padding: 10px 16px;
  border-radius: 8px;
  font-size: 20px;
  cursor: pointer;
  transition: all 0.3s;
}

.btn-icon:hover {
  border-color: #667eea;
  background: rgba(102, 126, 234, 0.05);
  transform: translateY(-2px);
}

.settings-modal {
  max-width: 500px;
}

.settings-description {
  background: #f8f9fa;
  padding: 16px;
  border-radius: 8px;
  margin-bottom: 24px;
}

.settings-description p {
  margin: 0;
  color: #666;
  font-size: 14px;
  line-height: 1.6;
}

.switch-label {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
  margin-bottom: 0;
}

.switch-label span {
  font-weight: 500;
  color: #333;
}

.switch {
  position: relative;
  display: inline-block;
  width: 52px;
  height: 28px;
}

.switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: #ccc;
  transition: 0.4s;
  border-radius: 28px;
}

.slider:before {
  position: absolute;
  content: "";
  height: 20px;
  width: 20px;
  left: 4px;
  bottom: 4px;
  background-color: white;
  transition: 0.4s;
  border-radius: 50%;
}

input:checked + .slider {
  background-color: #667eea;
}

input:checked + .slider:before {
  transform: translateX(24px);
}

.time-input {
  font-size: 16px;
  font-weight: 500;
  padding: 12px;
}

.settings-preview {
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%);
  border: 2px solid #667eea;
  border-radius: 12px;
  padding: 20px;
  margin-top: 24px;
  margin-bottom: 24px;
  display: flex;
  align-items: center;
  gap: 16px;
}

.preview-icon {
  font-size: 32px;
}

.preview-text {
  flex: 1;
  color: #333;
  font-size: 14px;
  line-height: 1.6;
}

.preview-text strong {
  color: #667eea;
  font-weight: 600;
  font-size: 16px;
}
</style>
