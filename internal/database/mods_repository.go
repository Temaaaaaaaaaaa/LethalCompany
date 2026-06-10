package database

import (
	"lethalcompany/internal/models"
)

// список модов
func (db *Database) GetMods(versionFilter string) ([]models.Mods, error) {

	var query string
	var args []interface{}

	if versionFilter != "" {
		query = `SELECT "id","название_мода","описание","версия_сборки"
		         FROM "моды"
		         WHERE "версия_сборки" LIKE '%' || $1 || '%'
		         ORDER BY "id"`
		args = append(args, versionFilter)
	} else {
		query = `SELECT "id","название_мода","описание","версия_сборки"
		         FROM "моды"
		         ORDER BY "id"`
	}

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mods []models.Mods

	for rows.Next() {
		var m models.Mods

		err := rows.Scan(
			&m.ID,
			&m.Name_mod,
			&m.Opisanye,
			&m.Version,
		)

		if err != nil {
			return nil, err
		}

		mods = append(mods, m)
	}

	return mods, nil
}

// мод по ID
func (db *Database) GetModByID(id int) (models.Mods, error) {

	var mod models.Mods

	err := db.DB.QueryRow(`
		SELECT "id","название_мода","описание","версия_сборки"
		FROM "моды"
		WHERE "id"=$1
	`, id).Scan(
		&mod.ID,
		&mod.Name_mod,
		&mod.Opisanye,
		&mod.Version,
	)

	return mod, err
}

// добавление
func (db *Database) AddMod(name, description, versions string) error {

	_, err := db.DB.Exec(`
		INSERT INTO "моды"
		("название_мода","описание","версия_сборки")
		VALUES ($1,$2,$3)
	`, name, description, versions)

	return err
}

// обновление
func (db *Database) UpdateMod(id int, name, description, versions string) error {

	_, err := db.DB.Exec(`
		UPDATE "моды"
		SET "название_мода"=$1,
		    "описание"=$2,
		    "версия_сборки"=$3
		WHERE "id"=$4
	`, name, description, versions, id)

	return err
}

// удаление
func (db *Database) DeleteMod(id int) error {

	_, err := db.DB.Exec(`
		DELETE FROM "моды"
		WHERE "id"=$1
	`, id)

	return err
}
