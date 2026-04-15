import axios from './axios'

export interface Note {
  id: number
  user_id: number
  content: string
  images?: string
  is_deleted: boolean
  deleted_at?: string
  created_at: string
  updated_at: string
  tags: Tag[]
}

export interface Tag {
  id: number
  name: string
  color: string
}

export interface CreateNoteRequest {
  content: string
  images?: string
  tag_ids?: number[]
}

export const noteAPI = {
  getNotes: (tagId?: number, showDeleted?: boolean) => 
    axios.get('/notes', { params: { tag_id: tagId, show_deleted: showDeleted } }),
  getNote: (id: number) => axios.get(`/notes/${id}`),
  createNote: (data: CreateNoteRequest) => axios.post('/notes', data),
  updateNote: (id: number, data: CreateNoteRequest) => axios.put(`/notes/${id}`, data),
  deleteNote: (id: number) => axios.delete(`/notes/${id}`),
  restoreNote: (id: number) => axios.post(`/notes/${id}/restore`),
  emptyTrash: () => axios.post('/notes/trash/empty'),
  getCalendar: (year: number, month: number) => axios.get(`/notes/calendar/${year}/${month}`),
  uploadImage: (file: File) => {
    const formData = new FormData()
    formData.append('image', file)
    return axios.post('/notes/upload-image', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  },
  deleteImage: (url: string) => axios.post('/notes/delete-image', { url }),
}
