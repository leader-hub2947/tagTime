# TagTime 后端服务

基于 Gin 框架的 TagTime 后端 API 服务。

## 技术栈

- Go 1.21+
- Gin Web Framework
- GORM (MySQL)
- JWT 认证
- bcrypt 密码加密

## 快速开始

### 1. 安装依赖

```bash
go mod download
```

### 2. 配置数据库

修改 `config/database.go` 中的数据库连接信息：

```go
dsn := "root:password@tcp(127.0.0.1:3306)/tagtime?charset=utf8mb4&parseTime=True&loc=Local"
```

### 3. 创建数据库

```sql
CREATE DATABASE tagtime CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 4. 运行服务

```bash
go run main.go
```

服务将在 `http://localhost:8080` 启动。

## API 文档

### 认证

- `POST /api/v1/auth/register` - 用户注册
- `POST /api/v1/auth/login` - 用户登录

### 标签管理

- `GET /api/v1/tags` - 获取标签列表
- `POST /api/v1/tags` - 创建标签
- `PUT /api/v1/tags/:id` - 更新标签
- `DELETE /api/v1/tags/:id` - 删除标签

### 便签管理

- `GET /api/v1/notes` - 获取便签列表
- `GET /api/v1/notes/:id` - 获取便签详情
- `POST /api/v1/notes` - 创建便签
- `PUT /api/v1/notes/:id` - 更新便签
- `DELETE /api/v1/notes/:id` - 删除便签
- `GET /api/v1/notes/calendar/:year/:month` - 获取月度日历数据

### 任务管理

- `GET /api/v1/tasks` - 获取任务列表
- `GET /api/v1/tasks/:id` - 获取任务详情
- `POST /api/v1/tasks` - 创建任务
- `PUT /api/v1/tasks/:id` - 更新任务
- `DELETE /api/v1/tasks/:id` - 删除任务

### 计时功能

- `POST /api/v1/tasks/:id/start` - 开始任务计时
- `POST /api/v1/timer/:id/end` - 结束计时
- `GET /api/v1/timer/current` - 获取当前计时
- `POST /api/v1/timer/switch` - 切换任务

### 数据可视化

- `GET /api/v1/dashboard/timeline` - 获取时间轴数据
- `GET /api/v1/dashboard/tag-ranking` - 获取标签排行
- `GET /api/v1/dashboard/task-statistics` - 获取任务统计

### AI 洞察

- `POST /api/v1/ai/crush` - 生成 AI 击溃语
- `GET /api/v1/ai/crush/remaining` - 获取剩余调用次数

## AI 洞察功能

### 功能说明

"用一句话击溃我"是一个创新的 AI 功能，通过分析用户的历史行为数据（便签内容、任务完成情况、计时记录等），生成直击内心痛点的个性化评语，帮助用户反思自己的行为模式。

### 配置说明

在 `.env` 文件中配置以下环境变量：

```bash
# AI 服务配置
# 支持的提供商: openai, zhipu, deepseek
AI_PROVIDER=deepseek                                  # AI 服务提供商（推荐使用 deepseek）
AI_API_KEY=your_api_key_here                         # API 密钥

# DeepSeek 配置（推荐）
AI_ENDPOINT=https://api.deepseek.com/chat/completions  # DeepSeek API 端点
AI_MODEL=deepseek-reasoner                            # DeepSeek 推理模型

# OpenAI 配置（备选）
# AI_ENDPOINT=https://api.openai.com/v1/chat/completions
# AI_MODEL=gpt-3.5-turbo

# 智谱AI 配置（备选）
# AI_ENDPOINT=https://open.bigmodel.cn/api/paas/v4/chat/completions
# AI_MODEL=glm-4

AI_TIMEOUT=30s                                        # 请求超时时间
AI_DAILY_LIMIT=3                                      # 每日调用次数限制
AI_CACHE_EXPIRE=10m                                   # 缓存过期时间
AI_GLOBAL_RATE_LIMIT=10                              # 全局限流（次/分钟）

# Redis 配置（用于缓存和限流）
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
```

### 获取 DeepSeek API Key

1. 访问 [DeepSeek 开放平台](https://platform.deepseek.com/)
2. 注册并登录账号
3. 在控制台创建 API Key
4. 将 API Key 配置到 `.env` 文件中的 `AI_API_KEY`

### 数据库索引

运行以下 SQL 脚本以优化查询性能：

```bash
mysql -u root -p tagtime < migrations/add_ai_crush_indexes.sql
```

### 使用示例

```bash
# 生成击溃语
curl -X POST http://localhost:8080/api/v1/ai/crush \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# 响应示例
{
  "crush_line": "你的'明天'已经说了三个月，但今天依然是昨天的重复。",
  "remaining_count": 2
}

# 获取剩余次数
curl -X GET http://localhost:8080/api/v1/ai/crush/remaining \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# 响应示例
{
  "remaining_count": 2
}
```

### 功能特性

- **智能分析**: 深度分析用户行为模式（拖延指数、坚持指数、压力指标等）
- **数据脱敏**: 自动移除敏感信息（手机号、邮箱、身份证等）
- **限流控制**: 每用户每天最多 3 次调用
- **缓存优化**: 10 分钟内重复请求返回缓存结果
- **降级策略**: AI 服务不可用时返回预设击溃语
- **性能优化**: 并发查询，数据提取 < 500ms

## 项目结构

```
gin-tagTime/
├── config/          # 配置文件
│   ├── ai.go        # AI 服务配置
│   ├── database.go  # 数据库配置
│   ├── jwt.go       # JWT 配置
│   ├── redis.go     # Redis 配置
│   └── utils.go     # 配置工具函数
├── controllers/     # 控制器
│   ├── ai_crush.go  # AI 洞察控制器
│   └── ...
├── services/        # 服务层
│   ├── ai_crush_service.go      # AI 击溃服务
│   ├── data_extractor.go        # 数据提取服务
│   ├── behavior_analyzer.go     # 行为分析服务
│   ├── data_sanitizer.go        # 数据脱敏服务
│   ├── prompt_builder.go        # 提示词构建服务
│   ├── ai_client.go             # AI 客户端
│   ├── cache_manager.go         # 缓存管理
│   ├── rate_limiter.go          # 限流器
│   └── fallback_strategy.go    # 降级策略
├── middleware/      # 中间件
├── models/          # 数据模型
│   ├── ai_crush.go  # AI 洞察模型
│   └── ...
├── routes/          # 路由
├── utils/           # 工具函数
│   └── error_handler.go  # 错误处理
├── migrations/      # 数据库迁移
│   └── add_ai_crush_indexes.sql
└── main.go          # 入口文件
```
