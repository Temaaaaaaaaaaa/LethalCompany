package handlers

import (
	"lethalcompany/internal/models"
	"net/http"
)

// список сборок
func (h *Header) ModPacksHandler(w http.ResponseWriter, r *http.Request) {

	modpacks, err := h.DB.GetAllModpacks()

	if err != nil {
		http.Error(
			w,
			"Ошибка запроса: "+err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	data := struct {
		Title    string
		User     *models.User
		ModPacks []models.ModPack
	}{
		Title:    "Сборки модов",
		User:     getCurrentUser(r),
		ModPacks: modpacks,
	}

	h.renderPage(
		w,
		"ui/html/base.html",
		"ui/html/modpacks.html",
		data,
	)
}

// форма добавления сборки
func (h *Header) AdminCreateModPackHandler(
	w http.ResponseWriter,
	r *http.Request,
) {

	if r.Method == "GET" {

		data := PageData{
			Title: "Добавить сборку",
			User:  getCurrentUser(r),
		}

		h.renderPage(
			w,
			"ui/html/base.html",
			"ui/html/admin_create_modpack.html",
			data,
		)

		return
	}

	// POST

	err := r.ParseForm()

	if err != nil {
		http.Error(
			w,
			"Ошибка обработки формы",
			http.StatusBadRequest,
		)
		return
	}

	version := r.FormValue("version")
	gameVersion := r.FormValue("game_version")
	description := r.FormValue("description")

	err = h.DB.CreateModpack(
		version,
		gameVersion,
		description,
	)

	if err != nil {
		http.Error(
			w,
			"Ошибка сохранения: "+err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	http.Redirect(
		w,
		r,
		"/modpacks",
		http.StatusSeeOther,
	)
}
