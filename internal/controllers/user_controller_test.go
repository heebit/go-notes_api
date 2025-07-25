package controllers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/heebit/notes-api/internal/controllers"
	"github.com/heebit/notes-api/middleware"
	"github.com/heebit/notes-api/models"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func TestUserController(t *testing.T) {
	testDB := setupTestDB()
	defer func() {
		sqlDB, _ := testDB.DB()
		sqlDB.Close()
	}()

	// Инициализируем роутер с middleware
	r := gin.Default()
	userRoutes := r.Group("/users")
	userRoutes.Use(middleware.AuthMiddleware())
	{
		userRoutes.GET("/:id", controllers.GetUser)
		userRoutes.PUT("/:id", controllers.UpdateUser)
		userRoutes.DELETE("/:id", controllers.DeleteUser)
		userRoutes.PUT("/me/password", controllers.ChangePassword)
	}

	// Создаем двух тестовых пользователей
	token, userID := registerAndLoginUser(t, testDB, "user1", "user1@example.com", "password123")
	_ , userID2 := registerAndLoginUser(t, testDB, "user2", "user2@example.com", "password123")

	// Переменная для хранения последнего хеша пароля user1
	var initialHashedPassword string
	testDB.First(&models.User{}, userID).Scan(&models.User{Password: initialHashedPassword})

	t.Run("GetUser - Successful", func(t *testing.T) {
		t.Log("Запуск: GetUser - Успешное получение своего профиля")
		req, _ := http.NewRequest(http.MethodGet, "/users/"+strconv.FormatUint(uint64(userID), 10), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var user models.UserSwagger
		json.Unmarshal(w.Body.Bytes(), &user)
		assert.Equal(t, userID, user.ID)
		assert.Equal(t, "user1", user.Username)
		assert.Equal(t, "user1@example.com", user.Email)
	})

	t.Run("GetUser - Unauthorized Access", func(t *testing.T) {
		t.Log("Запуск: GetUser - Неавторизованный доступ")
		req, _ := http.NewRequest(http.MethodGet, "/users/"+strconv.FormatUint(uint64(userID), 10), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("GetUser - Not Found or Not Owned", func(t *testing.T) {
		t.Log("Запуск: GetUser - Попытка получить профиль другого пользователя")
		req, _ := http.NewRequest(http.MethodGet, "/users/"+strconv.FormatUint(uint64(userID2), 10), nil) // Пробуем получить user2
		req.Header.Set("Authorization", "Bearer "+token)                                                  // С токеном user1
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Пользователь не найден")
	})

	t.Run("UpdateUser - Successful", func(t *testing.T) {
		t.Log("Запуск: UpdateUser - Успешное обновление своего профиля")
		updatedUserInput := models.UpdateUserInput{
			Username: "updated_user1",
			Email:    "updated_user1@example.com",
		}
		jsonValue, _ := json.Marshal(updatedUserInput)
		req, _ := http.NewRequest(http.MethodPut, "/users/"+strconv.FormatUint(uint64(userID), 10), bytes.NewBuffer(jsonValue))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var user models.UserSwagger
		json.Unmarshal(w.Body.Bytes(), &user)
		assert.Equal(t, "updated_user1", user.Username)
		assert.Equal(t, "updated_user1@example.com", user.Email)

		// Проверяем, что пароль не изменился
		var dbUser models.User
		testDB.First(&dbUser, userID)
		assert.NotEqual(t, "", dbUser.Password)
		assert.Equal(t, true, bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte("password123")) == nil)
	})

	t.Run("UpdateUser - Not Found or Not Owned", func(t *testing.T) {
		t.Log("Запуск: UpdateUser - Попытка обновить профиль другого пользователя")
		updatedUserInput := models.UpdateUserInput{
			Username: "hacker_name",
		}
		jsonValue, _ := json.Marshal(updatedUserInput)
		req, _ := http.NewRequest(http.MethodPut, "/users/"+strconv.FormatUint(uint64(userID2), 10), bytes.NewBuffer(jsonValue)) // Пробуем обновить user2
		req.Header.Set("Authorization", "Bearer "+token)                                                                         // С токеном user1
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Пользователь не найден")
	})

	t.Run("ChangePassword - Successful", func(t *testing.T) {
		t.Log("Запуск: ChangePassword - Успешная смена пароля")
		changePasswordInput := models.ChangePasswordInput{
			OldPassword: "password123",
			NewPassword: "new_password_strong",
		}
		jsonValue, _ := json.Marshal(changePasswordInput)
		req, _ := http.NewRequest(http.MethodPut, "/users/me/password", bytes.NewBuffer(jsonValue))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Пароль успешно изменен")

		// Проверяем, что пароль в БД действительно изменился
		var dbUser models.User
		testDB.First(&dbUser, userID)
		assert.True(t, bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte("new_password_strong")) == nil)
		assert.False(t, bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte("password123")) == nil)
	})

	t.Run("ChangePassword - Invalid Old Password", func(t *testing.T) {
		t.Log("Запуск: ChangePassword - Неверный старый пароль")
		changePasswordInput := models.ChangePasswordInput{
			OldPassword: "wrong_password",
			NewPassword: "new_password_strong",
		}
		jsonValue, _ := json.Marshal(changePasswordInput)
		req, _ := http.NewRequest(http.MethodPut, "/users/me/password", bytes.NewBuffer(jsonValue))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Неверный старый пароль")

		// Проверяем, что пароль в БД не изменился
		var dbUser models.User
		testDB.First(&dbUser, userID)
		assert.True(t, bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte("new_password_strong")) == nil)
	})

	t.Run("DeleteUser - Successful", func(t *testing.T) {
		t.Log("Запуск: DeleteUser - Успешное удаление своего профиля")
		req, _ := http.NewRequest(http.MethodDelete, "/users/"+strconv.FormatUint(uint64(userID), 10), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Пользователь успешно удален")

		// Проверяем, что пользователь действительно удален
		var deletedUser models.User
		result := testDB.First(&deletedUser, userID)
		assert.Error(t, result.Error)
		assert.Equal(t, gorm.ErrRecordNotFound, result.Error)
	})

	t.Run("DeleteUser - Not Found or Not Owned", func(t *testing.T) {
		t.Log("Запуск: DeleteUser - Попытка удалить профиль другого пользователя")
		req, _ := http.NewRequest(http.MethodDelete, "/users/"+strconv.FormatUint(uint64(userID2), 10), nil)
		req.Header.Set("Authorization", "Bearer "+token) // Используем токен уже удаленного пользователя
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Пользователь не найден")
	})
}
