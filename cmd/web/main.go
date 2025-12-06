package main

import (
	"fmt"
	"lethalcompany/internal/database"
	"lethalcompany/internal/handlers"
	"net/http"
)

func main() {

	database.Connect()

	http.HandleFunc("/", handlers.HomeHandler)
	http.HandleFunc("/register", handlers.RegisterHandler)
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/logout", handlers.LogoutHandler)
	http.HandleFunc("/modpacks", handlers.ModPacksHandler)
	http.HandleFunc("/versions", handlers.ModsHandler)
	http.HandleFunc("/profile", handlers.ProfileHandler)
	http.HandleFunc("/admin_modpack_create", handlers.AdminCreateModPackHandler)
	http.HandleFunc("/admin_modadd", handlers.Add_modHandler)
	http.HandleFunc("/admin_mod_edit/", handlers.EditModHandler)
	http.HandleFunc("/admin_mod_delete", handlers.DeleteModHandler)

	fmt.Println("Сервер запущен на http://localhost:8080")

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Ошибка запуска сервера:", err)

	}
}
