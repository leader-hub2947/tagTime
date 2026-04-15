package routes

import (
	"tagtime/config"
	"tagtime/controllers"
	"tagtime/middleware"
	"tagtime/services"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")

	// 认证路由
	auth := api.Group("/auth")
	{
		auth.POST("/register", controllers.Register)
		auth.POST("/login", controllers.Login)
	}

	// 需要认证的路由
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware())
	{
		// 标签
		tags := protected.Group("/tags")
		{
			tags.GET("", controllers.GetTags)
			tags.POST("", controllers.CreateTag)
			tags.PUT("/:id", controllers.UpdateTag)
			tags.DELETE("/:id", controllers.DeleteTag)
		}

		// 便签
		notes := protected.Group("/notes")
		{
			notes.GET("", controllers.GetNotes)
			notes.GET("/:id", controllers.GetNote)
			notes.POST("", controllers.CreateNote)
			notes.PUT("/:id", controllers.UpdateNote)
			notes.DELETE("/:id", controllers.DeleteNote)
			notes.POST("/:id/restore", controllers.RestoreNote)
			notes.POST("/trash/empty", controllers.EmptyTrash)
			notes.GET("/calendar/:year/:month", controllers.GetNoteCalendar)
			notes.POST("/upload-image", controllers.UploadImage)
			notes.POST("/delete-image", controllers.DeleteImage)
		}

		// 任务
		tasks := protected.Group("/tasks")
		{
			tasks.GET("", controllers.GetTasks)
			tasks.GET("/:id", controllers.GetTask)
			tasks.POST("", controllers.CreateTask)
			tasks.PUT("/:id", controllers.UpdateTask)
			tasks.DELETE("/:id", controllers.DeleteTask)
			tasks.POST("/:id/start", controllers.StartTimer)
			tasks.POST("/:id/complete", controllers.CompleteTask)
			tasks.POST("/:id/archive", controllers.ArchiveTask)
			tasks.POST("/:id/unarchive", controllers.UnarchiveTask)
			tasks.GET("/archived", controllers.GetArchivedTasks)
			tasks.GET("/:id/notes", controllers.GetTaskNotes)
		}

		// 计时
		timer := protected.Group("/timer")
		{
			timer.POST("/pause", controllers.PauseTimer)
			timer.POST("/resume", controllers.ResumeTimer)
			timer.POST("/end", controllers.EndTimer)
			timer.GET("/current", controllers.GetCurrentTimer)
			timer.POST("/switch", controllers.SwitchTimer)
		}

		// 数据可视化
		dashboard := protected.Group("/dashboard")
		{
			dashboard.GET("/timeline", controllers.GetTimeline)
			dashboard.GET("/tag-ranking", controllers.GetTagRanking)
			dashboard.GET("/task-statistics", controllers.GetTaskStatistics)
		}

		// 用户设置
		settings := protected.Group("/settings")
		{
			settings.GET("", controllers.GetSettings)
			settings.PUT("", controllers.UpdateSettings)
		}

		// AI洞察 - 用一句话击溃我
		if config.RedisClient != nil {
			aiConfig := config.LoadAIConfig()
			aiService := services.NewAICrushService(config.DB, config.RedisClient, aiConfig)
			aiController := controllers.NewAICrushController(aiService)

			ai := protected.Group("/ai")
			{
				ai.POST("/crush", aiController.GetCrushLine)
				ai.GET("/crush/remaining", aiController.GetRemainingCount)
			}
		}
	}
}
