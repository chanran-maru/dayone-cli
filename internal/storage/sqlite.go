package storage

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func OpenDB() (*sql.DB, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	dir := filepath.Join(configDir, "DayOne")

	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}

	dbPath := filepath.Join(dir, "dayone.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	// journals 테이블 생성
	query := `
        CREATE TABLE IF NOT EXISTS journals (
            date TEXT PRIMARY KEY,
            content TEXT NOT NULL,
            created_at TEXT NOT NULL,
            updated_at TEXT NOT NULL
        ); `
	if _, err := db.Exec(query); err != nil {
		db.Close()
		return nil, err
	}

	// tasks 테이블 생성
	query = `
        CREATE TABLE IF NOT EXISTS tasks (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            title TEXT NOT NULL,
            status TEXT NOT NULL DEFAULT 'pending',
            created_at TEXT NOT NULL
        ); `
	if _, err := db.Exec(query); err != nil {
		db.Close()
		return nil, err
	}

	// reviews 테이블 생성
	query = `
        CREATE TABLE IF NOT EXISTS reviews (
            date TEXT PRIMARY KEY,
            mood TEXT NOT NULL,
            rating INTEGER NOT NULL,
            content TEXT,
            created_at TEXT NOT NULL,
            updated_at TEXT NOT NULL
        ); `
	if _, err := db.Exec(query); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
