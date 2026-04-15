import axios from 'axios'

// 自动检测API地址：支持本地、局域网、外网访问
const getBaseURL = () => {
  // 优先使用环境变量配置
  if (import.meta.env.VITE_API_BASE_URL) {
    return import.meta.env.VITE_API_BASE_URL
  }
  
  // 使用相对路径，让浏览器自动使用当前域名和协议
  // 适用于前后端通过 Nginx 反向代理部署在同一域名的场景
  return '/api/v1'
}

const instance = axios.create({
  baseURL: getBaseURL(),
  timeout: 60000, // 60秒超时，适配 AI 接口调用
})

instance.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

instance.interceptors.response.use(
  (response) => {
    return response.data
  },
  (error) => {
    if (error.response?.status === 401) {
      // 清除token
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      // 只有不在登录页时才跳转
      if (window.location.pathname !== '/login') {
        window.location.href = '/login'
      }
    }
    return Promise.reject(error)
  }
)

export default instance
