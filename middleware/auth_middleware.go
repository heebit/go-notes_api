package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Требуется токен авторизации"})
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Неверный токен авторизации"})
			return
		}
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			c.Set("user_id", uint(claims["user_id"].(float64)))
			c.Next()
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Неверный токен авторизации"})
		}
	}
}
// AuthMiddleware проверяет наличие и валидность JWT токена в заголовке запроса.
// Если токен валиден, извлекает user_id и добавляет его в контекст запроса.
// Если токен отсутствует или недействителен, возвращает ошибку 401 Unauthorized с соответствующим сообщением.
// Этот middleware должен быть применен к защищенным маршрутам, чтобы обеспечить доступ только авторизованным пользователям.		