package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/heebit/notes-api/internal/controllers"
)

func AuthRoutes(r *gin.Engine) {
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/register", controllers.Register)
		authGroup.POST("/login", controllers.Login)
	}
}
