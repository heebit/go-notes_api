// models/auth.go
package models

type LoginInput struct {
    Identifier string `json:"identifier" binding:"required"`
    Password   string `json:"password" binding:"required,min=6"` 
}

