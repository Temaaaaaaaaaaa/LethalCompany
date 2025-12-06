package models

import (
	"time"
)

type User struct {
	ID       int    //id_пользователя
	Username string //имя_пользователя
	Login    string //логин
	Password string //паролик
	Role     string //роль
}

type ModPack struct {
	ID          int       //id_сборки
	Version     string    //версия_сборки
	GameVersion string    //версия_игры
	Description string    //описание_сборки
	PublishDate time.Time //дата_публикации
}

type Download struct {
	ID           int       //id_скачивания
	UserID       int       //id_пользователя
	ModPackID    int       //id_сборки
	DownloadDate time.Time //дата_скачивания
}

type Mods struct {
	ID       int
	Name_mod string
	Opisanye string
	Version  string
}
