package handlers

import (
	"lethalcompany/internal/database"
	"lethalcompany/internal/models"
	"net/http"
	"strconv"
	"strings"
)

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
