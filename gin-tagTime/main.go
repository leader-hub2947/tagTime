package main

import (
	"fmt"
	"log"
	"os"
	"tagtime/config"
	"tagtime/models"
	"tagtime/routes"
	"tagtime/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 强制输出到标准输出，确保日志可见
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags)

	fmt.Println("========================================")
	fmt.Println("TagTime 后端服务启动中...")
	fmt.Println("========================================")

	// 加载 .env 文件
	fmt.Println("\n[1/6] 加载环境配置...")
	if err := godotenv.Load(); err != nil {
		log.Println("⚠ 警告: 未找到 .env 文件，将使用默认配置或系统环境变量")
	} else {
		log.Println("✓ 成功加载 .env 文件")
	}

	// 验证 AI 配置（调试用）
	fmt.Println("\n[2/6] 验证 AI 配置...")
	log.Printf("  - AI Provider: %s", os.Getenv("AI_PROVIDER"))
	log.Printf("  - AI Endpoint: %s", os.Getenv("AI_ENDPOINT"))
	log.Printf("  - AI Model: %s", os.Getenv("AI_MODEL"))
	apiKey := os.Getenv("AI_API_KEY")
	if apiKey != "" && apiKey != "请替换为你的DeepSeek_API_Key" {
		log.Printf("  ✓ AI API Key 已配置（长度: %d）", len(apiKey))
	} else {
		log.Println("  ⚠ 警告: AI API Key 未配置或使用默认值，AI 功能将不可用")
	}

	// 初始化数据库
	fmt.Println("\n[3/6] 连接数据库...")
	config.InitDB()

	// 自动迁移数据库表
	fmt.Println("\n[4/6] 执行数据库迁移...")
	if err := models.AutoMigrate(config.DB); err != nil {
		log.Fatal("❌ 数据库迁移失败:", err)
	}
	log.Println("✓ 数据库迁移完成")

	// 初始化Redis
	fmt.Println("\n[5/6] 连接 Redis...")
	if err := config.InitRedis(); err != nil {
		log.Printf("⚠ Redis初始化失败（AI功能将不可用）: %v", err)
	} else {
		defer config.CloseRedis()
		log.Println("✓ Redis连接成功")
	}

	// 启动定时任务调度器
	fmt.Println("\n[6/6] 启动定时任务调度器...")
	utils.StartScheduler()
	defer utils.StopScheduler()
	log.Println("✓ 定时任务调度器已启动")

	// 创建 Gin 引擎
	fmt.Println("\n初始化 Web 服务...")

	// 根据环境设置 Gin 模式
	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		ginMode = "debug"
	}
	gin.SetMode(ginMode)

	r := gin.Default()

	// 配置 CORS - 从环境变量读取允许的源
	allowedOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	var corsConfig cors.Config

	if allowedOrigins != "" {
		// 生产环境：使用配置的允许源
		origins := []string{}
		for _, origin := range splitAndTrim(allowedOrigins, ",") {
			origins = append(origins, origin)
		}
		corsConfig = cors.Config{
			AllowOrigins:     origins,
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
		}
		log.Printf("✓ CORS 配置: 允许的源 = %v", origins)
	} else {
		// 开发环境：允许所有源
		corsConfig = cors.Config{
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
			AllowOriginFunc: func(origin string) bool {
				return true // 开发环境允许所有源
			},
		}
		log.Println("✓ CORS 配置: 允许所有源（开发模式）")
	}

	r.Use(cors.New(corsConfig))

	// 静态文件服务（用于访问上传的图片）
	r.Static("/uploads", "./uploads")

	// 注册路由
	routes.SetupRoutes(r)

	// 从环境变量读取服务器配置
	serverHost := os.Getenv("SERVER_HOST")
	if serverHost == "" {
		serverHost = "0.0.0.0" // 默认监听所有网络接口
	}

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8080" // 默认端口
	}

	serverAddr := fmt.Sprintf("%s:%s", serverHost, serverPort)

	fmt.Println("\n========================================")
	fmt.Println("✓ 服务器启动成功！")
	fmt.Println("========================================")
	fmt.Printf("运行模式: %s\n", ginMode)
	fmt.Printf("监听地址: %s\n", serverAddr)
	fmt.Printf("本地访问: http://localhost:%s\n", serverPort)
	fmt.Println("局域网/外网访问: http://<你的IP>:" + serverPort)

	if err := r.Run(serverAddr); err != nil {
		log.Fatal("❌ 服务器启动失败:", err)
	}
}

// splitAndTrim 分割字符串并去除空格
func splitAndTrim(s, sep string) []string {
	parts := []string{}
	for _, part := range splitString(s, sep) {
		trimmed := trimSpace(part)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

func splitString(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	result := []string{}
	current := ""
	for _, char := range s {
		if string(char) == sep {
			result = append(result, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	result = append(result, current)
	return result
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}
