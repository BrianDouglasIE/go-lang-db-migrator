package migrations

import (
	"database/sql"
	"fmt"
	"go-migrations/pkg/db_utils"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func Migrate(opts *MigrationOptions) error {
	dbPath := os.Getenv("DB_PATH")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open SQLite database: %v", err)
	}
	defer db.Close()

	migrationFiles, err := getAllUpMigrationFiles()
	if err != nil {
		return err
	}

	currentMigration, err := GetCurrentMigrationVersion(db)
	if err != nil {
		return err
	}

	if currentMigration.Version == 0 {
		fmt.Println("No migrations currently applied.")
	} else {
		fmt.Printf("Current migration version is [%v]: %q. \n", currentMigration.Version, currentMigration.Name)
	}

	var migrations MigrationSlice
	for _, file := range migrationFiles {
		m, err := createMigrationFromSourceFile(file)
		if err != nil {
			return err
		}

		migrations = append(migrations, m)
	}
	sort.Sort(migrations)

	if opts.All {
		migrationsToRun := migrations[currentMigration.Version:]
		return applyMigrations(migrationsToRun, db)
	}

	if opts.Up > 0 {
		migrationsToRun := migrations[currentMigration.Version : currentMigration.Version+opts.Up]
		return applyMigrations(migrationsToRun, db)
	}

	if opts.Down > 0 {
		migrationsToRun := migrations[currentMigration.Version-opts.Down : currentMigration.Version]
		return revertMigrations(migrationsToRun, db)
	}

	if opts.To > 0 {
		if currentMigration.Version < opts.To {
			migrationsToRun := migrations[currentMigration.Version:opts.To]
			return applyMigrations(migrationsToRun, db)
		}

		if currentMigration.Version > opts.To {
			migrationsToRun := migrations[opts.To:currentMigration.Version]
			return revertMigrations(migrationsToRun, db)
		}
	}

	if opts.Clean && currentMigration.Version > 0 {
		return revertMigrations(migrations, db)
	}

	return nil
}

func applyMigrations(migrations MigrationSlice, db *sql.DB) error {
	for _, m := range migrations {
		fmt.Printf("Applying %v: %q \n", m.Version, m.Name)
		_, err := m.Apply(db)
		if err != nil {
			return err
		}
	}
	return nil
}

func revertMigrations(migrations MigrationSlice, db *sql.DB) error {
	for i := len(migrations) - 1; i >= 0; i-- {
		m := migrations[i]
		fmt.Printf("Reverting %v: %q \n", m.Version, m.Name)
		_, err := m.Revert(db)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetCurrentMigrationVersion(db *sql.DB) (*Migration, error) {
	query := "SELECT version, name FROM migrations ORDER BY version DESC LIMIT 1"
	row := db.QueryRow(query)
	migration := Migration{Version: 0, Name: ""}

	migrationTableExists, err := db_utils.TableExists(db, "migrations")
	if err != nil {
		return &migration, err
	}

	if !migrationTableExists {
		return &migration, nil
	}

	err = row.Scan(&migration.Version, &migration.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return &migration, nil
		}

		return nil, err
	}
	return &migration, nil
}

func createMigrationFromSourceFile(sourceFile string) (*Migration, error) {
	filename := filepath.Base(sourceFile)

	parts := strings.Split(filename, "-")
	version, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, err
	}

	return &Migration{
		Version:  version,
		Name:     getMigrationNameFromFilenameParts(parts),
		UpFile:   sourceFile,
		DownFile: strings.Replace(sourceFile, "up.sql", "down.sql", 1),
	}, nil
}

func getMigrationNameFromFilenameParts(parts []string) string {
	var name []string
	for _, part := range parts[1:] {
		if strings.Contains(part, ".sql") {
			name = append(name, strings.Split(part, ".")[0])
			break
		}
		name = append(name, part)
	}
	return strings.Join(name, " ")
}

func getAllUpMigrationFiles() ([]string, error) {
	migrationDir := os.Getenv("MIGRATION_DIR")

	var sqlFiles []string
	err := filepath.Walk(migrationDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.Contains(path, ".up.sql") {
			sqlFiles = append(sqlFiles, path)
		}

		return nil
	})

	return sqlFiles, err
}
