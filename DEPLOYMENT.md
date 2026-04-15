# TagTime 项目部署指南

## 项目结构

```
tagTime/
├── gin-tagTime/          # 后端服务（Go + Gin）
└── vue-tagTime/          # 前端应用（Vue 3 + TypeScript）
```

## 一、后端部署（gin-tagTime）

### 1.1 环境要求

- Go 1.19+
- MySQL 5.7+ 或 8.0+
- Redis 5.0+
- Linux/Windows 服务器

### 1.2 配置步骤

#### 步骤 1：配置环境变量

```bash
cd gin-tagTime
cp .env.production .env
```

编辑 `.env` 文件，修改以下关键配置：

```env
# 服务器配置
SERVER_HOST=0.0.0.0              # 监听所有网络接口
SERVER_PORT=8080                 # 后端端口
GIN_MODE=release                 # 生产模式

# CORS 配置（重要！）
# 配置你的前端域名，多个用逗号分隔
CORS_ALLOWED_ORIGINS=https://your-domain.com,https://www.your-domain.com

# 数据库配置
DB_USER=root
DB_PASSWORD=your_secure_password
DB_HOST=127.0.0.1
DB_PORT=3306
DB_NAME=tagtime

# Redis 配置
REDIS_HOST=127.0.0.1
REDIS_PORT=6379
REDIS_PASSWORD=your_redis_password
REDIS_DB=0

# AI 服务配置
AI_PROVIDER=deepseek
AI_API_KEY=your_deepseek_api_key
AI_ENDPOINT=https://api.deepseek.com/chat/completions
AI_MODEL=deepseek-reasoner
```

#### 步骤 2：创建数据库

```sql
CREATE DATABASE tagtime CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

#### 步骤 3：编译后端

```bash
# 安装依赖
go mod download

# 编译（Linux）
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o tagtime-server main.go

# 编译（Windows）
go build -o tagtime-server.exe main.go
```

#### 步骤 4：运行后端

```bash
# 直接运行
./tagtime-server

# 或使用 nohup 后台运行（Linux）
nohup ./tagtime-server > tagtime.log 2>&1 &

# 或使用 systemd 服务（推荐）
```

### 1.3 使用 systemd 管理服务（Linux 推荐）

创建服务文件 `/etc/systemd/system/tagtime.service`：

```ini
[Unit]
Description=TagTime Backend Service
After=network.target mysql.service redis.service

[Service]
Type=simple
User=www-data
WorkingDirectory=/path/to/gin-tagTime
ExecStart=/path/to/gin-tagTime/tagtime-server
Restart=on-failure
RestartSec=5s

# 环境变量（可选，也可以使用 .env 文件）
Environment="GIN_MODE=release"

[Install]
WantedBy=multi-user.target
```

启动服务：

```bash
sudo systemctl daemon-reload
sudo systemctl enable tagtime
sudo systemctl start tagtime
sudo systemctl status tagtime
```

查看日志：

```bash
sudo journalctl -u tagtime -f
```

---

## 二、前端部署（vue-tagTime）

### 2.1 环境要求

- Node.js 16+
- npm 或 yarn

### 2.2 配置步骤

#### 步骤 1：配置 API 地址

编辑 `.env.production` 文件：

```env
# 方式 1：使用相对路径（推荐，前后端同域名部署）
# 留空，前端会使用相对路径 /api/v1，适用于 Nginx 反向代理
VITE_API_BASE_URL=

# 方式 2：指定完整的后端 API 地址（前后端分离部署）
# 取消下面的注释并修改为实际的后端地址
# VITE_API_BASE_URL=https://api.your-domain.com/api/v1

# 方式 3：使用 IP 地址（开发/测试环境）
# VITE_API_BASE_URL=http://your-server-ip:8080/api/v1
```

注意：
- 前端已优化为使用相对路径，不再硬编码 localhost:8080
- 图片上传路径也已优化，支持任意域名部署
- 推荐使用 Nginx 反向代理，前后端部署在同一域名下

#### 步骤 2：构建前端

```bash
cd vue-tagTime

# 安装依赖
npm install

# 构建生产版本
npm run build
```

构建完成后，会在 `dist/` 目录生成静态文件。

### 2.3 部署方式

#### 方式 1：使用 Nginx（推荐）

安装 Nginx：

```bash
sudo apt install nginx  # Ubuntu/Debian
sudo yum install nginx  # CentOS/RHEL
```

配置 Nginx `/etc/nginx/sites-available/tagtime`：

```nginx
server {
    listen 80;
    server_name your-domain.com;

    # 前端静态文件
    root /path/to/vue-tagTime/dist;
    index index.html;

    # 前端路由支持
    location / {
        try_files $uri $uri/ /index.html;
    }

    # 代理后端 API
    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # WebSocket 支持（如果需要）
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }

    # 静态资源缓存
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
```

启用配置：

```bash
sudo ln -s /etc/nginx/sites-available/tagtime /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

#### 方式 2：使用 HTTPS（推荐生产环境）

使用 Let's Encrypt 免费证书：

```bash
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d your-domain.com
```

Nginx 会自动配置 HTTPS，并添加证书自动续期。

#### 方式 3：直接使用静态文件服务器

```bash
# 使用 serve
npm install -g serve
serve -s dist -l 3000

# 或使用 Python
cd dist
python3 -m http.server 3000
```

---

## 三、配置说明

### 3.1 后端配置项说明

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| SERVER_HOST | 监听地址，0.0.0.0 表示所有接口 | 0.0.0.0 |
| SERVER_PORT | 后端服务端口 | 8080 |
| GIN_MODE | 运行模式：debug/release/test | debug |
| CORS_ALLOWED_ORIGINS | 允许的前端域名，多个用逗号分隔 | 空（允许所有） |
| DB_* | 数据库连接配置 | - |
| REDIS_* | Redis 连接配置 | - |
| AI_* | AI 服务配置 | - |

### 3.2 前端配置项说明

| 配置项 | 说明 |
|--------|------|
| VITE_API_BASE_URL | 后端 API 地址，留空则自动检测 |

### 3.3 CORS 配置说明

- **开发环境**：`CORS_ALLOWED_ORIGINS` 留空，允许所有源访问
- **生产环境**：必须配置具体的前端域名，例如：
  ```env
  CORS_ALLOWED_ORIGINS=https://your-domain.com,https://www.your-domain.com
  ```

---

## 四、部署检查清单

### 4.1 后端检查

- [ ] 数据库已创建并可连接
- [ ] Redis 已安装并运行
- [ ] `.env` 文件已配置正确
- [ ] CORS_ALLOWED_ORIGINS 已配置（生产环境）
- [ ] AI API Key 已配置（如需使用 AI 功能）
- [ ] 后端服务可正常启动
- [ ] 防火墙已开放端口（如 8080）

### 4.2 前端检查

- [ ] `.env.production` 已配置
- [ ] 构建成功，dist 目录存在
- [ ] Nginx 配置正确
- [ ] 前端可正常访问
- [ ] API 请求可正常连接后端

### 4.3 安全检查

- [ ] 数据库密码已修改为强密码
- [ ] Redis 密码已设置
- [ ] API Key 已妥善保管
- [ ] .env 文件未提交到版本控制
- [ ] 生产环境使用 HTTPS
- [ ] 防火墙规则已配置

---

## 五、常见问题

### 5.1 后端无法启动

检查：
1. 数据库连接是否正常
2. Redis 是否运行
3. 端口是否被占用
4. 查看日志文件

### 5.2 前端无法连接后端

检查：
1. 后端服务是否运行
2. CORS 配置是否正确
3. API 地址配置是否正确
4. 防火墙是否开放端口

### 5.3 跨域问题

确保后端 `.env` 中的 `CORS_ALLOWED_ORIGINS` 包含前端域名。

### 5.4 AI 功能不可用

检查：
1. AI_API_KEY 是否配置正确
2. Redis 是否正常运行
3. 网络是否可访问 AI 服务

---

## 六、监控和维护

### 6.1 日志查看

```bash
# systemd 服务日志
sudo journalctl -u tagtime -f

# 或查看日志文件
tail -f tagtime.log
```

### 6.2 性能监控

建议使用：
- Prometheus + Grafana
- 云服务商监控工具

### 6.3 备份

定期备份：
1. 数据库数据
2. 上传的文件（uploads 目录）
3. 配置文件（.env）

---

## 七、快速部署脚本

### 7.1 后端部署脚本（deploy-backend.sh）

```bash
#!/bin/bash
set -e

echo "开始部署后端..."

# 进入后端目录
cd gin-tagTime

# 拉取最新代码
git pull

# 编译
echo "编译中..."
go build -o tagtime-server main.go

# 重启服务
echo "重启服务..."
sudo systemctl restart tagtime

echo "后端部署完成！"
sudo systemctl status tagtime
```

### 7.2 前端部署脚本（deploy-frontend.sh）

```bash
#!/bin/bash
set -e

echo "开始部署前端..."

# 进入前端目录
cd vue-tagTime

# 拉取最新代码
git pull

# 安装依赖
echo "安装依赖..."
npm install

# 构建
echo "构建中..."
npm run build

# 复制到 Nginx 目录
echo "部署到 Nginx..."
sudo rm -rf /var/www/tagtime/*
sudo cp -r dist/* /var/www/tagtime/

# 重启 Nginx
sudo systemctl reload nginx

echo "前端部署完成！"
```

---

## 八、联系支持

如有问题，请查看：
- 项目 README.md
- 后端日志
- Nginx 日志：`/var/log/nginx/error.log`
