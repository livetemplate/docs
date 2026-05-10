package todos

import (
	_ "embed"

	"database/sql"
	"fmt"
	"log"

	"github.com/livetemplate/docs/content/recipes/todos/_app/db"
	_ "modernc.org/sqlite"
)

// schemaSQL is the canonical schema definition shared with sqlc — both
// the runtime migration here and the generated query layer in db/ read
// the same source file. Inlining a Go-string copy of the schema (the
// upstream pattern) creates a drift hazard if either side gets edited
// independently; embed eliminates the second source of truth.
//
//go:embed db/schema.sql
var schemaSQL string

var (
	database *sql.DB
	queries  *db.Queries
)

// InitDB initializes the SQLite database and runs migrations
func InitDB(dbPath string) (*db.Queries, error) {
	var err error

	// Open database connection
	database, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := database.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Run migrations (create tables)
	if err := runMigrations(database); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Create queries instance
	queries = db.New(database)

	log.Printf("Database initialized at: %s", dbPath)
	return queries, nil
}

// runMigrations creates the database schema, handling upgrades from older versions.
func runMigrations(db *sql.DB) error {
	// Check if the todos table exists with an outdated schema (missing user_id column).
	// CREATE TABLE IF NOT EXISTS won't modify an existing table, so we must detect
	// and migrate the old schema before ensuring the current one.
	if needsMigration, err := hasOutdatedSchema(db); err != nil {
		return fmt.Errorf("checking schema: %w", err)
	} else if needsMigration {
		log.Println("Detected outdated todos table (missing user_id column), adding column...")
		if _, err := db.Exec(`ALTER TABLE todos ADD COLUMN user_id TEXT NOT NULL DEFAULT ''`); err != nil {
			return fmt.Errorf("adding user_id column: %w", err)
		}
	}

	_, err := db.Exec(schemaSQL)
	return err
}

// hasOutdatedSchema returns true if the todos table exists but lacks the user_id column.
func hasOutdatedSchema(db *sql.DB) (bool, error) {
	rows, err := db.Query("PRAGMA table_info(todos)")
	if err != nil {
		return false, err
	}
	defer rows.Close()

	var hasTable, hasUserID bool
	for rows.Next() {
		hasTable = true
		var cid int
		var name, ctype string
		var notnull int
		var dfltValue sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
			return false, err
		}
		if name == "user_id" {
			hasUserID = true
		}
	}

	return hasTable && !hasUserID, rows.Err()
}

// CloseDB closes the database connection
func CloseDB() {
	if database != nil {
		if err := database.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		} else {
			log.Println("Database connection closed")
		}
	}
}
