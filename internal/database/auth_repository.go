package database

import (
	"lethalcompany/internal/models"
)

// проверка существует ли пользователь
func (db *Database) UserExists(login string) (bool, error) {
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS(
			SELECT 1
			FROM "пользователи"
			WHERE "логин" = $1
		)
	`, login).Scan(&exists)

	return exists, err
}

// создание пользователя
func CreateUser(username, login, passwordHash string) error {

	db := DB

	_, err := db.Exec(`
		INSERT INTO "пользователи"
		("имя_пользователя","логин","паролик","роль")
		VALUES ($1,$2,$3,$4)
	`,
		username,
		login,
		passwordHash,
		"пользователь",
	)

	return err
}

// получить пользователя по логину
func GetUserByLogin(login string) (models.User, error) {

	db := DB

	var user models.User

	err := db.QueryRow(`
		SELECT
			"id_пользователя",
			"имя_пользователя",
			"логин",
			"паролик",
			"роль"
		FROM "пользователи"
		WHERE "логин" = $1
	`, login).Scan(
		&user.ID,
		&user.Username,
		&user.Login,
		&user.Password,
		&user.Role,
	)

	return user, err
}
