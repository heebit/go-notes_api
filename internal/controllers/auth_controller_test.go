package controllers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/heebit/notes-api/internal/controllers"
	"github.com/heebit/notes-api/models" // Убедитесь, что импортировали models
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)



func TestRegitster(t *testing.T) {
	testDB := setupTestDB()
	defer func() {
		sqlDB, _ := testDB.DB()
		sqlDB.Close()
	}()

	r := gin.Default()
	r.POST("/register", controllers.Register)

	t.Run("Successful Registration", func(t *testing.T) {
		newUser := models.User{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "password123",
		}
		jsonValue, _ := json.Marshal(newUser)
		req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), "Пользователь успешно зарегистрирован")

		var createdUser models.User
		testDB.Where("username = ?", newUser.Username).First(&createdUser)
		assert.Equal(t, newUser.Username, createdUser.Username)
		assert.Equal(t, newUser.Email, createdUser.Email)
		assert.NotEmpty(t, createdUser.Password) // Пароль должен быть захеширован
	})

	t.Run("invalid Input", func(t *testing.T) {
		invalidUser := models.User{
			Username: "inv",           // Слишком короткий username
			Email:    "invalid-email", // Невалидный email
			Password: "123",           // Слишком короткий пароль
		}
		jsonValue, _ := json.Marshal(invalidUser)
		req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Неверный ввод")
	})

	// Добавьте тест для конфликта (пользователь уже существует)
	t.Run("Existing Username or Email", func(t *testing.T) {
		// Сначала зарегистрируем пользователя
		existingUser := models.User{
			Username: "existinguser",
			Email:    "existing@example.com",
			Password: "existingpassword",
		}
		testDB.Create(&existingUser) // Предполагается, что в реальном Register это делается

		// Попытка зарегистрировать пользователя с тем же username
		duplicateUser := models.User{
			Username: "existinguser",
			Email:    "another@example.com", // Новый email, но существующий username
			Password: "newpassword",
		}
		jsonValue, _ := json.Marshal(duplicateUser)
		req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), "Пользователь с таким именем или email уже существует")

		// Попытка зарегистрировать пользователя с тем же email
		duplicateEmailUser := models.User{
			Username: "another_user",
			Email:    "existing@example.com", // Новый username, но существующий email
			Password: "newpassword",
		}
		jsonValue, _ = json.Marshal(duplicateEmailUser)
		req, _ = http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), "Пользователь с таким именем или email уже существует")
	})
}

func TestLogin(t *testing.T) {
	testDB := setupTestDB()
	defer func() {
		sqlDB, _ := testDB.DB()
		sqlDB.Close()
	}()

	// Создаем пользователя для входа (УБЕДИТЕСЬ, что Email валиден!)
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), 14)
	user := models.User{
		Username: "testuser_login",
		Email:    "login@example.com", // <--- Важно: email должен быть валиден!
		Password: string(hashedPassword),
	}
	testDB.Create(&user)

	r := gin.Default()
	r.POST("/login", controllers.Login)

	t.Run("Successful Login by Username", func(t *testing.T) {
		credentials := models.LoginInput{ // <--- Используем LoginInput
			Identifier: "testuser_login",
			Password:   "password123",
		}
		jsonValue, _ := json.Marshal(credentials)
		req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "token")
	})

	t.Run("Successful Login by Email", func(t *testing.T) {
		credentials := models.LoginInput{ // <--- Используем LoginInput
			Identifier: "login@example.com",
			Password:   "password123",
		}
		jsonValue, _ := json.Marshal(credentials)
		req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "token")
	})

	t.Run("Invalid Password", func(t *testing.T) {
		credentials := models.LoginInput{ // <--- Используем LoginInput
			Identifier: "testuser_login",
			Password:   "wrongpassword",
		}
		jsonValue, _ := json.Marshal(credentials)
		req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Неверное имя пользователя или пароль")
	})

	t.Run("User Not Found (Username)", func(t *testing.T) {
		credentials := models.LoginInput{ // <--- Используем LoginInput
			Identifier: "nonexistent_username",
			Password:   "password123",
		}
		jsonValue, _ := json.Marshal(credentials)
		req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Неверное имя пользователя или пароль")
	})

	t.Run("User Not Found (Email)", func(t *testing.T) {
		credentials := models.LoginInput{ // <--- Используем LoginInput
			Identifier: "nonexistent@example.com",
			Password:   "password123",
		}
		jsonValue, _ := json.Marshal(credentials)
		req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Неверное имя пользователя или пароль")
	})

	t.Run("Invalid Input (Missing Identifier)", func(t *testing.T) {
		invalidCredentials := models.LoginInput{ // <--- Используем LoginInput
			Password: "anypassword",
		}
		jsonValue, _ := json.Marshal(invalidCredentials)
		req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Неверный ввод")
	})
	t.Run("Invalid Input (Short Password)", func(t *testing.T) {
		invalidCredentials := models.LoginInput{ // <--- Используем LoginInput
			Identifier: "testuser_login",
			Password:   "123", // Too short
		}
		jsonValue, _ := json.Marshal(invalidCredentials)
		req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Неверный ввод")
	})
}