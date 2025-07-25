package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username string `json:"username" gorm:"unique" binding:"required,min=3,max=30"`
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email" gorm:"unique" binding:"required,email" `
	Notes    []Note `gorm:"foreignKey:UserID"`
}

type UserSwagger struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type UpdateUserInput struct {
	Username string `json:"username,omitempty" binding:"omitempty,min=3,max=30"`
	Email    string `json:"email,omitempty" binding:"omitempty,email"`
}

type ChangePasswordInput struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}
