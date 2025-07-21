package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/heebit/notes-api/config"
	"github.com/heebit/notes-api/db"
	_ "github.com/heebit/notes-api/docs"
	"github.com/heebit/notes-api/routes"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Notes API
// @version 1.0
// @description API для управления заметками
// @host localhost:8080
// @BasePath /

func main() {
	config.LoadEnv()
	db.Connect()

	defer func() {
		if db.SqlDB != nil {
			if err := db.SqlDB.Close(); err != nil {
				log.Fatalf("Ошибка закрытия базы данных: %v", err)
			}
		}
	}()

	r := gin.Default()

	r.GET("/swagger/*any",
		ginSwagger.WrapHandler(
			swaggerFiles.Handler,
			ginSwagger.PersistAuthorization(true),
		),
	)

	routes.NoteRoutes(r)
	routes.AuthRoutes(r)

	r.Run(":8080") // Запуск сервера на порту 8080

}
