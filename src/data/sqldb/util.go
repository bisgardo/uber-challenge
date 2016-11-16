package sqldb

import (
	"database/sql"
	"strings"
)

func forEachRow(rows *sql.Rows, callback func(*sql.Rows) error) error {
	defer rows.Close()
	
	for rows.Next() {
		if err := callback(rows); err != nil {
			return err
		}
	}
	
	// Get any error encountered during iteration.
	return rows.Err()
}

func transaction(db *sql.DB, callback func (*sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	// TODO What if rollback returns error?
	defer tx.Rollback()
	
	if err := callback(tx); err != nil {
		return err
	}
	
	return tx.Commit()
}

func escapeSingleQuotes(s string) string {
	return strings.Replace(s, "'", "\\'", -1)
}
