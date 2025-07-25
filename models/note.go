package models

type Note struct {
	GormModelSwagger
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
	UserID  uint   `json:"user_id"`
}

type NoteSwagger struct {
	ID      uint   `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
	UserID  uint   `json:"user_id"`
}
