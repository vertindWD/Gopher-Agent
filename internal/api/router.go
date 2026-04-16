package api

import "github.com/gin-gonic/gin"

func SetupRouter() *gin.Engine {
	// 生产环境建议用 gin.ReleaseMode
	r := gin.Default()

	v1 := r.Group("/api/v1")
	{
		v1.POST("/agent/task", SubmitTask)
		v1.GET("/agent/task/:task_id", QueryTask)
	}

	return r
}
