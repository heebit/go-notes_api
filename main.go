package main

import (
	"github.com/gin-gonic/gin"
	"github.com/heebit/notes-api/internal/config"
	"github.com/heebit/notes-api/internal/db"
	"github.com/heebit/notes-api/internal/routes"
	"github.com/heebit/notes-api/models"
)

func main() {
	config.LoadEnv()
	db.Connect()
	
	db.DB.AutoMigrate(&models.Note{})
	r := gin.Default()
	routes.NoteRoutes(r)
	
	r.Run(":8080") // Запуск сервера на порту 8080
}