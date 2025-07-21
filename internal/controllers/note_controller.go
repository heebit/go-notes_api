package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/heebit/notes-api/db"
	"github.com/heebit/notes-api/models"
)

// GetNotes godoc
// @Summary Получить все заметки
// @Description Возвращает список всех заметок
// @Tags notes
// @Produce json
// @Success 200 {array} models.NoteSwagger
// @Router /notes [get]
func GetNotes(c *gin.Context) {
	var notes []models.Note
	db.DB.Find(&notes)
	c.JSON(http.StatusOK, notes)
}

// CreateNote godoc
// @Summary Создать новую заметку
// @Description Создает новую заметку по переданным данным
// @Tags notes
// @Accept json
// @Produce json
// @Param note body models.NoteSwagger true "Данные заметки"
// @Success 200 {object} models.NoteSwagger
// @Failure 400 {object} models.ErrorResponse 
// @Router /notes [post]
func CreateNote(c *gin.Context) {
	var note models.Note
	if err := c.ShouldBindJSON(&note); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Create(&note)
	c.JSON(http.StatusOK, note)
}

// DeleteNote godoc
// @Summary Удалить заметку
// @Description Удаляет заметку по ID
// @Tags notes
// @Produce json
// @Param id path int true "ID заметки"
// @Success 200 {object} models.MessageResponse
// @Router /notes/{id} [delete]
func DeleteNote(c *gin.Context) {
	id := c.Param("id")
	db.DB.Delete(&models.Note{}, id)
}

// UpdateNote godoc
// @Summary Обновить заметку
// @Description Обновляет заметку по ID
// @Tags notes
// @Accept json
// @Produce json
// @Param id path int true "ID заметки"
// @Param note body models.Note true "Обновленные данные"
// @Param note body models.NoteSwagger true "Обновленные данные"
// @Success 200 {object} models.NoteSwagger
// @Failure 400 {object} models.ErrorResponse
// @Router /notes/{id} [put]
func UpdateNote(c *gin.Context) {
	id := c.Param("id")
	var note models.Note
	if err := c.ShouldBindJSON(&note); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Ошибка": err.Error()})
		return
	}
	
	db.DB.Model(&models.Note{}).Where("id = ?", id).Updates(note)
	c.JSON(http.StatusOK, note)
}