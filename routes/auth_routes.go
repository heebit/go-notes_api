package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/heebit/notes-api/internal/controllers"
)

func AuthRoutes(r *gin.Engine) {
	r.POST("/register", controllers.Register)
	r.POST("/login", controllers.Login)
}