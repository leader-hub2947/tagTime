# TagTime 前端

基于 Vue 3 + TypeScript + Vite 的 TagTime 前端应用。

## 技术栈

- Vue 3 (Composition API)
- TypeScript
- Vue Router
- Pinia
- Axios
- Vite

## 快速开始

### 1. 安装依赖

```bash
npm install
```

### 2. 启动开发服务器

```bash
npm run dev
```

应用将在 `http://localhost:5173` 启动。

### 3. 构建生产版本

```bash
npm run build
```

## 功能模块

- 用户登录/注册
- 便签管理（创建、编辑、删除、标签筛选）
- 任务管理（创建、编辑、删除、计时）
- 数据统计（任务完成度、标签排行、时间轴）
- 标签管理（创建、编辑、删除）

## 项目结构

```
vue-tagTime/
├── src/
│   ├── api/           # API 接口
│   ├── router/        # 路由配置
│   ├── views/         # 页面组件
│   ├── App.vue        # 根组件
│   ├── main.ts        # 入口文件
│   └── style.css      # 全局样式
├── public/            # 静态资源
└── index.html         # HTML 模板
```

## 注意事项

- 确保后端服务已启动在 `http://localhost:8080`
- 首次使用需要注册账户
