package database

import (
	"lethalcompany/internal/models"
)

func CreateModpack(version, gameVersion, description string) error {

	db := DB

	_, err := db.Exec(`
		INSERT INTO "сборка_мода"
		(версия_сборки, версия_игры, ссылка, дата_публикации)
		VALUES ($1, $2, $3, NOW())
	`,
		version,
		gameVersion,
		description,
	)

	return err
}
func GetAllVersions() ([]string, error) {
	db := DB

	rows, err := db.Query(`
		SELECT DISTINCT "версия_сборки"
		FROM "сборка_мода"
		ORDER BY "версия_сборки"
	`)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var versions []string

	for rows.Next() {
		var version string

		if err := rows.Scan(&version); err != nil {
			continue
		}

		versions = append(versions, version)
	}

	return versions, nil
}

func GetAllModpacks() ([]models.ModPack, error) {
	db := DB

	rows, err := db.Query(`
		SELECT id_сборки,
		       версия_сборки,
		       версия_игры,
		       дата_публикации
		FROM "сборка_мода"
		ORDER BY дата_публикации DESC
	`)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var modpacks []models.ModPack

	for rows.Next() {
		var mp models.ModPack

		err := rows.Scan(
			&mp.ID,
			&mp.Version,
			&mp.GameVersion,
			&mp.PublishDate,
		)

		if err != nil {
			return nil, err
		}

		modpacks = append(modpacks, mp)
	}

	return modpacks, nil
}

func GetVersionByModpackID(id string) (string, error) {
	db := DB

	var version string

	err := db.QueryRow(
		`SELECT версия_сборки
		 FROM "сборка_мода"
		 WHERE id_сборки=$1`,
		id,
	).Scan(&version)

	return version, err
}
