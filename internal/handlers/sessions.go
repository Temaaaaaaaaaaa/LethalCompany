package handlers

import (
	"lethalcompany/internal/models"
	"net/http"
)

// Переменная для хранения сессии
var sessions = map[string]*models.User{} //здесь ключ sessionID значение *modells.User
// Получение текущего пользователя по cookie
func getCurrentUser(r *http.Request) *models.User {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return nil
	}
	user, ok := sessions[cookie.Value] //уневирасальный индентификатор
	if !ok {
		return nil
	}
	return user
}
