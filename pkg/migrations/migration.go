package migrations

import (
	"database/sql"
	"go-migrations/pkg/db_utils"
	"log"
	"os"
)

type Migration struct {
	Version  int
	Name     string
	UpFile   string
	DownFile string
}

func (m *Migration) Apply(db *sql.DB) (sql.Result, error) {
	content, err := m.getUpMigrationFileContent()
	if err != nil {
		return nil, err
	}

	_, err = m.executeMigrationFile(content, db)
	if err != nil {
		return nil, err
	}

	return m.storeMigration(db)
}

func (m *Migration) Revert(db *sql.DB) (sql.Result, error) {
	content, err := m.getDownMigrationFileContent()
	if err != nil {
		return nil, err
	}
	_, err = m.executeMigrationFile(content, db)
	if err != nil {
		return nil, err
	}

	migrationTableExists, err := db_utils.TableExists(db, "migrations")
	if err != nil {
		return nil, err
	}

	if migrationTableExists {
		return m.deleteMigration(db)
	}

	return nil, nil
}

func (m *Migration) IsApplied(db *sql.DB) bool {
	query := "SELECT 1 FROM migrations WHERE version = ? LIMIT 1"
	row := db.QueryRow(query, m.Version)
	exists, err := db_utils.RowExists(row)
	if err != nil {
		log.Printf("error checking if migration is applied: [%s] %s", m.Name, err)
	}
	return exists
}

func (m *Migration) executeMigrationFile(content []byte, db *sql.DB) (sql.Result, error) {
	return db.Exec(string(content))
}

func (m *Migration) getUpMigrationFileContent() ([]byte, error) {
	return os.ReadFile(m.UpFile)
}

func (m *Migration) getDownMigrationFileContent() ([]byte, error) {
	return os.ReadFile(m.DownFile)
}

func (m *Migration) storeMigration(db *sql.DB) (sql.Result, error) {
	query := "INSERT INTO migrations (version, name) VALUES (?, ?)"
	return db.Exec(query, m.Version, m.Name)
}

func (m *Migration) deleteMigration(db *sql.DB) (sql.Result, error) {
	query := "DELETE FROM migrations WHERE version = ?"
	return db.Exec(query, m.Version)
}
