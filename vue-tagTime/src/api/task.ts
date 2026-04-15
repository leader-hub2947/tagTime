import axios from './axios'

export interface Task {
  id: number
  user_id: number
  tag_id: number | null
  name: string
  description: string
  status: number
  total_duration: number
  created_at: string
  updated_at: string
  completed_at?: string
  archived_at?: string
  tag?: {
    id: number
    name: string
    color: string
  }
}

export interface CreateTaskRequest {
  tag_id: number
  name: string
  description?: string
}

export interface TimerConfig {
  mode: 'free' | 'pomodoro'
  workMinutes?: number
  breakMinutes?: number
}

export const taskAPI = {
  getTasks: (params?: { tag_id?: number; status?: number; archived?: boolean }) => axios.get('/tasks', { params: { ...params, archived: params?.archived ? 'true' : undefined } }),
  getTask: (id: number) => axios.get(`/tasks/${id}`),
  createTask: (data: CreateTaskRequest) => axios.post('/tasks', data),
  updateTask: (id: number, data: any) => axios.put(`/tasks/${id}`, data),
  deleteTask: (id: number) => axios.delete(`/tasks/${id}`),
  startTimer: (id: number, config: TimerConfig) => axios.post(`/tasks/${id}/start`, { 
    timer_mode: config.mode,
    work_minutes: config.workMinutes,
    break_minutes: config.breakMinutes
  }),
  completeTask: (id: number) => axios.post(`/tasks/${id}/complete`),
  completeTaskWithTime: (id: number, completedAt: string) => axios.post(`/tasks/${id}/complete`, { completed_at: completedAt }),
  archiveTask: (id: number) => axios.post(`/tasks/${id}/archive`),
  unarchiveTask: (id: number) => axios.post(`/tasks/${id}/unarchive`),
  getArchivedTasks: () => axios.get('/tasks/archived'),
}

export const timerAPI = {
  pauseTimer: () => axios.post('/timer/pause'),
  resumeTimer: () => axios.post('/timer/resume'),
  endTimer: () => axios.post('/timer/end'),
  getCurrentTimer: () => axios.get('/timer/current'),
  switchTimer: (newTaskId: number, config: TimerConfig) => axios.post('/timer/switch', { 
    new_task_id: newTaskId,
    timer_mode: config.mode,
    work_minutes: config.workMinutes,
    break_minutes: config.breakMinutes
  }),
}

export const dashboardAPI = {
  getTimeline: (date?: string) => axios.get('/dashboard/timeline', { params: { date } }),
  getTagRanking: () => axios.get('/dashboard/tag-ranking'),
  getTaskStatistics: (period: string) => axios.get('/dashboard/task-statistics', { params: { period } }),
}
