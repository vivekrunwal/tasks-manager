package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	// Get database connection string from environment or use default
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/tasks?sslmode=disable"
	}

	// Connect to the database
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	// Check connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Error pinging database: %v", err)
	}

	fmt.Println("Connected to database successfully")

	// Get migration direction (up or down)
	direction := "up"
	if len(os.Args) > 1 {
		if os.Args[1] == "down" {
			direction = "down"
		}
	}

	// Get migration files
	migrationsDir := "db/migrations"
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		log.Fatalf("Error reading migrations directory: %v", err)
	}

	// Filter and sort migration files
	var migrationFiles []string
	for _, file := range files {
		if !file.IsDir() {
			// Check for both formats: either "0001_init.up.sql" or just "0001_init.up.sql"
			if (direction == "up" && strings.Contains(file.Name(), ".up.sql")) ||
				(direction == "down" && strings.Contains(file.Name(), ".down.sql")) {
				migrationFiles = append(migrationFiles, file.Name())
			}
		}
	}

	if len(migrationFiles) == 0 {
		log.Fatalf("No %s migration files found in %s", direction, migrationsDir)
	}

	// Execute migrations in order
	fmt.Printf("Running %s migrations...\n", direction)
	for _, fileName := range migrationFiles {
		filePath := filepath.Join(migrationsDir, fileName)
		fmt.Printf("Applying migration: %s\n", fileName)

		// Read migration file
		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Fatalf("Error reading migration file %s: %v", fileName, err)
		}

		// Execute migration
		_, err = db.Exec(string(content))
		if err != nil {
			log.Fatalf("Error executing migration %s: %v", fileName, err)
		}

		fmt.Printf("Successfully applied migration: %s\n", fileName)
	}

	fmt.Printf("All %s migrations completed successfully\n", direction)
}
