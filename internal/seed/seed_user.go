package seed

import (
	"log"

	"github.com/heebit/notes-api/db"
	"github.com/heebit/notes-api/models"
	"golang.org/x/crypto/bcrypt"
)

func hashPassword(password string) string {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Fatal(err)
	}
	return string(hashed)
}

func Load_users() {
	users := []models.User{
		{Username: "testuser1", Email: "test1@example.com", Password: hashPassword("password123")},
		{Username: "testuser2", Email: "test2@example.com", Password: hashPassword("password456")},
	}

	for _, user := range users {
		var existingUser models.User
		if err := db.DB.Where("username = ? OR email = ?", user.Username, user.Email).First(&existingUser).Error; err == nil {
			log.Printf("Пользователь %s или %s уже существует, пропускаем создание.\n", user.Username, user.Email)
			continue
		}
		if err := db.DB.Create(&user).Error; err != nil {
			log.Printf("Ошибка при создании пользователя %s: %v\n", user.Username, err)
		}else{
			log.Printf("Пользователь %s успешно создан.\n", user.Username)
		}
	}

}
