package controllers_test

import (
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/heebit/notes-api/db"
	"github.com/heebit/notes-api/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	gin.SetMode(gin.TestMode)

	testDB, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("Не удалось подключиться к тестовой базе данных")
	}
	if err := testDB.AutoMigrate(&models.User{}, &models.Note{}); err != nil {
		panic("Не удалось выполнить миграцию тестовой базы данных")
	}
	db.DB = testDB
	return testDB
}

func TestMain(m *testing.M) {
    // Установите тестовые переменные среды
    os.Setenv("JWT_SECRET", "supersecretkeyfortesting")
    
    // Настройте тестовую базу данных
    gin.SetMode(gin.TestMode)
    testDB, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
    if err != nil {
        panic("Не удалось подключиться к тестовой базе данных")
    }
    if err := testDB.AutoMigrate(&models.User{}, &models.Note{}); err != nil {
        panic("Не удалось выполнить миграцию тестовой базы данных")
    }
    db.DB = testDB
    
    // Запустите тесты
    code := m.Run()
    
    // Очистка
    os.Unsetenv("JWT_SECRET")
    os.Exit(code)
}

// func TestMain(m *testing.M) {
//     // Установите тестовые переменные среды
//     os.Setenv("JWT_SECRET", "supersecretkeyfortesting")

//     // В TestMain только установка глобальных переменных среды и общих настроек,
//     // не связанных с состоянием БД между подтестами.
//     // Запуск gin.SetMode(gin.TestMode) здесь тоже хороший тон.
//     gin.SetMode(gin.TestMode)

//     // Запустите тесты
//     code := m.Run()

//     // Очистка
//     os.Unsetenv("JWT_SECRET")
//     os.Exit(code)
// }
