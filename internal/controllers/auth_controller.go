package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/heebit/notes-api/db"
	"github.com/heebit/notes-api/models"
	"github.com/heebit/notes-api/utils"
	"golang.org/x/crypto/bcrypt"
)

func Register(c *gin.Context) {
	var input models.User
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ввод"})
		return
	}

	if hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), 14); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка хеширования пароля"})
		return
	} else {
		input.Password = string(hashed)
		db.DB.Create(&input)
		c.JSON(http.StatusCreated, gin.H{"message": "Пользователь успешно зарегистрирован"})
	}
}

func Login(c *gin.Context) {
	var input models.User
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ввод"})
		return
	}

	var user models.User
	db.DB.Where("username = ?", input.Username).First(&user)

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверное имя пользователя или пароль"})
		return
	}

	if token, err := utils.GenerateJWT(user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка генерации токена"})
	} else {
		c.JSON(http.StatusOK, gin.H{"token": token})
	}
}
// Register создает нового пользователя, хешируя его пароль перед сохранением в базу данных.
// Login проверяет введенные учетные данные пользователя и, если они верны, генерирует JWT токен для авторизации.
// Оба метода используют модели и функции из пакета db для взаимодействия с базой данных и utils для генерации JWT токена.
// Эти функции должны быть защищены от SQL-инъекций и других уязвимостей, что достигается использованием ORM GORM и безопасных методов хеширования паролей.
