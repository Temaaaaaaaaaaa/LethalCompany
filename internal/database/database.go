package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq" // нужен для postgres а "_" так как он на прямую не используется
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "864250"
	dbname   = "lethalcompany"
)

func Connect() *sql.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	// создаём объект *sql.DB
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
	}
	// подключаем
	err = db.Ping()
	if err != nil {
	}

	fmt.Println("Успешное подключение к базе данных!")
	return db
}
