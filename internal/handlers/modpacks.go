package handlers

import (
	"lethalcompany/internal/database"
	"lethalcompany/internal/models"
	"net/http"
)

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
