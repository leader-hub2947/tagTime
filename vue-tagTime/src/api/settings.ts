import axios from './axios'

export interface UserSettings {
  id: number
  user_id: number
  auto_archive_time: string
  auto_archive_enabled: boolean
  created_at: string
  updated_at: string
}

export interface UpdateSettingsRequest {
  auto_archive_time?: string
  auto_archive_enabled?: boolean
}

export const settingsAPI = {
  getSettings: () => axios.get('/settings'),
  updateSettings: (data: UpdateSettingsRequest) => axios.put('/settings', data),
}
