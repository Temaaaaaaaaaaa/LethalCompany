package handlers

import (
	"fmt"
	"html/template"                   // для работы с html-ками
	"lethalcompany/internal/database" // для работы с базой данных
	"lethalcompany/internal/models"   // для доступа к структурам
	"net/http"                        // для http запросов
	"strings"                         // для удаления пробелов
	"time"

	"golang.org/x/crypto/bcrypt" // для хеширования паролей
)

// Функция для рендеринга страницы
// baseFile -базовый шаблон и pageFile -дочерний шаблон
func renderPage(w http.ResponseWriter, baseFile, pageFile string, data any) {
	tmpl := template.Must(template.ParseFiles(baseFile, pageFile)) //must что бы ошибок при парсинге не было
	err := tmpl.ExecuteTemplate(w, "base.html", data)              // а ParseFiles парсит их в шаблоны что бы {.User} были заполнены данными
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// PageData
// что бы каждый раз не создавать структуру
type PageData struct {
	Title string
	User  *models.User
}

// первая страница
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title: "Главная",
		User:  getCurrentUser(r), // узнаём пользователя по куки
	}
	renderPage(w, "ui/html/base.html", "ui/html/home.html", data)
}

// регистриация
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		data := PageData{
			Title: "Регистрация",
			User:  getCurrentUser(r),
		}
		renderPage(w, "ui/html/base.html", "ui/html/register.html", data)
		return
	}

	err := r.ParseForm() // парсим данные
	if err != nil {
		return
	}
	username := strings.TrimSpace(r.FormValue("username")) //удаление пробелов в начале и конце и берём боле из html username
	login := strings.TrimSpace(r.FormValue("login"))
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")

	if password != confirmPassword {
		http.Error(w, "Пароли не совпадают", http.StatusBadRequest)
		return
	}
	// bcrypt.DefaultCost стандартный параемтр пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	db := database.Connect()
	defer db.Close() //откладываем закрытие базы данных до конца выполнения этой функции

	var user_yet bool
	err = db.QueryRow(`SELECT EXISTS(SELECT 1 FROM "пользователи" WHERE "логин" = $1)`, login).Scan(&user_yet)
	if err != nil {
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}
	if user_yet {
		http.Error(w, "Такой логин уже существует", http.StatusBadRequest)
		return
	}
	// "_" - так как нам не нужно ничег от этой переменной
	_, err = db.Exec(`INSERT INTO "пользователи" ("имя_пользователя","логин","паролик","роль")
	                  VALUES ($1,$2,$3,$4)`, // плейсхолдеры для защиты от SQL-инъекций
		username, login, string(hashedPassword), "пользователь")
	if err != nil {
		http.Error(w, "Ошибка при регистрации", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther) // если успешная регистрация, то перенаправление на вход
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

	err := r.ParseForm() // берём данные что бы получить потом логин и пароль, без него FormValue не робит
	if err != nil {
		return
	}
	login := strings.TrimSpace(r.FormValue("login")) //убираем пробелы слева и справа и берём данные
	password := r.FormValue("password")

	db := database.Connect()
	defer db.Close()

	var user models.User
	err = db.QueryRow(`SELECT "id_пользователя","имя_пользователя","логин","паролик","роль"
	                    FROM "пользователи"
	                    WHERE "логин" = $1`, login).
		Scan(&user.ID, &user.Username, &user.Login, &user.Password, &user.Role) //записываем результаты в поля структуры; & адрес переменной, что бы изменить напрямую
	if err != nil {
		http.Error(w, "Пользователь не найден", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) //среавниваем пароли
	if err != nil {
		http.Error(w, "Неверный пароль", http.StatusUnauthorized)
		return
	}

	// Создание сессии
	sessionID := fmt.Sprintf("%d_%d", user.ID, time.Now().UnixNano()) //в нано секундах, что бы сессии были уникальными
	sessions[sessionID] = &user                                       //сохраняем сессию
	http.SetCookie(w, &http.Cookie{ // отправка браузеру куки
		Name:     "session_id",                   // имя
		Value:    sessionID,                      // индентификатор
		Path:     "/",                            // доступен на всех страницах
		HttpOnly: true,                           // только через http
		Expires:  time.Now().Add(24 * time.Hour), // куки дляться 24 часа
	})

	http.Redirect(w, r, "/", http.StatusSeeOther) // после успешного входа идём на главную страницу
}

// выход из профиля
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id") // получаем ссесию по куки
	if err == nil {
		delete(sessions, cookie.Value)                  // удаляем сессию
		cookie.Expires = time.Now().Add(-1 * time.Hour) //меняем время куки, что бы браузер удалил его
		http.SetCookie(w, cookie)                       // отправляем назад куки браузеру
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
