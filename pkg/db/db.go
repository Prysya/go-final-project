package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

var (
	db *sql.DB
)

const schema = `
create table if not exists scheduler (
    id integer primary key autoincrement,
    date CHAR(8) not null default "",
    title VARCHAR(255) not null,
    comment text,
    repeat VARCHAR(128)
);

create index if not exists idx_scheduler_date on scheduler(date);
`

func Init(dbFile string) error {
	_, err := os.Stat(dbFile)
	fileExists := err == nil

	db, err = sql.Open("sqlite", dbFile)
	if err != nil {
		return fmt.Errorf("ошибка открытия базы данных: %w", err)
	}

	if err = db.Ping(); err != nil {
		return fmt.Errorf("ошибка подключения к базе данных: %w", err)
	}

	if !fileExists {
		log.Printf("Файл %s не найден, создаем новую базу данных", dbFile)
		if _, err = db.Exec(schema); err != nil {
			return fmt.Errorf("ошибка создания таблицы scheduler: %w", err)
		}
		log.Println("Таблица scheduler успешно создана")
	} else {
		if _, err = db.Exec(schema); err != nil {
			return fmt.Errorf("ошибка проверки таблицы scheduler: %w", err)
		}
		log.Printf("База данных %s уже существует", dbFile)
	}

	return nil
}

func GetDB() *sql.DB {
	return db
}

func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}
