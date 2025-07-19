package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/heebit/notes-api/internal/db"
	"github.com/heebit/notes-api/models"
)

func GetNotes(c *gin.Context) {
	var notes []models.Note
	db.DB.Find(&notes)
	c.JSON(http.StatusOK, notes)
}

func CreateNote(c *gin.Context) {
	var note models.Note
	if err := c.ShouldBindJSON(&note); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Create(&note)
	c.JSON(http.StatusOK, note)
}
func DeleteNote(c *gin.Context) {
	id := c.Param("id")
	db.DB.Delete(&models.Note{}, id)
}

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