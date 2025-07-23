// models/auth.go
package models

type LoginInput struct {
    Identifier string `json:"identifier" binding:"required"`
    Password   string `json:"password" binding:"required,min=6"` 
}

type MessageResponse struct {
    Message string `json:"message"`
}

type ErrorResponse struct {
    Error string `json:"error"`
}