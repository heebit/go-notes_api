package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/heebit/notes-api/config"
	"github.com/heebit/notes-api/db"
	"github.com/heebit/notes-api/models"
	"github.com/heebit/notes-api/routes"
)

func main() {
	config.LoadEnv()
	db.Connect()
	
	db.DB = db.DB.Debug()  // Включение режима отладки для GORM, если необходимо
	if err := db.DB.AutoMigrate(&models.User{}, &models.Note{}); err != nil {
    log.Fatalf("Ошибка миграции: %v", err)
}
	r := gin.Default()
	routes.NoteRoutes(r)
	routes.AuthRoutes(r)
	
	r.Run(":8080") // Запуск сервера на порту 8080
}