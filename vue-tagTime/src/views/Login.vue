<template>
  <div class="login-container">
    <!-- 左侧动画区域 -->
    <div class="login-left">
      <div class="login-left-header">
        <div class="logo">
          <svg width="32" height="32" viewBox="0 0 32 32" fill="none" xmlns="http://www.w3.org/2000/svg">
            <rect width="32" height="32" rx="8" fill="white" fill-opacity="0.1"/>
            <circle cx="16" cy="16" r="10" stroke="white" stroke-width="2"/>
            <circle cx="16" cy="16" r="4" fill="white"/>
          </svg>
          <span>TagTime</span>
        </div>
      </div>

      <!-- 动画角色区域 -->
      <div class="animated-characters-container">
        <AnimatedCharacters
          :isTyping="isTyping"
          :showPassword="showPassword"
          :passwordLength="form.password.length"
        />
      </div>

      <div class="login-left-footer">
        <a href="#">隐私政策</a>
        <a href="#">服务条款</a>
      </div>

      <!-- 装饰元素 -->
      <div class="decorative-grid"></div>
      <div class="decorative-blob blob-1"></div>
      <div class="decorative-blob blob-2"></div>
    </div>

    <!-- 右侧表单区域 -->
    <div class="login-right">
      <div class="login-form-wrapper">
        <!-- 移动端 Logo -->
        <div class="mobile-logo">
          <svg width="32" height="32" viewBox="0 0 32 32" fill="none" xmlns="http://www.w3.org/2000/svg">
            <rect width="32" height="32" rx="8" fill="#333"/>
            <circle cx="16" cy="16" r="10" stroke="white" stroke-width="2"/>
            <circle cx="16" cy="16" r="4" fill="white"/>
          </svg>
          <span>TagTime</span>
        </div>

        <!-- 标题 -->
        <div class="login-header">
          <h1>{{ isRegister ? '创建账户' : '欢迎回来！' }}</h1>
          <p>{{ isRegister ? '请填写以下信息完成注册' : '请输入您的登录信息' }}</p>
        </div>

        <!-- 登录/注册表单 -->
        <form @submit.prevent="handleSubmit" class="login-form">
          <!-- 用户名 -->
          <div class="form-group">
            <label for="username">用户名</label>
            <input
              id="username"
              v-model="form.username"
              type="text"
              placeholder="请输入用户名"
              required
              @focus="isTyping = true"
              @blur="isTyping = false"
            />
          </div>

          <!-- 邮箱（仅注册时显示） -->
          <div v-if="isRegister" class="form-group">
            <label for="email">邮箱</label>
            <input
              id="email"
              v-model="form.email"
              type="email"
              placeholder="you@example.com"
              required
            />
          </div>

          <!-- 密码 -->
          <div class="form-group">
            <label for="password">密码</label>
            <div class="password-input-wrapper">
              <input
                id="password"
                v-model="form.password"
                :type="showPassword ? 'text' : 'password'"
                placeholder="••••••••"
                required
              />
              <button
                type="button"
                class="toggle-password"
                @click="showPassword = !showPassword"
              >
                <svg v-if="showPassword" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M9.88 9.88a3 3 0 1 0 4.24 4.24"/>
                  <path d="M10.73 5.08A10.43 10.43 0 0 1 12 5c7 0 10 7 10 7a13.16 13.16 0 0 1-1.67 2.68"/>
                  <path d="M6.61 6.61A13.526 13.526 0 0 0 2 12s3 7 10 7a9.74 9.74 0 0 0 5.39-1.61"/>
                  <line x1="2" x2="22" y1="2" y2="22"/>
                </svg>
                <svg v-else width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M2 12s3-7 10-7 10 7 10 7-3 7-10 7-10-7-10-7Z"/>
                  <circle cx="12" cy="12" r="3"/>
                </svg>
              </button>
            </div>
          </div>

          <!-- 记住我和忘记密码 -->
          <div v-if="!isRegister" class="form-options">
            <div class="remember-me">
              <input
                id="remember"
                v-model="rememberMe"
                type="checkbox"
              />
              <label for="remember">30天内记住我</label>
            </div>
            <a href="#" class="forgot-password">忘记密码？</a>
          </div>

          <!-- 错误提示 -->
          <div v-if="error" class="error-message">
            {{ error }}
          </div>

          <!-- 提交按钮 -->
          <button
            type="submit"
            class="btn-submit"
            :disabled="isLoading"
          >
            <span class="btn-text">{{ isLoading ? '处理中...' : (isRegister ? '注册' : '登录') }}</span>
            <span class="btn-hover-text">{{ isLoading ? '处理中...' : (isRegister ? '注册' : '登录') }}</span>
            <svg class="btn-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M5 12h14"/>
              <path d="m12 5 7 7-7 7"/>
            </svg>
          </button>
        </form>

        <!-- 切换登录/注册 -->
        <div class="toggle-mode">
          {{ isRegister ? '已有账户？' : '还没有账户？' }}
          <a @click="toggleMode">{{ isRegister ? '立即登录' : '立即注册' }}</a>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { authAPI } from '../api/auth'
import { toast } from '../utils/message'
import AnimatedCharacters from '../components/AnimatedCharacters.vue'

const router = useRouter()
const isRegister = ref(false)
const isLoading = ref(false)
const error = ref('')
const showPassword = ref(false)
const isTyping = ref(false)
const rememberMe = ref(false)

const form = reactive({
  username: '',
  email: '',
  password: ''
})

const toggleMode = () => {
  isRegister.value = !isRegister.value
  error.value = ''
  form.password = ''
  if (!isRegister.value) {
    form.email = ''
  }
}

const handleSubmit = async () => {
  error.value = ''
  isLoading.value = true

  try {
    if (isRegister.value) {
      await authAPI.register({
        username: form.username,
        email: form.email,
        password: form.password
      })
      toast.success('注册成功，请登录')
      isRegister.value = false
      form.password = ''
      form.email = ''
    } else {
      const res: any = await authAPI.login({
        username: form.username,
        password: form.password
      })

      // 存储 token
      localStorage.setItem('token', res.token)
      localStorage.setItem('user', JSON.stringify(res.user))

      // 记住我功能
      if (rememberMe.value) {
        localStorage.setItem('rememberUsername', form.username)
      } else {
        localStorage.removeItem('rememberUsername')
      }

      toast.success('登录成功')
      router.push('/notes')
    }
  } catch (err: any) {
    error.value = err.response?.data?.error || (isRegister.value ? '注册失败' : '登录失败')
    toast.error(error.value)
  } finally {
    isLoading.value = false
  }
}

// 检查是否有记住的用户名
const savedUsername = localStorage.getItem('rememberUsername')
if (savedUsername) {
  form.username = savedUsername
  rememberMe.value = true
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: grid;
  grid-template-columns: 1fr 1fr;
  overflow: hidden;
}

/* 左侧动画区域 */
.login-left {
  position: relative;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding: 48px;
  background: linear-gradient(135deg, #9ca3af 0%, #6b7280 50%, #4b5563 100%);
  color: white;
  overflow: hidden;
}

.login-left-header {
  position: relative;
  z-index: 20;
}

.logo {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 20px;
  font-weight: 600;
}

.animated-characters-container {
  position: relative;
  z-index: 20;
  display: flex;
  align-items: flex-end;
  justify-content: center;
  height: 500px;
}

.login-left-footer {
  position: relative;
  z-index: 20;
  display: flex;
  gap: 32px;
  font-size: 14px;
  color: rgba(255, 255, 255, 0.8);
}

.login-left-footer a {
  color: rgba(255, 255, 255, 0.8);
  text-decoration: none;
  transition: color 0.2s;
}

.login-left-footer a:hover {
  color: white;
}

/* 装饰元素 */
.decorative-grid {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(255,255,255,0.03) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255,255,255,0.03) 1px, transparent 1px);
  background-size: 20px 20px;
  pointer-events: none;
}

.decorative-blob {
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
  pointer-events: none;
}

.blob-1 {
  top: 25%;
  right: 25%;
  width: 256px;
  height: 256px;
  background: rgba(156, 163, 175, 0.3);
}

.blob-2 {
  bottom: 25%;
  left: 25%;
  width: 384px;
  height: 384px;
  background: rgba(209, 213, 219, 0.2);
}

/* 右侧表单区域 */
.login-right {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 32px;
  background: #ffffff;
}

.login-form-wrapper {
  width: 100%;
  max-width: 420px;
}

.mobile-logo {
  display: none;
  align-items: center;
  justify-content: center;
  gap: 12px;
  font-size: 20px;
  font-weight: 600;
  margin-bottom: 48px;
  color: #333;
}

.login-header {
  text-align: center;
  margin-bottom: 32px;
}

.login-header h1 {
  font-size: 28px;
  font-weight: 700;
  color: #111827;
  margin-bottom: 8px;
  letter-spacing: -0.02em;
}

.login-header p {
  font-size: 14px;
  color: #6b7280;
}

/* 表单样式 */
.login-form {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-group label {
  font-size: 14px;
  font-weight: 500;
  color: #374151;
}

.form-group input[type="text"],
.form-group input[type="email"],
.form-group input[type="password"] {
  height: 48px;
  padding: 0 16px;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  font-size: 14px;
  background: #ffffff;
  color: #111827;
  transition: all 0.2s;
}

.form-group input:focus {
  outline: none;
  border-color: #333;
  box-shadow: 0 0 0 3px rgba(0, 0, 0, 0.05);
}

.form-group input::placeholder {
  color: #9ca3af;
}

.password-input-wrapper {
  position: relative;
}

.password-input-wrapper input {
  width: 100%;
  padding-right: 48px;
}

.toggle-password {
  position: absolute;
  right: 12px;
  top: 50%;
  transform: translateY(-50%);
  background: none;
  border: none;
  padding: 4px;
  cursor: pointer;
  color: #9ca3af;
  transition: color 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.toggle-password:hover {
  color: #6b7280;
}

/* 选项区域 */
.form-options {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: -4px;
}

.remember-me {
  display: flex;
  align-items: center;
  gap: 8px;
}

.remember-me input[type="checkbox"] {
  width: 16px;
  height: 16px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  cursor: pointer;
  accent-color: #333;
}

.remember-me label {
  font-size: 14px;
  color: #4b5563;
  cursor: pointer;
}

.forgot-password {
  font-size: 14px;
  color: #333;
  text-decoration: none;
  font-weight: 500;
  transition: opacity 0.2s;
}

.forgot-password:hover {
  opacity: 0.7;
}

/* 错误消息 */
.error-message {
  padding: 12px 16px;
  background: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 8px;
  color: #dc2626;
  font-size: 14px;
}

/* 提交按钮 */
.btn-submit {
  position: relative;
  height: 48px;
  border: 1px solid #e5e7eb;
  border-radius: 24px;
  background: #ffffff;
  color: #111827;
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  overflow: hidden;
  transition: all 0.3s ease;
  margin-top: 8px;
}

.btn-submit:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.btn-text {
  display: inline-block;
  transition: all 0.3s ease;
}

.btn-hover-text {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  background: #333;
  color: white;
  border-radius: 24px;
  opacity: 0;
  transform: translateX(-100%);
  transition: all 0.3s ease;
}

.btn-icon {
  opacity: 0;
  transform: translateX(-10px);
  transition: all 0.3s ease;
}

.btn-submit:hover:not(:disabled) .btn-text {
  opacity: 0;
  transform: translateX(20px);
}

.btn-submit:hover:not(:disabled) .btn-hover-text {
  opacity: 1;
  transform: translateX(0);
}

.btn-submit:hover:not(:disabled) .btn-icon {
  opacity: 1;
  transform: translateX(0);
}

/* 切换模式 */
.toggle-mode {
  text-align: center;
  margin-top: 24px;
  font-size: 14px;
  color: #6b7280;
}

.toggle-mode a {
  color: #111827;
  font-weight: 600;
  text-decoration: none;
  cursor: pointer;
  margin-left: 4px;
  transition: opacity 0.2s;
}

.toggle-mode a:hover {
  opacity: 0.7;
  text-decoration: underline;
}

/* 响应式设计 */
@media (max-width: 1024px) {
  .login-container {
    grid-template-columns: 1fr;
  }

  .login-left {
    display: none;
  }

  .mobile-logo {
    display: flex;
  }

  .login-right {
    padding: 24px;
  }
}

@media (max-width: 640px) {
  .login-form-wrapper {
    max-width: 100%;
  }

  .login-header h1 {
    font-size: 24px;
  }

  .form-options {
    flex-direction: column;
    gap: 12px;
    align-items: flex-start;
  }
}
</style>
