package data

import "database/sql"

func Insert(tx *sql.Tx, sql string, args ...interface{}) (int64, error) {
	r, err := tx.Exec(sql, args...)
	if err != nil {
		return -1, err
	}
	id, err := r.LastInsertId()
	if err != nil {
		return -1, err
	}
	return id, nil
}

func ForEachRow(rows *sql.Rows, callback func(*sql.Rows) error) error {
	defer rows.Close()
	
	for rows.Next() {
		if err := callback(rows); err != nil {
			return err
		}
	}
	
	// Get any error encountered during iteration.
	return rows.Err()
}

func QuerySingleInt(db *sql.DB, sql string, args ...interface{}) (int, error) {
	row := db.QueryRow(sql, args...)
	var i int
	err := row.Scan(&i)
	return i, err
}
