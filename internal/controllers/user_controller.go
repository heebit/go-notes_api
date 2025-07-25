package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/heebit/notes-api/db"
	"github.com/heebit/notes-api/models"
	"golang.org/x/crypto/bcrypt"
)

// --- Вспомогательная функция для получения ID пользователя из токена ---
// Она поможет избежать дублирования кода в каждом контроллере.
func getValidatedUserID(c *gin.Context, paramID string) (uint, bool) {
	// Получаем ID пользователя из токена. В middleware мы положили туда uint
	userIDFromToken, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	
	// Конвертируем ID из URL-параметра в uint64, а затем в uint
	paramIDUint64, err := strconv.ParseUint(paramID, 10, 64)
	if err != nil {
		return 0, false
	}
	paramIDUint := uint(paramIDUint64)

	// Сравниваем ID из токена (uint) с ID из URL-параметра (uint)
	// Здесь мы просто приводим userIDFromToken к типу uint
	if userIDFromToken.(uint) != paramIDUint {
		return 0, false
	}

	return paramIDUint, true
}


// GetUsers godoc
// @Summary Получить всех пользователей
// @Tags users
// @Produce json
// @Success 200 {array} models.UserSwagger
// @Router /users [get]
func GetUsers(c *gin.Context) {
	var users []models.User
	db.DB.Find(&users)

	// Важно: не возвращаем поле Password!
	var usersPublic []models.UserSwagger
	for _, user := range users {
		usersPublic = append(usersPublic, models.UserSwagger{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt.String(),
			UpdatedAt: user.UpdatedAt.String(),
		})
	}
	c.JSON(http.StatusOK, usersPublic)
}

// GetUser godoc
// @Summary Получить пользователя по ID
// @Description Возвращает информацию о пользователе по ID. Доступно только для владельца токена.
// @Tags users
// @Param id path int true "User ID"
// @Success 200 {object} models.UserSwagger
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /users/{id} [get]
func GetUser(c *gin.Context) {
	id := c.Param("id")

	// Проверяем, что запрашиваемый ID принадлежит текущему пользователю
	if _, ok := getValidatedUserID(c, id); !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		return
	}

	var user models.User
	if err := db.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		return
	}

	// Возвращаем только безопасную информацию
	c.JSON(http.StatusOK, models.UserSwagger{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.String(),
		UpdatedAt: user.UpdatedAt.String(),
	})
}

// UpdateUser godoc
// @Summary Обновить пользователя
// @Description Обновляет информацию о пользователе по ID. Доступно только для владельца токена.
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param user body models.UpdateUserInput true "Данные пользователя для обновления"
// @Success 200 {object} models.UserSwagger
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /users/{id} [put]
func UpdateUser(c *gin.Context) {
	id := c.Param("id")

	// Проверяем, что запрашиваемый ID принадлежит текущему пользователю
	if _, ok := getValidatedUserID(c, id); !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		return
	}

	var user models.User
	if err := db.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		return
	}

	var input models.UpdateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных: " + err.Error()})
		return
	}

	// Обновляем поля, которые пришли в input, не трогая пароль
	db.DB.Model(&user).Updates(&input)

	// Возвращаем обновленные данные, исключая пароль
	c.JSON(http.StatusOK, models.UserSwagger{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.String(),
		UpdatedAt: user.UpdatedAt.String(),
	})
}

// DeleteUser godoc
// @Summary Удалить пользователя
// @Description Удаляет пользователя по ID. Доступно только для владельца токена.
// @Tags users
// @Param id path int true "User ID"
// @Success 200 {object} models.MessageResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /users/{id} [delete]
func DeleteUser(c *gin.Context) {
	id := c.Param("id")

	// Проверяем, что запрашиваемый ID принадлежит текущему пользователю
	if _, ok := getValidatedUserID(c, id); !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		return
	}

	result := db.DB.Delete(&models.User{}, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Пользователь успешно удален"})
}

// ChangePassword godoc
// @Summary Сменить пароль
// @Description Смена пароля для текущего авторизованного пользователя
// @Tags users
// @Accept json
// @Produce json
// @Param password body models.ChangePasswordInput true "Старый и новый пароли"
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /users/me/password [put]
func ChangePassword(c *gin.Context) {
	userIDFromToken, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
        return
    }

	var input models.ChangePasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	// 1. Найти пользователя по ID
	 var user models.User
    if err := db.DB.First(&user, userIDFromToken.(uint)).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
        return
    }

	// 2. Проверить старый пароль
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.OldPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный старый пароль"})
		return
	}

	// 3. Захешировать новый пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось захешировать пароль"})
		return
	}

	// 4. Обновить пароль в БД
	db.DB.Model(&user).Update("password", string(hashedPassword))

	c.JSON(http.StatusOK, gin.H{"message": "Пароль успешно изменен"})
}
