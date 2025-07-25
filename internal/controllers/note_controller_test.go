package controllers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv" // Для strconv.FormatUint
	"testing"
	"time" // Для jwt.MapClaims в registerAndLoginUser

	"github.com/gin-gonic/gin"
	"github.com/heebit/notes-api/internal/controllers"
	"github.com/heebit/notes-api/middleware" // Путь к вашему middleware
	"github.com/heebit/notes-api/models"

	"github.com/golang-jwt/jwt/v5" // Добавлен для ручной генерации JWT в registerAndLoginUser, если utils.GenerateJWT недоступен
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// registerAndLoginUser - вспомогательная функция для создания тестового пользователя
// и получения его JWT-токена.
func registerAndLoginUser(t *testing.T, testDB *gorm.DB, username, email, password string) (string, uint) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	assert.NoError(t, err) // Проверяем, что хеширование прошло без ошибок

	user := models.User{
		Username: username,
		Email:    email,
		Password: string(hashedPassword),
	}
	result := testDB.Create(&user)  // Создаем пользователя напрямую в БД
	assert.NoError(t, result.Error) // Проверяем, что пользователь успешно создан

	// Генерируем JWT-токен для только что созданного пользователя.
	// Используйте utils.GenerateJWT, если он корректно работает и доступен.
	// Если нет, можно использовать прямую генерацию токена для тестов:
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": float64(user.ID),                      // JWT claims обычно используют float64 для чисел
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Токен действует 24 часа
	})
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	assert.NoError(t, err)

	return tokenString, user.ID
}

// --- Основные тестовые сценарии для NoteController ---

func TestNoteController(t *testing.T) {
	// Настраиваем новую тестовую БД для этого тестового набора
	testDB := setupTestDB()
	// Откладываем закрытие подключения к БД до завершения всех тестов в этом файле
	defer func() {
		sqlDB, _ := testDB.DB()
		sqlDB.Close()
	}()

	// Инициализируем роутер Gin для тестирования
	r := gin.Default()
	// Применяем ваш AuthMiddleware ко всем маршрутам заметок
	r.Use(middleware.AuthMiddleware())
	r.GET("/notes", controllers.GetNotes)
	r.GET("/notes/:id", controllers.GetNote)
	r.POST("/notes", controllers.CreateNote)
	r.PUT("/notes/:id", controllers.UpdateNote)
	r.DELETE("/notes/:id", controllers.DeleteNote)

	// Регистрируем двух тестовых пользователей и получаем их токены/ID
	token, userID := registerAndLoginUser(t, testDB, "testuser_notes", "notes@example.com", "password123")
	_, userID2 := registerAndLoginUser(t, testDB, "anotheruser_notes", "another@example.com", "password123")

	// Эта переменная будет хранить ID заметки, созданной в первом тесте,
	// чтобы другие подтесты могли её использовать.
	var createdNoteID uint

	// --- Подтесты ---

	t.Run("CreateNote - Successful", func(t *testing.T) {
		t.Log("Запуск: CreateNote - Успешное создание заметки") // Логируем начало подтеста
		newNote := models.Note{
			Title:   "Моя первая заметка",
			Content: "Это содержимое моей первой заметки.",
		}
		jsonValue, err := json.Marshal(newNote)
		assert.NoError(t, err) // Проверяем, что маршалинг JSON прошел без ошибок

		req, _ := http.NewRequest(http.MethodPost, "/notes", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token) // Добавляем токен первого пользователя
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code, "Ожидаем статус 201 Created")
		var createdNote models.Note
		err = json.Unmarshal(w.Body.Bytes(), &createdNote)
		assert.NoError(t, err) // Проверяем, что размаршалинг JSON прошел без ошибок

		assert.Equal(t, newNote.Title, createdNote.Title, "Заголовок заметки должен совпадать")
		assert.Equal(t, newNote.Content, createdNote.Content, "Содержимое заметки должно совпадать")
		assert.Equal(t, userID, createdNote.UserID, "UserID заметки должен быть равен ID создавшего пользователя")
		assert.NotZero(t, createdNote.ID, "ID созданной заметки не должен быть равен нулю")

		createdNoteID = createdNote.ID // Сохраняем ID созданной заметки для последующих тестов
	})

	t.Run("CreateNote - Invalid Input (missing title)", func(t *testing.T) {
		t.Log("Запуск: CreateNote - Неверный ввод (отсутствует заголовок)")
		// Заголовок отсутствует, что должно вызвать ошибку валидации из `binding:"required"`
		newNote := models.Note{
			Content: "Это заметка без заголовка.",
		}
		jsonValue, err := json.Marshal(newNote)
		assert.NoError(t, err)

		req, _ := http.NewRequest(http.MethodPost, "/notes", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "Ожидаем статус 400 Bad Request")
		assert.Contains(t, w.Body.String(), "Неверный ввод", "Сообщение об ошибке должно содержать 'Неверный ввод'")
	})

	t.Run("GetNotes - Successful", func(t *testing.T) {
		t.Log("Запуск: GetNotes - Успешное получение всех заметок пользователя")
		// Создаем еще одну заметку для testuser_notes
		// ИСПРАВЛЕНИЕ:
		result1 := testDB.Create(&models.Note{Title: "Вторая заметка", Content: "Контент 2", UserID: userID})
		assert.NoError(t, result1.Error, "Ошибка при создании второй заметки для testuser_notes")
		// assert.Greater(t, result1.RowsAffected, int64(0), "Вторая заметка должна быть создана") // Опционально: проверка RowsAffected

		// Создаем заметку для другого пользователя, она не должна быть получена первым пользователем
		// ИСПРАВЛЕНИЕ:
		result2 := testDB.Create(&models.Note{Title: "Заметка другого пользователя", Content: "Контент 3", UserID: userID2})
		assert.NoError(t, result2.Error, "Ошибка при создании заметки для другого пользователя")
		// assert.Greater(t, result2.RowsAffected, int64(0), "Заметка другого пользователя должна быть создана") // Опционально

		req, _ := http.NewRequest(http.MethodGet, "/notes", nil)
		req.Header.Set("Authorization", "Bearer "+token) // Токен первого пользователя
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Ожидаем статус 200 OK")
		var notes []models.Note
		err := json.Unmarshal(w.Body.Bytes(), &notes) // Переменная err уже объявлена выше, используем := только при первом объявлении
		assert.NoError(t, err)

		assert.Len(t, notes, 2, "Должны быть получены только 2 заметки, принадлежащие testuser_notes")
		for _, note := range notes {
			assert.Equal(t, userID, note.UserID, "Все полученные заметки должны принадлежать текущему пользователю")
		}
	})

	t.Run("GetNote - Successful", func(t *testing.T) {
		t.Log("Запуск: GetNote - Успешное получение одной заметки по ID")
		// Используем createdNoteID, полученный из первого подтеста
		req, _ := http.NewRequest(http.MethodGet, "/notes/"+strconv.FormatUint(uint64(createdNoteID), 10), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Ожидаем статус 200 OK")
		var fetchedNote models.Note
		err := json.Unmarshal(w.Body.Bytes(), &fetchedNote)
		assert.NoError(t, err)

		assert.Equal(t, createdNoteID, fetchedNote.ID, "ID полученной заметки должен совпадать")
		assert.Equal(t, userID, fetchedNote.UserID, "UserID полученной заметки должен совпадать")
	})

	t.Run("GetNote - Not Found or Not Owned", func(t *testing.T) {
		t.Log("Запуск: GetNote - Заметка не найдена или не принадлежит пользователю")
		// Попытка получить заметку, которая не принадлежит текущему пользователю (userID)
		// Для этого найдем ID заметки, принадлежащей userID2
		var anotherNote models.Note
		result := testDB.Where("user_id = ?", userID2).First(&anotherNote)
		assert.NoError(t, result.Error, "Должна быть заметка для userID2")

		req, _ := http.NewRequest(http.MethodGet, "/notes/"+strconv.FormatUint(uint64(anotherNote.ID), 10), nil)
		req.Header.Set("Authorization", "Bearer "+token) // Токен первого пользователя
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code, "Ожидаем статус 404 Not Found")
		assert.Contains(t, w.Body.String(), "Заметка не найдена или не принадлежит вам", "Сообщение об ошибке должно указывать на отсутствие или непринадлежность заметки")
	})

	t.Run("UpdateNote - Successful", func(t *testing.T) {
		t.Log("Запуск: UpdateNote - Успешное обновление заметки")
		updatedNotePayload := models.Note{
			Title:   "Обновленная заметка",
			Content: "Новое содержимое обновленной заметки.",
		}
		jsonValue, err := json.Marshal(updatedNotePayload)
		assert.NoError(t, err)

		req, _ := http.NewRequest(http.MethodPut, "/notes/"+strconv.FormatUint(uint64(createdNoteID), 10), bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Ожидаем статус 200 OK")
		var returnedNote models.Note
		err = json.Unmarshal(w.Body.Bytes(), &returnedNote)
		assert.NoError(t, err)

		assert.Equal(t, updatedNotePayload.Title, returnedNote.Title, "Заголовок должен быть обновлен")
		assert.Equal(t, updatedNotePayload.Content, returnedNote.Content, "Содержимое должно быть обновлено")
		assert.Equal(t, createdNoteID, returnedNote.ID, "ID обновленной заметки должен совпадать")
		assert.Equal(t, userID, returnedNote.UserID, "UserID обновленной заметки должен совпадать")

		// Проверяем, что изменения сохранены в БД
		var dbNote models.Note
		result := testDB.First(&dbNote, createdNoteID)
		assert.NoError(t, result.Error)
		assert.Equal(t, updatedNotePayload.Title, dbNote.Title, "Заголовок в БД должен быть обновлен")
		assert.Equal(t, updatedNotePayload.Content, dbNote.Content, "Содержимое в БД должно быть обновлено")
	})

	t.Run("UpdateNote - Not Found or Not Owned", func(t *testing.T) {
		t.Log("Запуск: UpdateNote - Заметка не найдена или не принадлежит пользователю")
		// Попытка обновить заметку, которая не принадлежит текущему пользователю
		var anotherNote models.Note
		result := testDB.Where("user_id = ?", userID2).First(&anotherNote)
		assert.NoError(t, result.Error, "Должна быть заметка для userID2")

		updatedNotePayload := models.Note{
			Title:   "Попытка обновить чужую заметку",
			Content: "Не должно сработать.",
		}
		jsonValue, err := json.Marshal(updatedNotePayload)
		assert.NoError(t, err)

		req, _ := http.NewRequest(http.MethodPut, "/notes/"+strconv.FormatUint(uint64(anotherNote.ID), 10), bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token) // Токен первого пользователя
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code, "Ожидаем статус 404 Not Found")
		assert.Contains(t, w.Body.String(), "Заметка не найдена или не принадлежит вам", "Сообщение об ошибке должно указывать на отсутствие или непринадлежность заметки")
	})

	t.Run("DeleteNote - Successful", func(t *testing.T) {
		t.Log("Запуск: DeleteNote - Успешное удаление заметки")
		// Используем createdNoteID, который был создан в первом подтесте
		req, _ := http.NewRequest(http.MethodDelete, "/notes/"+strconv.FormatUint(uint64(createdNoteID), 10), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Ожидаем статус 200 OK")
		assert.Contains(t, w.Body.String(), "Заметка успешно удалена", "Сообщение должно подтверждать удаление")

		// Проверяем, что заметка действительно удалена из БД
		var deletedNote models.Note
		result := testDB.First(&deletedNote, createdNoteID)
		assert.Error(t, result.Error, "Ожидаем ошибку 'запись не найдена' после удаления")
		assert.Equal(t, gorm.ErrRecordNotFound, result.Error, "Ошибка должна быть gorm.ErrRecordNotFound")
	})

	t.Run("DeleteNote - Not Found or Not Owned", func(t *testing.T) {
		t.Log("Запуск: DeleteNote - Заметка не найдена или не принадлежит пользователю")
		// Попытка удалить заметку, которая не принадлежит текущему пользователю
		var anotherNote models.Note
		result := testDB.Where("user_id = ?", userID2).First(&anotherNote)
		assert.NoError(t, result.Error, "Должна быть заметка для userID2")

		req, _ := http.NewRequest(http.MethodDelete, "/notes/"+strconv.FormatUint(uint64(anotherNote.ID), 10), nil)
		req.Header.Set("Authorization", "Bearer "+token) // Токен первого пользователя
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code, "Ожидаем статус 404 Not Found")
		assert.Contains(t, w.Body.String(), "Заметка не найдена или не принадлежит вам", "Сообщение об ошибке должно указывать на отсутствие или непринадлежность заметки")
	})

	t.Run("Unauthorized Access - Missing Token", func(t *testing.T) {
		t.Log("Запуск: Unauthorized Access - Отсутствует токен")
		req, _ := http.NewRequest(http.MethodGet, "/notes", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code, "Ожидаем статус 401 Unauthorized")
		assert.Contains(t, w.Body.String(), "Требуется токен авторизации", "Сообщение об ошибке должно указывать на отсутствие токена")
	})

	t.Run("Unauthorized Access - Invalid Token", func(t *testing.T) {
		t.Log("Запуск: Unauthorized Access - Недействительный токен")
		req, _ := http.NewRequest(http.MethodGet, "/notes", nil)
		req.Header.Set("Authorization", "Bearer invalid.token.string")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code, "Ожидаем статус 401 Unauthorized")
		assert.Contains(t, w.Body.String(), "Неверный токен авторизации", "Сообщение об ошибке должно указывать на недействительный токен")
	})
}
