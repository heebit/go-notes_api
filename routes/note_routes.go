package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/heebit/notes-api/internal/controllers"
	"github.com/heebit/notes-api/middleware"
)

func NoteRoutes(r *gin.Engine) {
	note := r.Group("/notes").Use(middleware.AuthMiddleware())
	{
		note.GET("/", controllers.GetNotes)
		note.GET("/:id", controllers.GetNote)
		note.POST("/", controllers.CreateNote)
		note.DELETE("/:id", controllers.DeleteNote)
		note.PUT("/:id", controllers.UpdateNote)
	}
}
