package main

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func openDB(dataDir string) *sql.DB {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("Datenverzeichnis erstellen: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dataDir, "uploads"), 0755); err != nil {
		log.Fatalf("Upload-Verzeichnis erstellen: %v", err)
	}

	dbPath := filepath.Join(dataDir, "yummi.db")
	db, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(wal)&_pragma=foreign_keys(1)")
	if err != nil {
		log.Fatalf("Datenbank oeffnen: %v", err)
	}

	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		log.Fatalf("Datenbank ping: %v", err)
	}

	return db
}
