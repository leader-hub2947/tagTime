import axios from './axios'

export interface Tag {
  id: number
  user_id: number
  name: string
  color: string
  created_at: string
  updated_at: string
}

export interface CreateTagRequest {
  name: string
  color?: string
}

export interface DeleteTagRequest {
  delete_notes: boolean
}

export const tagAPI = {
  getTags: () => axios.get('/tags'),
  createTag: (data: CreateTagRequest) => axios.post('/tags', data),
  updateTag: (id: number, data: CreateTagRequest) => axios.put(`/tags/${id}`, data),
  deleteTag: (id: number, data: DeleteTagRequest) => axios.delete(`/tags/${id}`, { data }),
}
