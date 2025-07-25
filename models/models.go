package models

import (
	"time"

	"gorm.io/gorm"
)

type GormModelSwagger struct {
	ID        uint           `json:"id" example:"1"`
	CreatedAt time.Time      `json:"created_at" example:"2023-01-01T12:00:00Z"`
	UpdatedAt time.Time      `json:"updated_at" example:"2023-01-01T12:00:00Z"`
	DeletedAt gorm.DeletedAt `json:"-" swaggerignore:"true"`
}
