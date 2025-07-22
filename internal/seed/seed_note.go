package seed

import (
	"log"

	"github.com/heebit/notes-api/db"
	"github.com/heebit/notes-api/models"
)

func Load_notes() {
notes := []models.Note{
		{Title: "Первая заметка", Content: "Текст первой заметки", UserID: 1},
		{Title: "Вторая заметка", Content: "Текст второй заметки", UserID: 1},	
		{Title: "Третья заметка", Content: "Текст третьей замтки", UserID: 2},	
	}

	for _, note := range notes {
		var existingNote models.Note
		if err := db.DB.Where("title = ? AND user_id = ?", note.Title, note.UserID).First(&existingNote).Error; err == nil {
			log.Printf("Заметка с заголовком %s для пользователя %d уже существует, пропускаем создание.\n", note.Title, note.UserID)
			continue
		}
		// Создание заметки, если она не существует
		if err := db.DB.Create(&note).Error; err != nil {
			log.Printf("Ошибка при создании заметки %s: %v\n", note.Title, err)
		} else {
			log.Printf("Заметка %s успешно создана для пользователя %d.\n", note.Title, note.UserID)
		}
	}
	
}