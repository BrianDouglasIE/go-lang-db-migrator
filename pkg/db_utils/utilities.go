package db_utils

import "database/sql"

func RowExists(row *sql.Row) (bool, error) {
	var exists int
	if err := row.Scan(&exists); err != nil {
		if err != sql.ErrNoRows {
			return false, err
		}
		return false, nil
	}
	return true, nil
}

func TableExists(db *sql.DB, tableName string) (bool, error) {
	query := "SELECT name FROM sqlite_master WHERE type='table' AND name=?"
	row := db.QueryRow(query, tableName)

	var name string
	err := row.Scan(&name)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
