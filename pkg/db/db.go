package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/lib/pq"
)

// DB wraps sql.DB for easier usage
var DB *sql.DB

// Connect connects to the database using a DSN
func Connect(dsn string) error {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}
	if err := db.Ping(); err != nil {
		return err
	}
	DB = db
	return nil
}

// Migrate runs all .up.sql migrations in the given directory
func Migrate(migrationsDir string) error {
	files, err := filepath.Glob(filepath.Join(migrationsDir, "**", "*.up.sql"))
	if err != nil {
		return err
	}
	for _, file := range files {
		fmt.Printf("Running migration: %s\n", file)
		query, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		_, err = DB.Exec(string(query))
		if err != nil {
			return fmt.Errorf("error in migration %s: %w", file, err)
		}
	}
	return nil
}

// Exec executes a query without returning any rows.
func Exec(query string, args ...any) (sql.Result, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not connected")
	}
	return DB.Exec(query, args...)
}

// Query executes a query that returns rows.
func Query(query string, args ...any) (*sql.Rows, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not connected")
	}
	return DB.Query(query, args...)
}

// QueryRow executes a query that is expected to return at most one row.
func QueryRow(query string, args ...any) *sql.Row {
	if DB == nil {
		return nil
	}
	return DB.QueryRow(query, args...)
}

// Begin starts a new transaction.
func Begin() (*sql.Tx, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not connected")
	}
	return DB.Begin()
}

// Close closes the database connection.
func Close() error {
	if DB == nil {
		return nil
	}
	return DB.Close()
}
