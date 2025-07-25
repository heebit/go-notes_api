package controllers_test

import (
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/heebit/notes-api/db"
	"github.com/heebit/notes-api/models"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
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

// registerAndLoginUser - вспомогательная функция для создания тестового пользователя
// и получения его JWT-токена.
func registerAndLoginUser(t *testing.T, testDB *gorm.DB, username, email, password string) (string, uint) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	assert.NoError(t, err) // Проверяем, что хеширование прошло без ошибок

	user := models.User{
		Username: username,
		Email:    email,
		Password: string(hashedPassword),
	}
	result := testDB.Create(&user)  // Создаем пользователя напрямую в БД
	assert.NoError(t, result.Error) // Проверяем, что пользователь успешно создан

	// Генерируем JWT-токен для только что созданного пользователя.
	// Используйте utils.GenerateJWT, если он корректно работает и доступен.
	// Если нет, можно использовать прямую генерацию токена для тестов:
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": float64(user.ID),                      // JWT claims обычно используют float64 для чисел
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Токен действует 24 часа
	})
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	assert.NoError(t, err)

	return tokenString, user.ID
}
