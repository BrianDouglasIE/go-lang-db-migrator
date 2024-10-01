package migrator

import (
	"context"
	"database/sql"
	"go-migrations/utilities"
	"os"
)

type Migration struct {
	Version  int
	Name     string
	UpFile   string
	DownFile string
}

func (m *Migration) Apply(tx *sql.Tx, ctx context.Context) (sql.Result, error) {
	content, err := m.getUpMigrationFileContent()
	if err != nil {
		return nil, err
	}

	_, err = m.executeMigrationFile(content, tx, ctx)
	if err != nil {
		return nil, err
	}

	return m.storeMigration(tx, ctx)
}

func (m *Migration) Revert(tx *sql.Tx, ctx context.Context) (sql.Result, error) {
	content, err := m.getDownMigrationFileContent()
	if err != nil {
		return nil, err
	}

	_, err = m.executeMigrationFile(content, tx, ctx)
	if err != nil {
		return nil, err
	}

	migrationTableExists, err := utilities.TableExists(tx, "migrations")
	if err != nil {
		return nil, err
	}

	if migrationTableExists {
		return m.deleteMigration(tx, ctx)
	}

	return nil, nil
}

func (m *Migration) executeMigrationFile(content []byte, tx *sql.Tx, ctx context.Context) (sql.Result, error) {
	return tx.ExecContext(ctx, string(content))
}

func (m *Migration) getUpMigrationFileContent() ([]byte, error) {
	return os.ReadFile(m.UpFile)
}

func (m *Migration) getDownMigrationFileContent() ([]byte, error) {
	return os.ReadFile(m.DownFile)
}

func (m *Migration) storeMigration(tx *sql.Tx, ctx context.Context) (sql.Result, error) {
	query := "INSERT INTO migrations (version, name) VALUES (?, ?)"
	return tx.ExecContext(ctx, query, m.Version, m.Name)
}

func (m *Migration) deleteMigration(tx *sql.Tx, ctx context.Context) (sql.Result, error) {
	query := "DELETE FROM migrations WHERE version = ?"
	return tx.ExecContext(ctx, query, m.Version)
}
