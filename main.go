package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello World!")
	})
	log.Print("Серва на http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
	}
}
