package config

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	// 从环境变量读取数据库配置，如果没有则使用默认值
	dbUser := getEnv("DB_USER", "root")
	dbPassword := getEnv("DB_PASSWORD", "Root@1234")
	dbHost := getEnv("DB_HOST", "127.0.0.1")
	dbPort := getEnv("DB_PORT", "3306")
	dbName := getEnv("DB_NAME", "tagtime")

	log.Printf("尝试连接数据库: %s@%s:%s/%s", dbUser, dbHost, dbPort, dbName)

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=10s",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Printf("❌ 数据库连接失败: %v", err)
		log.Printf("请检查以下配置:")
		log.Printf("- MySQL服务是否已启动: systemctl status mysql")
		log.Printf("- 数据库用户名: %s", dbUser)
		log.Printf("- 数据库主机: %s:%s", dbHost, dbPort)
		log.Printf("- 数据库名称: %s 是否已创建", dbName)
		log.Printf("- 密码是否正确")
		log.Printf("- 防火墙是否允许连接")
		log.Fatal("无法连接到数据库，程序退出")
	}

	log.Println("✓ 数据库连接成功")
}
