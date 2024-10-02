package migrator

import (
	"context"
	"database/sql"
	"fmt"
	"go-migrations/utilities"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func Migrate(ctx context.Context, opts *Options) (lastestMigration *Migration, err error) {
	fail := func(err error) (*Migration, error) {
		return nil, fmt.Errorf("Migrate: %v", err)
	}

	dbPath := os.Getenv("DB_PATH")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fail(err)
	}
	defer db.Close()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fail(err)
	}

	defer tx.Rollback()

	migrationFiles, err := getAllUpMigrationFiles()
	if err != nil {
		return fail(err)
	}

	currentMigration, err := GetLastMigration(tx, ctx)
	if err != nil {
		return fail(err)
	}

	var migrations Slice
	for _, file := range migrationFiles {
		m, err := createMigrationFromSourceFile(file)
		if err != nil {
			return fail(err)
		}

		migrations = append(migrations, m)
	}
	sort.Sort(migrations)

	if opts.All {
		migrationsToRun := migrations[currentMigration.Version:]
		err := applyMigrations(migrationsToRun, tx, ctx)
		if err != nil {
			return fail(err)
		}
	}

	if opts.Up > 0 {
		migrationsToRun := migrations[currentMigration.Version : currentMigration.Version+opts.Up]
		err := applyMigrations(migrationsToRun, tx, ctx)
		if err != nil {
			return fail(err)
		}
	}

	if opts.Down > 0 {
		migrationsToRun := migrations[currentMigration.Version-opts.Down : currentMigration.Version]
		err := revertMigrations(migrationsToRun, tx, ctx)
		if err != nil {
			return fail(err)
		}
	}

	if opts.To > 0 {
		if currentMigration.Version < opts.To {
			migrationsToRun := migrations[currentMigration.Version:opts.To]
			err := applyMigrations(migrationsToRun, tx, ctx)
			if err != nil {
				return fail(err)
			}
		}

		if currentMigration.Version > opts.To {
			migrationsToRun := migrations[opts.To:currentMigration.Version]
			err := revertMigrations(migrationsToRun, tx, ctx)
			if err != nil {
				return fail(err)
			}
		}
	}

	if opts.Reset && currentMigration.Version > 0 {
		err := revertMigrations(migrations, tx, ctx)
		if err != nil {
			return fail(err)
		}
	}

	lastMigration, err := GetLastMigration(tx, ctx)
	if err != nil {
		return fail(err)
	}

	if err = tx.Commit(); err != nil {
		return fail(err)
	}

	return lastMigration, nil
}

func applyMigrations(migrations Slice, tx *sql.Tx, ctx context.Context) error {
	for _, m := range migrations {
		fmt.Printf("Applying %v: %q \n", m.Version, m.Name)
		_, err := m.Apply(tx, ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func revertMigrations(migrations Slice, tx *sql.Tx, ctx context.Context) error {
	for i := len(migrations) - 1; i >= 0; i-- {
		m := migrations[i]
		fmt.Printf("Reverting %v: %q \n", m.Version, m.Name)
		_, err := m.Revert(tx, ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetLastMigration(tx *sql.Tx, ctx context.Context) (*Migration, error) {
	query := "SELECT version, name FROM migrations ORDER BY version DESC LIMIT 1"
	row := tx.QueryRowContext(ctx, query)
	migration := Migration{Version: 0, Name: ""}

	migrationTableExists, err := utilities.TableExists(tx, "migrations")
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
