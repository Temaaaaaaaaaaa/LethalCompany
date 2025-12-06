package handlers

import (
	"fmt"
	"html/template"                   // для работы с html-ками
	"lethalcompany/internal/database" // для работы с базой данных
	"lethalcompany/internal/models"   // для доступа к структурам
	"net/http"                        // для http запросов
	"strconv"                         // для конвертации
	"strings"                         // для удаления пробелов
	"time"

	"golang.org/x/crypto/bcrypt" // для хеширования паролей
)

// Переменная для хранения сессии
var sessions = map[string]*models.User{} //здесь ключ sessionID значение *modells.User

// Функция для рендеринга страницы
// baseFile -базовый шаблон и pageFile -дочерний шаблон
func renderPage(w http.ResponseWriter, baseFile, pageFile string, data any) {
	tmpl := template.Must(template.ParseFiles(baseFile, pageFile)) //must что бы ошибок при парсинге не было
	err := tmpl.ExecuteTemplate(w, "base.html", data)              // а ParseFiles парсит их в шаблоны что бы {.User} были заполнены данными
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

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

// сборка мода
func ModPacksHandler(w http.ResponseWriter, r *http.Request) {
	db := database.Connect()
	defer db.Close()

	rows, err := db.Query(`
        SELECT id_сборки, версия_сборки, версия_игры, ссылка, дата_публикации
        FROM "сборка_мода"
        ORDER BY дата_публикации DESC`) // читаем таблицу
	if err != nil {
		http.Error(w, "Ошибка запроса: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var modpacks []models.ModPack // динамический массив, что бы сохранить сборки

	for rows.Next() { //идём по каждой строке
		var mp models.ModPack
		err := rows.Scan(&mp.ID, &mp.Version, &mp.GameVersion, &mp.Description, &mp.PublishDate) // считываем данные и с & записываем в mp
		if err != nil {
			http.Error(w, "Ошибка чтения данных: "+err.Error(), http.StatusInternalServerError)
			return
		}
		modpacks = append(modpacks, mp) // добавляем строку в массив
	}

	data := struct { // создаём анонимную структуру и записываем в неё данные
		Title    string
		User     *models.User
		ModPacks []models.ModPack
	}{
		Title:    "Сборки модов",
		User:     getCurrentUser(r),
		ModPacks: modpacks,
	}

	renderPage(w, "ui/html/base.html", "ui/html/modpacks.html", data)
}

// моды
func ModsHandler(w http.ResponseWriter, r *http.Request) {
	db := database.Connect()
	defer db.Close()

	// Получаем параметр фильтра из URL, если есть
	versionFilter := r.URL.Query().Get("version")

	//  Получаем список всех версий сборок из таблицы "сборка_мода" все они уникальные
	versionRows, err := db.Query(`SELECT DISTINCT "версия_сборки" FROM "сборка_мода" ORDER BY "версия_сборки"`)
	if err != nil {
		http.Error(w, "Ошибка получения версий сборок: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer versionRows.Close()

	var versions []string    // слайс или динамичный массив
	for versionRows.Next() { // идём по каждой строке версии
		var v string
		if err := versionRows.Scan(&v); err != nil {
			continue
		} // присваеваем v  версию сборки
		versions = append(versions, v) // добавляем в слайс версию сборки
	}

	// Получаем моды, при необходимости фильтруем по версии сборки
	var query string       // ханим sql запросы
	var args []interface{} // не нулевой слайс интерфейсов (может хранить разные типы данных)

	if versionFilter != "" { // если фильтр задан то
		// Ищем подстроку в поле "версия_сборки", чтобы учесть моды с несколькими версиями
		query = `SELECT "id", "название_мода", "описание", "версия_сборки" 
		         FROM "моды" 
		         WHERE "версия_сборки" LIKE '%' || $1 || '%' 
		         ORDER BY "id"`
		args = append(args, versionFilter)
	} else { // если фильтр пустой то показываем всё
		query = `SELECT "id", "название_мода", "описание", "версия_сборки" 
		         FROM "моды" ORDER BY "id"`
	}

	rows, err := db.Query(query, args...) // "..." разворачивание слайса
	if err != nil {
		http.Error(w, "Ошибка чтения таблицы модов: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var mods []models.Mods // пустой слайс mods типа структуры из моделей таблиц
	for rows.Next() {      // идём по
		var mod models.Mods                                                                    // создаём переменную типа структыр из моделей таблицы\
		if err := rows.Scan(&mod.ID, &mod.Name_mod, &mod.Opisanye, &mod.Version); err != nil { // записываем в mod всё про мод
			http.Error(w, "Ошибка scan: "+err.Error(), http.StatusInternalServerError)
			return
		}
		mods = append(mods, mod) // добавляем мод в моды
	}

	//  Передаем данные в шаблон
	data := struct {
		Title    string
		User     *models.User
		Mods     []models.Mods
		Versions []string
		Filter   string
	}{
		Title:    "Моды",
		User:     getCurrentUser(r),
		Mods:     mods,
		Versions: versions,
		Filter:   versionFilter,
	}

	renderPage(w, "ui/html/base.html", "ui/html/versions.html", data)
}

// профиль
func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title: "Профиль",
		User:  getCurrentUser(r),
	}
	renderPage(w, "ui/html/base.html", "ui/html/profile.html", data)
}

// форма добавления сборки
func AdminCreateModPackHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		data := PageData{
			Title: "Добавить сборку",
			User:  getCurrentUser(r),
		}
		renderPage(w, "ui/html/base.html", "ui/html/admin_create_modpack.html", data)
		return
	}

	// POST
	err := r.ParseForm()
	if err != nil {
		return
	}

	version := r.FormValue("version")          //версия сборки
	gameVersion := r.FormValue("game_version") //версия игры
	description := r.FormValue("description")  //описание

	db := database.Connect()
	defer db.Close()

	_, err = db.Exec(
		`INSERT INTO "сборка_мода" 
        (версия_сборки, версия_игры, ссылка, дата_публикации)
         VALUES ($1, $2, $3, NOW())`,
		version, gameVersion, description,
	)

	if err != nil {
		http.Error(w, "Ошибка сохранения: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/modpacks", http.StatusSeeOther)
}

// форма для ддобавления мода
func Add_modHandler(w http.ResponseWriter, r *http.Request) {
	db := database.Connect()
	defer db.Close()

	if r.Method == "GET" {
		// Получаем все существующие сборки, чтобы выбрать их в форме
		rows, err := db.Query(`SELECT id_сборки, версия_сборки, версия_игры FROM "сборка_мода" ORDER BY дата_публикации DESC`)
		if err != nil {
			http.Error(w, "Ошибка чтения сборок: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var modpacks []models.ModPack
		for rows.Next() {
			var mp models.ModPack
			if err := rows.Scan(&mp.ID, &mp.Version, &mp.GameVersion); err != nil {
				http.Error(w, "Ошибка чтения данных: "+err.Error(), http.StatusInternalServerError)
				return
			}
			modpacks = append(modpacks, mp)
		}

		// Передаем сборки в шаблон
		data := struct {
			Title    string
			User     *models.User
			ModPacks []models.ModPack
		}{
			Title:    "Добавить мод",
			User:     getCurrentUser(r),
			ModPacks: modpacks,
		}

		renderPage(w, "ui/html/base.html", "ui/html/add_mod.html", data)
		return
	}

	// POST — обработка формы
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Ошибка обработки формы", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	description := strings.TrimSpace(r.FormValue("description"))
	modpackIDs := r.Form["modpack_ids"]

	if name == "" || description == "" || len(modpackIDs) == 0 {
		http.Error(w, "Все поля обязательны", http.StatusBadRequest)
		return
	}

	// Получаем текстовые версии выбранных сборок
	var versions []string
	for _, idStr := range modpackIDs {
		var version string
		err := db.QueryRow(`SELECT версия_сборки FROM "сборка_мода" WHERE id_сборки=$1`, idStr).Scan(&version)
		if err != nil {
			continue // если какой-то ID некорректен — пропускаем
		}
		versions = append(versions, version)
	}

	// объединяем версии через запятую
	versionsStr := strings.Join(versions, ", ")

	// вставляем мод с объединёнными версиями
	_, err = db.Exec(`INSERT INTO "моды" ("название_мода", "описание", "версия_сборки") VALUES ($1, $2, $3)`,
		name, description, versionsStr)
	if err != nil {
		http.Error(w, "Ошибка добавления мода: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/versions", http.StatusSeeOther)
}

// редактирование модов
func EditModHandler(w http.ResponseWriter, r *http.Request) {
	db := database.Connect()
	defer db.Close()

	// Получаем ID мода из query
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "ID мода не указан", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(idStr) // конвертируем в число типа int
	if err != nil {
		http.Error(w, "Некорректный ID мода", http.StatusBadRequest)
		return
	}

	if r.Method == "GET" {
		// Получаем данные мода
		var mod models.Mods
		err := db.QueryRow(`SELECT "id", "название_мода", "описание", "версия_сборки" 
		                    FROM "моды" WHERE "id"=$1`, id).
			Scan(&mod.ID, &mod.Name_mod, &mod.Opisanye, &mod.Version)
		if err != nil {
			http.Error(w, "Мод не найден: "+err.Error(), http.StatusNotFound)
			return
		}

		// Получаем все сборки
		rows, err := db.Query(`SELECT id_сборки, версия_сборки, версия_игры FROM "сборка_мода" ORDER BY "версия_сборки"`)
		if err != nil {
			http.Error(w, "Ошибка получения сборок: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		// собираем сборки в срез
		var modpacks []models.ModPack
		for rows.Next() {
			var mp models.ModPack
			if err := rows.Scan(&mp.ID, &mp.Version, &mp.GameVersion); err != nil {
				continue
			}
			modpacks = append(modpacks, mp)
		}

		// Создаем map выбранных версий
		selectedVersions := map[string]bool{}
		for _, v := range strings.Split(mod.Version, ", ") {
			selectedVersions[v] = true // создаём запись в мапе
		}

		data := struct {
			Title            string
			User             *models.User
			Mod              models.Mods
			ModPacks         []models.ModPack
			SelectedVersions map[string]bool
		}{
			Title:            "Редактировать мод",
			User:             getCurrentUser(r),
			Mod:              mod,
			ModPacks:         modpacks,
			SelectedVersions: selectedVersions,
		}

		renderPage(w, "ui/html/base.html", "ui/html/edit_mod.html", data)
		return
	}

	// POST — сохраняем изменения
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Ошибка обработки формы", http.StatusBadRequest)
		return
	}
	// получаем данные
	name := strings.TrimSpace(r.FormValue("name"))
	description := strings.TrimSpace(r.FormValue("description"))
	selectedVersions := r.Form["modpack_ids"]
	// есди хотя бы одно поле пустое или не выбраан версия сборки то выдаём ошибку
	if name == "" || description == "" || len(selectedVersions) == 0 {
		http.Error(w, "Все поля обязательны", http.StatusBadRequest)
		return
	}

	versionStr := strings.Join(selectedVersions, ", ") // превращаем в строку
	// обновляем данные
	_, err = db.Exec(`UPDATE "моды" SET "название_мода"=$1, "описание"=$2, "версия_сборки"=$3 WHERE "id"=$4`,
		name, description, versionStr, id)
	if err != nil {
		http.Error(w, "Ошибка обновления мода: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/versions", http.StatusSeeOther)
}

// удаление мода
func DeleteModHandler(w http.ResponseWriter, r *http.Request) {
	db := database.Connect()
	defer db.Close()
	// берём id
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "ID мода не указан", http.StatusBadRequest)
		return
	}
	// строку в число
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Некорректный ID мода", http.StatusBadRequest)
		return
	}
	// удаление в базе данных
	_, err = db.Exec(`DELETE FROM "моды" WHERE "id"=$1`, id)
	if err != nil {
		http.Error(w, "Ошибка удаления мода: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/versions", http.StatusSeeOther)
}
