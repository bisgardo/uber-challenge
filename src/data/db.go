package data

import (
	"src/logging"
	"database/sql"
	"sync"
)

type LocationDb struct {
	SqlDb *sql.DB
}

var InitUpdateMutex *sync.Mutex = &sync.Mutex{}

func Open(driverName string, dataSourceName string) (*LocationDb, error) {
	sqlDb, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}
	return &LocationDb{sqlDb}, nil
}

func (db *LocationDb) Init(filename string, logger logging.Logger) (bool, error) {
	InitUpdateMutex.Lock()
	defer InitUpdateMutex.Unlock()
	
	initialized, err := db.IsInitialized()
	if err != nil {
		return false, err
	}
	
	if initialized {
		logger.Infof("Database is already initialized")
		return false, nil
	}
	
	logger.Infof("Initializing database from cached file...")
	
	ms, err := FetchFromFile(filename)
	if err != nil {
		return true, err
	}
	
	err = db.transaction(func (tx *sql.Tx) error {
		if err := InitTables(tx, logger); err != nil {
			return err
		}
		if err := InsertMoviesMoreOptimized(tx, ms, logger); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return true, err
	}
	return true, nil
}

func (db *LocationDb) transaction(callback func (*sql.Tx) error) error {
	tx, err := db.SqlDb.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	if err := callback(tx); err != nil {
		return err
	}
	
	return tx.Commit()
}

func (db *LocationDb) IsInitialized() (bool, error) {
	rows, err := db.SqlDb.Query("SHOW TABLES")
	if err != nil {
		return false, err
	}
	defer rows.Close()
	return rows.Next(), rows.Err()
}
