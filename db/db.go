package db

import (
	"database/sql"
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	DB    *gorm.DB
	SqlDB *sql.DB
)

func Connect() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"))

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Errorf("ошибка подключения к базе данных: %w", err))
	}
	DB = db

	if sqlDB, err := db.DB(); err != nil {
		panic(fmt.Errorf("ошибка получения SQL DB: %w", err))
	} else {
		SqlDB = sqlDB

		fmt.Println("Успешное подключение к базе данных")
	}
}
