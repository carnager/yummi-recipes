package main

import (
	"database/sql"
	"embed"
	"log"
	"sort"
	"strings"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// splitStatements splits SQL respecting BEGIN...END blocks (for triggers).
func splitStatements(sql string) []string {
	var stmts []string
	var current strings.Builder
	inBlock := false

	for _, line := range strings.Split(sql, "\n") {
		trimmed := strings.TrimSpace(strings.ToUpper(line))
		if strings.HasSuffix(trimmed, "BEGIN") {
			inBlock = true
		}
		current.WriteString(line)
		current.WriteString("\n")

		if inBlock && strings.HasSuffix(trimmed, "END;") {
			inBlock = false
			stmt := strings.TrimSpace(current.String())
			if stmt != "" {
				stmts = append(stmts, stmt)
			}
			current.Reset()
		} else if !inBlock && strings.Contains(line, ";") {
			// Split on semicolons outside blocks
			parts := strings.Split(current.String(), ";")
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p != "" {
					stmts = append(stmts, p)
				}
			}
			current.Reset()
		}
	}

	if remaining := strings.TrimSpace(current.String()); remaining != "" {
		stmts = append(stmts, remaining)
	}

	return stmts
}

func runMigrations(db *sql.DB) {
	db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		filename TEXT PRIMARY KEY,
		applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`)

	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		log.Fatalf("Migrationen lesen: %v", err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		var count int
		db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE filename = ?", entry.Name()).Scan(&count)
		if count > 0 {
			continue
		}

		data, err := migrationsFS.ReadFile("migrations/" + entry.Name())
		if err != nil {
			log.Fatalf("Migration %s lesen: %v", entry.Name(), err)
		}

		for _, stmt := range splitStatements(string(data)) {
			if _, err := db.Exec(stmt); err != nil {
				log.Fatalf("Migration %s ausfuehren: %v\nStatement: %s", entry.Name(), err, stmt)
			}
		}

		db.Exec("INSERT INTO schema_migrations (filename) VALUES (?)", entry.Name())
		log.Printf("Migration angewendet: %s", entry.Name())
	}
}
