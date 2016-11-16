package sqldb

import (
	"database/sql"
	"strings"
	"fmt"
	"log"
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

type BatchInsertStmtBuilder struct {
	colCount   int
	prpStmtStr string
	valuesStmt string
	values     []interface{}
}

func NewBulkInserter(colCount int) BatchInsertStmtBuilder {
	qstComRep := strings.Repeat("?,", colCount)
	prpStmtStr := "(" + qstComRep[:len(qstComRep)-1] + ")"
	
	log.Println(prpStmtStr)
	
	return BatchInsertStmtBuilder{colCount: colCount, prpStmtStr: prpStmtStr}
}

func (b *BatchInsertStmtBuilder) Add(values ...interface{}) *BatchInsertStmtBuilder {
	if len(values) != b.colCount {
		panic(fmt.Sprintf("Expected %d values but got %d", b.colCount, len(values)))
	}
	
	// TODO Although this isn't a performance bottleneck, string should be built in a buffer ala Java's `StringBuilder`.
	if len(b.valuesStmt) > 0 {
		b.valuesStmt += ","
	}
	b.valuesStmt += b.prpStmtStr
	
	b.values = append(b.values, values...)
	
	return b
}

func (b *BatchInsertStmtBuilder) build(tableName string) string {
	if len(b.values) == 0 {
		return "";
	}
	return "INSERT INTO " + tableName + " VALUES" + b.valuesStmt
}

func (b *BatchInsertStmtBuilder) Exec(tx *sql.Tx, tableName string, logger logging.Logger) (sql.Result, error) {
	if len(b.values) == 0 {
		return nil, nil
	}
	
	stmt := b.build(tableName)
	if logger != nil {
		logger.Debugf("Executing query '%s' with values %s", stmt, fmt.Sprintln(b.values))
	}
	return tx.Exec(stmt, b.values...)
}
