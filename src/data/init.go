package data

import (
	"src/data/fetch"
	"src/data/sqldb"
	"src/logging"
	"sync"
	"database/sql"
)

var InitUpdateMutex = &sync.Mutex{}

func Init(db *sql.DB, filename string, log logging.Logger) (bool, error) {
	InitUpdateMutex.Lock()
	defer InitUpdateMutex.Unlock()
	
	alreadyInitialized, err := IsInitialized(db)
	if err != nil {
		return !alreadyInitialized, err
	}
	
	if alreadyInitialized {
		log.Infof("Database is already initialized")
		return false, nil
	}
	
	// Database is uninitialized. Try and initialize it...
	
	log.Infof("Initializing database from cached file...")
	
	movies, err := fetch.FetchFromFile(filename)
	if err != nil {
		return true, err
	}
	
	return true, sqldb.InitTablesAndStoreMovies(db, movies, log)
}

func IsInitialized(db *sql.DB) (bool, error) {
	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		return false, err
	}
	defer rows.Close()
	return rows.Next(), rows.Err()
}
