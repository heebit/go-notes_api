package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/heebit/notes-api/internal/controllers"
)

func NoteRoutes(r *gin.Engine) {
	note := r.Group("/notes")
	{
		note.GET("/", controllers.GetNotes)
		note.POST("/", controllers.CreateNote)
		note.DELETE("/:id", controllers.DeleteNote)
		note.PUT("/:id", controllers.UpdateNote)
	}
}
