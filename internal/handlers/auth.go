package handlers

import (
	"fmt"
	"html/template"
	"lethalcompany/internal/database"
	"lethalcompany/internal/models"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// renderPage — рендер HTML
func renderPage(w http.ResponseWriter, baseFile, pageFile string, data any) {
	tmpl := template.Must(template.ParseFiles(baseFile, pageFile))
	err := tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// PageData
type PageData struct {
	Title string
	User  *models.User
}

// главная
func HomeHandler(w http.ResponseWriter, r *http.Request) {

	data := PageData{
		Title: "Главная",
		User:  getCurrentUser(r),
	}

	renderPage(w, "ui/html/base.html", "ui/html/home.html", data)
}

// регистрация
func (h *Header) RegisterHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {

		data := PageData{
			Title: "Регистрация",
			User:  getCurrentUser(r),
		}

		renderPage(w, "ui/html/base.html", "ui/html/register.html", data)
		return
	}

	_ = r.ParseForm()

	username := strings.TrimSpace(r.FormValue("username"))
	login := strings.TrimSpace(r.FormValue("login"))
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")

	if password != confirmPassword {
		http.Error(w, "Пароли не совпадают", http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	exists, err := h.db.UserExists(login)
	if err != nil {
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	if exists {
		http.Error(w, "Такой логин уже существует", http.StatusBadRequest)
		return
	}

	err = database.CreateUser(username, login, string(hash))
	if err != nil {
		http.Error(w, "Ошибка регистрации", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// вход
func LoginHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {

		data := PageData{
			Title: "Вход",
			User:  getCurrentUser(r),
		}

		renderPage(w, "ui/html/base.html", "ui/html/login.html", data)
		return
	}

	_ = r.ParseForm()

	login := strings.TrimSpace(r.FormValue("login"))
	password := r.FormValue("password")

	user, err := database.GetUserByLogin(login)
	if err != nil {
		http.Error(w, "Пользователь не найден", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(password),
	)

	if err != nil {
		http.Error(w, "Неверный пароль", http.StatusUnauthorized)
		return
	}

	// сессия
	sessionID := fmt.Sprintf("%d_%d", user.ID, time.Now().UnixNano())
	sessions[sessionID] = &user

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now().Add(24 * time.Hour),
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// выход
func LogoutHandler(w http.ResponseWriter, r *http.Request) {

	cookie, err := r.Cookie("session_id")

	if err == nil {
		delete(sessions, cookie.Value)

		cookie.Expires = time.Now().Add(-1 * time.Hour)
		http.SetCookie(w, cookie)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// профиль
func ProfileHandler(w http.ResponseWriter, r *http.Request) {

	data := PageData{
		Title: "Профиль",
		User:  getCurrentUser(r),
	}

	renderPage(w, "ui/html/base.html", "ui/html/profile.html", data)
}
