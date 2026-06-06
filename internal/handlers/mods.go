package handlers

import (
	"lethalcompany/internal/database"
	"lethalcompany/internal/models"
	"net/http"
	"strconv"
	"strings"
)

// список модов
func ModsHandler(w http.ResponseWriter, r *http.Request) {

	versionFilter := r.URL.Query().Get("version")

	versions, err := database.GetAllVersions()
	if err != nil {
		http.Error(w, "Ошибка получения версий: "+err.Error(), http.StatusInternalServerError)
		return
	}

	mods, err := database.GetMods(versionFilter)
	if err != nil {
		http.Error(w, "Ошибка получения модов: "+err.Error(), http.StatusInternalServerError)
		return
	}

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

// добавление мода
func Add_modHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {

		modpacks, err := database.GetAllModpacks()
		if err != nil {
			http.Error(w, "Ошибка получения сборок: "+err.Error(), http.StatusInternalServerError)
			return
		}

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

	if err := r.ParseForm(); err != nil {
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

	var versions []string

	for _, idStr := range modpackIDs {

		version, err := database.GetVersionByModpackID(idStr)
		if err != nil {
			continue
		}

		versions = append(versions, version)
	}

	versionsStr := strings.Join(versions, ", ")

	err := database.AddMod(
		name,
		description,
		versionsStr,
	)

	if err != nil {
		http.Error(w, "Ошибка добавления мода: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/versions", http.StatusSeeOther)
}

// редактирование мода
func EditModHandler(w http.ResponseWriter, r *http.Request) {

	idStr := r.URL.Query().Get("id")

	if idStr == "" {
		http.Error(w, "ID мода не указан", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Некорректный ID мода", http.StatusBadRequest)
		return
	}

	if r.Method == "GET" {

		mod, err := database.GetModByID(id)
		if err != nil {
			http.Error(w, "Мод не найден", http.StatusNotFound)
			return
		}

		modpacks, err := database.GetAllModpacks()
		if err != nil {
			http.Error(w, "Ошибка получения сборок", http.StatusInternalServerError)
			return
		}

		selectedVersions := map[string]bool{}

		for _, version := range strings.Split(mod.Version, ", ") {
			selectedVersions[version] = true
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

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Ошибка обработки формы", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	description := strings.TrimSpace(r.FormValue("description"))
	selectedVersions := r.Form["modpack_ids"]

	if name == "" || description == "" || len(selectedVersions) == 0 {
		http.Error(w, "Все поля обязательны", http.StatusBadRequest)
		return
	}

	versionStr := strings.Join(selectedVersions, ", ")

	err = database.UpdateMod(
		id,
		name,
		description,
		versionStr,
	)

	if err != nil {
		http.Error(w, "Ошибка обновления мода: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/versions", http.StatusSeeOther)
}

// удаление мода
func DeleteModHandler(w http.ResponseWriter, r *http.Request) {

	idStr := r.URL.Query().Get("id")

	if idStr == "" {
		http.Error(w, "ID мода не указан", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Некорректный ID мода", http.StatusBadRequest)
		return
	}

	err = database.DeleteMod(id)
	if err != nil {
		http.Error(w, "Ошибка удаления мода: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/versions", http.StatusSeeOther)
}
