import axios from './axios'

export interface LoginRequest {
  username: string
  password: string
}

export interface RegisterRequest {
  username: string
  email: string
  password: string
}

export const authAPI = {
  login: (data: LoginRequest) => axios.post('/auth/login', data),
  register: (data: RegisterRequest) => axios.post('/auth/register', data),
}
