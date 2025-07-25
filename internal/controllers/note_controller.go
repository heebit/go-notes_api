package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/heebit/notes-api/db"
	"github.com/heebit/notes-api/models"
	"gorm.io/gorm"
)

func getUserIdFromContext(c *gin.Context) (uint, bool) {
	userId, exists := c.Get("user_id")
	if !exists {
		c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponse{Error: "UserID не найден в контексте (middleware issue)"})
		return 0, false
	}
	id, ok := userId.(uint)
	if !ok {
		c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Неверный формат UserID в контексте (type assertion issue)"})
		return 0, false
	}
	return id, true
}

// GetNotes godoc
// @Summary Получить все заметки
// @Description Возвращает список всех заметок
// @Tags notes
// @Produce json
// @Success 200 {array} models.NoteSwagger
// @Router /notes [get]
func GetNotes(c *gin.Context) {
	userId, ok := getUserIdFromContext(c)
	if !ok {
		return // Ошибка уже обработана в getUserIdFromContext
	}

	var notes []models.Note
	db.DB.Where("user_id = ?", userId).Find(&notes)
	c.JSON(http.StatusOK, notes)
}

// @Summary Получить заметку по ID
// @Description Возвращает заметку по её ID, если она принадлежит текущему пользователю
// @Tags notes
// @Accept  json
// @Produce  json
// @Param id path int true "ID заметки"
// @Security ApiKeyAuth
// @Success 200 {object} models.Note "Успешный запрос"
// @Failure 400 {object} models.ErrorResponse "Неверный формат ID"
// @Failure 404 {object} models.ErrorResponse "Заметка не найдена или не принадлежит пользователю"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /notes/{id} [get]

func GetNote(c *gin.Context) {
	userId, ok := getUserIdFromContext(c)
	if !ok {
		return // Ошибка уже обработана в getUserIdFromContext
	}
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, models.ErrorResponse{Error: "Неверный формат ID заметки"})
		return
	}
	var note models.Note
	result := db.DB.Where("id = ? AND user_id = ?", uint(id), userId).First(&note)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.AbortWithStatusJSON(http.StatusNotFound, models.ErrorResponse{Error: "Заметка не найдена или не принадлежит вам"})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Ошибка при поиске заметки"})
		return
	}
	c.JSON(http.StatusOK, note)
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
	userID, ok := getUserIdFromContext(c)
	if !ok {
		return
	}
	var note models.Note
	if err := c.ShouldBindJSON(&note); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, models.ErrorResponse{Error: "Неверный ввод: " + err.Error()})
		return
	}
	note.UserID = userID
	db.DB.Create(&note)
	c.JSON(http.StatusCreated, note)
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
    userID, ok := getUserIdFromContext(c)
    if !ok { return }

    idStr := c.Param("id")
    id, err := strconv.ParseUint(idStr, 10, 32)
    if err != nil {
        c.AbortWithStatusJSON(http.StatusBadRequest, models.ErrorResponse{Error: "Неверный формат ID заметки"})
        return
    }

    var inputNote models.Note
    if err := c.ShouldBindJSON(&inputNote); err != nil {
        c.AbortWithStatusJSON(http.StatusBadRequest, models.ErrorResponse{Error: "Неверный ввод: " + err.Error()})
        return
    }

    var existingNote models.Note
    result := db.DB.Where("id = ? AND user_id = ?", uint(id), userID).First(&existingNote)
    if result.Error != nil {
        if result.Error == gorm.ErrRecordNotFound {
            c.AbortWithStatusJSON(http.StatusNotFound, models.ErrorResponse{Error: "Заметка не найдена или не принадлежит вам"})
            return
        }
        c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Ошибка при поиске заметки"})
        return
    }

    existingNote.Title = inputNote.Title
    existingNote.Content = inputNote.Content

    db.DB.Save(&existingNote)
    c.JSON(http.StatusOK, existingNote)
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
    userID, ok := getUserIdFromContext(c)
    if !ok { return }

    idStr := c.Param("id")
    id, err := strconv.ParseUint(idStr, 10, 32)
    if err != nil {
        c.AbortWithStatusJSON(http.StatusBadRequest, models.ErrorResponse{Error: "Неверный формат ID заметки"})
        return
    }

    result := db.DB.Where("id = ? AND user_id = ?", uint(id), userID).Delete(&models.Note{})

    if result.Error != nil {
        c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Ошибка при удалении заметки"})
        return
    }
    if result.RowsAffected == 0 {
        c.AbortWithStatusJSON(http.StatusNotFound, models.ErrorResponse{Error: "Заметка не найдена или не принадлежит вам"})
        return
    }

    c.JSON(http.StatusOK, models.MessageResponse{Message: "Заметка успешно удалена"})
}

