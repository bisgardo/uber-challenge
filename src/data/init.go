package data

import (
	"src/data/fetch"
	"src/data/sqldb"
	"src/logging"
	"sync"
	"database/sql"
)

var InitUpdateMutex = &sync.Mutex{}

func Init(db *sql.DB, filename string, logger logging.Logger) (bool, error) {
	InitUpdateMutex.Lock()
	defer InitUpdateMutex.Unlock()
	
	initialized, err := IsInitialized(db)
	if err != nil {
		return false, err
	}
	
	if initialized {
		logger.Infof("Database is already initialized")
		return false, nil
	}
	
	logger.Infof("Initializing database from cached file...")
	
	ms, err := fetch.FetchFromFile(filename)
	if err != nil {
		return true, err
	}
	
	return true, sqldb.InitTablesAndInsertMovies(db, ms, logger)
}

func IsInitialized(db *sql.DB) (bool, error) {
	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		return false, err
	}
	defer rows.Close()
	return rows.Next(), rows.Err()
}
