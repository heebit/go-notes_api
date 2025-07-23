package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username string `json:"username" gorm:"unique" binding:"required,min=3,max=30"`
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email" gorm:"unique" binding:"required,email" `
	Notes    []Note	`gorm:"foreignKey:UserID"`
} 

type UserSwagger struct {
    ID        uint   `json:"id"`
    Username  string `json:"username"`
    Password  string `json:"password"`
    CreatedAt string `json:"created_at"`
    UpdatedAt string `json:"updated_at"`
}
