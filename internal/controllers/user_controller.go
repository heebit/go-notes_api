package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/heebit/notes-api/db"
	"github.com/heebit/notes-api/models"
)

// GetUsers godoc
// @Summary Получить всех пользователей
// @Tags users
// @Produce json
// @Success 200 {array} models.UserSwagger
// @Router /users [get]
func GetUsers(c *gin.Context) {
	var users []models.User
	db.DB.Find(&users)
	c.JSON(http.StatusOK, users)
}

// GetUser godoc
// @Summary Получить пользователя по ID
// @Description Возвращает информацию о пользователе по ID
// @Tags users
// @Param id path int true "User ID"
// @Success 200 {object} models.UserSwagger
// @Failure 400 {object} models.ErrorResponse 
// @Router /users/{id} [get]

func GetUser(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := db.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// DeleteUser godoc
// @Summary Удалить пользователя
// @Tags users
// @Param id path int true "User ID"
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.ErrorResponse 
// @Router /users/{id} [delete]
func DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if err := db.DB.Delete(&models.User{}, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
	}
	c.JSON(http.StatusOK, gin.H{"message": "Пользователь успешно удален"})	
}

// UpdateUser godoc
// @Summary Обновить пользователя
// @Description Обновляет информацию о пользователе по ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param user body models.UserSwagger true "Данные пользователя"
// @Success 200 {object} models.UserSwagger
// @Failure 400 {object} models.ErrorResponse
// @Router /users/{id} [put]
func UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := db.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		return
	}

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных	"})
		return
	}

	db.DB.Save(&user)
	c.JSON(http.StatusOK, user)
}
