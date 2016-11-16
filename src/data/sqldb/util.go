package sqldb

import (
	"database/sql"
	"strings"
	"fmt"
	"src/logging"
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

func transaction(db *sql.DB, callback func(*sql.Tx) error) error {
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

type BulkInsertStmtBuilder struct {
	colCount int
	rowCount int
	values   []interface{}
}

func NewBulkInserter(colCount int) BulkInsertStmtBuilder {
	return BulkInsertStmtBuilder{colCount: colCount, rowCount: 0}
}

func (b *BulkInsertStmtBuilder) Add(values ...interface{}) *BulkInsertStmtBuilder {
	if len(values) != b.colCount {
		panic(fmt.Sprintf("Expected %d values but got %d", b.colCount, len(values)))
	}
	
	b.values = append(b.values, values...)
	b.rowCount++
	
	// Allow chaining.
	return b
}

func (b *BulkInsertStmtBuilder) build(tableName string) string {
	if (b.rowCount * b.colCount != len(b.values)) {
		panic("Unexpected number of values...")
	}
	
	if len(b.values) == 0 {
		return "";
	}
	
	// Construct string with format "(?, ?, ..., ?)".
	prpStmtStr := fancyRepeat("(", "?", b.colCount, ",", ")")
	
	// Construct string with format "INSERT INTO table VALUES prpStmtStr, prpStmtStr, ..., prpStmtStr".
	return fancyRepeat("INSERT INTO " + tableName + " VALUES", prpStmtStr, b.rowCount, ",", "")
}

func (b *BulkInsertStmtBuilder) Exec(tx *sql.Tx, tableName string, logger logging.Logger) (sql.Result, error) {
	if len(b.values) == 0 {
		return nil, nil
	}
	
	stmt := b.build(tableName)
	if logger != nil {
		logger.Debugf("Executing query '%s' with values %s", stmt, fmt.Sprintln(b.values))
	}
	return tx.Exec(stmt, b.values...)
}

func fancyRepeat(prefix string, rep string, count int, sep string, suffix string) string {
	repeated := strings.Repeat(rep + sep, count)
	return prefix + repeated[:len(repeated) - len(sep)] + suffix
}
