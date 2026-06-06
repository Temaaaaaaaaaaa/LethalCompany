package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type Database struct {
	db *sql.DB
}

func New() (*Database, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	fmt.Println("Успешное подключение к базе данных!")

	return &Database{db: db}, nil
}

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "864250"
	dbname   = "lethalcompany"
)

var DB *sql.DB

func InitDB() error {

	DB = db
	return nil
}
