package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username string `json:"username" gorm:"unique"`
	Password string `json:"password"`
	Email    string `json:"email" gorm:"unique"`
	Notes    []Note	`gorm:"foreignKey:UserID"`
} 

type UserSwagger struct {
    ID        uint   `json:"id"`
    Username  string `json:"username"`
    Password  string `json:"password"`
    CreatedAt string `json:"created_at"`
    UpdatedAt string `json:"updated_at"`
}
type MessageResponse struct {
    Message string `json:"message"`
}

// ErrorResponse стандартная ошибка
type ErrorResponse struct {
    Error string `json:"error"`
}