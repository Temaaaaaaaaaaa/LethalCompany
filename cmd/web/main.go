package main

import (
	"fmt"
	"lethalcompany/internal/database"
	"lethalcompany/internal/handlers"
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	DB, err := database.New()
	if err != nil {
		log.Fatal(err)
	}
	h := handlers.Header{
		DB: DB,
	}
	fs := http.FileServer(http.Dir("ui/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", h.HomeHandler)
	http.HandleFunc("/register", h.RegisterHandler)
	http.HandleFunc("/login", h.LoginHandler)
	http.HandleFunc("/logout", h.LogoutHandler)
	http.HandleFunc("/modpacks", h.ModPacksHandler)
	http.HandleFunc("/versions", h.ModsHandler)
	http.HandleFunc("/profile", h.ProfileHandler)
	http.HandleFunc("/admin_modpack_create", h.AdminCreateModPackHandler)
	http.HandleFunc("/admin_modadd", h.Add_modHandler)
	http.HandleFunc("/admin_mod_edit/", h.EditModHandler)
	http.HandleFunc("/admin_mod_delete", h.DeleteModHandler)

	fmt.Println("Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
