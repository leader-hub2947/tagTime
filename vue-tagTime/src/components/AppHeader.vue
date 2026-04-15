<template>
  <header class="header">
    <div class="logo" @click="$router.push('/notes')">TagTime</div>
    <nav class="nav-menu">
      <router-link to="/notes" class="nav-link">
        <span class="nav-text">全部笔记</span>
      </router-link>
      <router-link to="/tasks" class="nav-link">
        <span class="nav-text">任务管理</span>
      </router-link>
      <router-link to="/dashboard" class="nav-link">
        <span class="nav-text">数据统计</span>
      </router-link>
      <router-link to="/ai-crush" class="nav-link">
        <span class="nav-text">AI洞悉</span>
      </router-link>
    </nav>
    <button class="logout-btn" @click="handleLogout">退出登录</button>
  </header>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { toast, confirm } from '../utils/message'

const router = useRouter()

const handleLogout = async () => {
  const confirmed = await confirm('确定要退出登录吗？', '退出登录')
  if (confirmed) {
    // 清除本地存储
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    toast.success('已退出登录')
    // 跳转到登录页
    router.push('/login')
  }
}
</script>

<style scoped>
.nav-link {
  position: relative;
  padding: 8px 0;
}

.nav-link::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 0;
  width: 0;
  height: 2px;
  background-color: #333;
  transition: width 0.3s ease;
}

.nav-link:hover::after {
  width: 100%;
}

.nav-link.router-link-active {
  color: #333;
  font-weight: 600;
}

.nav-link.router-link-active::after {
  width: 100%;
}

.nav-text {
  display: inline-block;
  transition: transform 0.2s ease;
}

.nav-link:hover .nav-text {
  transform: translateY(-2px);
}

.logout-btn {
  padding: 8px 16px;
  border: 1px solid #e5e5e5;
  border-radius: 6px;
  background-color: #fff;
  color: #666;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.3s ease;
}

.logout-btn:hover {
  background-color: #333;
  color: #fff;
  border-color: #333;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.logout-btn:active {
  transform: translateY(0);
}
</style>
