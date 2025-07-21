package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/heebit/notes-api/internal/controllers"
)

func UserRoutes(r *gin.Engine) {
	user := r.Group("/users")
	{
		user.GET("/", controllers.GetUsers)
		user.GET("/:id", controllers.GetUser)
		user.DELETE("/:id", controllers.DeleteUser)
		user.PUT("/:id", controllers.UpdateUser)
	}
}
