package sqldb

import (
	"src/data/types"
	"src/logging"
	"src/watch"
	"database/sql"
	"fmt"
	"strings"
)

// TODO Turn bulk insertion stuff into nice utility (that is safe regarding injection and missing data) and replace current solutions with it...

func InitTablesAndStoreMovies(db *sql.DB, ms []types.Movie, logger logging.Logger) error {
	return transaction(db, func (tx *sql.Tx) error {
		if err := InitTables(tx, logger); err != nil {
			return err
		}
		if err := StoreMovies(tx, ms, logger); err != nil {
			return err
		}
		return nil
	})
}

func escapeSingleQuotes(s string) string {
	return strings.Replace(s, "'", "\\'", -1)
}

func StoreMovies(tx *sql.Tx, ms []types.Movie, logger logging.Logger) error {
	if len(ms) == 0 {
		return nil
	}
	
	logger.Infof("Inserting %d movies into database", len(ms))
	
	sw := watch.NewStopWatch()
	
	// Batch insert movies.
	insertMoviesVals := ""
	for _, m := range ms {
		// TODO Make robust towards injection!
		insertMovieVal := fmt.Sprintf(
			"\n(NULL, '%s', '%s', '%s', '%s', '%s', %d)",
			escapeSingleQuotes(m.Title),
			escapeSingleQuotes(m.Writer),
			escapeSingleQuotes(m.Director),
			escapeSingleQuotes(m.Distributor),
			escapeSingleQuotes(m.ProductionCompany),
			m.ReleaseYear,
		)
		if len(insertMoviesVals) > 0 {
			insertMoviesVals += ","
		}
		insertMoviesVals += insertMovieVal
	}
	
	if len(ms) > 0 {
		insertMoviesStmt := "INSERT INTO movies VALUES" + insertMoviesVals
		//logger.Debugf("Executing query %s", insertMoviesStmt)
		if _, err := tx.Exec(insertMoviesStmt); err != nil {
			return err
		}
	}
	
	logger.Infof("Inserted %d movies in %d ms", len(ms), sw.ElapsedTimeMillis(true))
	
	// Query movies in order to get their IDs.
	movieTitleIdMap, err := movieTitleIdMap(tx)
	if err != nil {
		return err
	}
	
	// Bulk insert locations.
	
	insertLocationsVals := ""
	locationCount := 0
	
	for _, m := range ms {
		mId := movieTitleIdMap[m.Title]
		for _, l := range m.Locations {
			// TODO Make robust towards injection!
			insertLocationsVal := fmt.Sprintf(
				"\n(NULL, %d, '%s', '%s')",
				mId,
				escapeSingleQuotes(l.Name),
				escapeSingleQuotes(l.FunFact),
			)
			
			if len(insertLocationsVals) > 0 {
				insertLocationsVals += ","
			}
			insertLocationsVals += insertLocationsVal
			locationCount++
		}
	}
	
	if locationCount > 0 {
		insertLocationsStmt := "INSERT INTO locations VALUES" + insertLocationsVals
		//logger.Debugf("Executing query %s", insertLocationsStmt)
		if _, err := tx.Exec(insertLocationsStmt); err != nil {
			return err
		}
	}
	
	logger.Infof("Inserted %d locations in %d ms", locationCount, sw.ElapsedTimeMillis(true))
	
	// Bulk insert actors.
	
	insertActorsVals := ""
	as := make(map[string]bool)
	for _, m := range ms {
		for _, a := range m.Actors {
			if _, exists := as[a]; !exists {
				as[a] = true
				// TODO Make robust towards injection!
				insertActorVal := fmt.Sprintf("\n(NULL, '%s')", escapeSingleQuotes(a))
				
				if len(insertActorsVals) > 0 {
					insertActorsVals += ","
				}
				insertActorsVals += insertActorVal
			}
		}
	}
	
	if len(as) > 0 {
		insertActorsStmt := "INSERT INTO actors VALUES" + insertActorsVals
		//logger.Debugf("Executing query %s", insertActorsStmt)
		if _, err := tx.Exec(insertActorsStmt); err != nil {
			return err
		}
	}
	
	logger.Infof("Inserted %d actors in %d ms", len(as), sw.ElapsedTimeMillis(true))
	
	// Query actors in order to get their IDs.
	actorIdMap, err := actorIdMap(tx)
	if err != nil {
		return err
	}
	
	// Bulk insert movie-actor relations.
	insertMovieActorVals := ""
	movieActorCount := 0
	for _, m := range ms {
		mId := movieTitleIdMap[m.Title]
		for _, a := range m.Actors {
			aId := actorIdMap[a]
			insertMovieActorVal := fmt.Sprintf("\n(%d, %d)", mId, aId)
			
			if len(insertMovieActorVals) > 0 {
				insertMovieActorVals += ","
			}
			insertMovieActorVals += insertMovieActorVal
			movieActorCount++
		}
	}
	
	if movieActorCount > 0 {
		insertMovieActorStmt := "INSERT INTO movies_actors VALUES" + insertMovieActorVals
		//logger.Debugf("Executing query %s", insertMovieActorStmt)
		if _, err := tx.Exec(insertMovieActorStmt); err != nil {
			return err
		}
	}
	
	logger.Infof("Inserted %d movie-actor relations in %d ms", movieActorCount, sw.ElapsedTimeMillis(true))
	
	logger.Infof("Updated database in %d ms", sw.TotalElapsedTimeMillis())
	
	return nil
}

func movieTitleIdMap(tx *sql.Tx) (map[string]int64, error) {
	rows, err := tx.Query("SELECT title, id FROM movies")
	if err != nil {
		return nil, err
	}
	
	m := make(map[string]int64)
	err = forEachRow(rows, func (rows *sql.Rows) error {
		var title string
		var id int64
		if err := rows.Scan(&title, &id); err != nil {
			return err
		}
		m[title] = id
		return nil
	})
	return m, err
}

func actorIdMap(tx *sql.Tx) (map[string]int64, error) {
	rows, err := tx.Query("SELECT name, id FROM actors")
	if err != nil {
		return nil, err
	}
	
	m := make(map[string]int64)
	err = forEachRow(rows, func (rows *sql.Rows) error {
		var name string
		var id int64
		if err := rows.Scan(&name, &id); err != nil {
			return err
		}
		m[name] = id
		return nil
	})
	return m, err
}

func StoreMovieInfo(db *sql.DB, movieInfo map[string]string, logger logging.Logger) error {
	if len(movieInfo) == 0 {
		return nil
	}
	
	logger.Infof("Inserting %d movie infos into database", len(movieInfo))
	
	return transaction(db, func (tx *sql.Tx) error {
		sw := watch.NewStopWatch()
		
		vals := ""
		
		for t, i := range movieInfo {
			// TODO Make robust towards injection!
			val := fmt.Sprintf("\n('%s', '%s')", escapeSingleQuotes(t), escapeSingleQuotes(i))
			if len(vals) > 0 {
				vals += ","
			}
			vals += val
			//_, err := Insert(tx, "INSERT INTO movie_info VALUES (?, ?)", t, i)
			//if err != nil {
			//	return err
			//}
		}
		
		insertStmt := "INSERT INTO movie_info VALUES" + vals
		logger.Debugf("Executing query %s", insertStmt)
		if _, err := tx.Exec(insertStmt); err != nil {
			return err
		}
		
		logger.Infof("Inserted %d movie infos in %d ms", len(movieInfo), sw.TotalElapsedTimeMillis())
		return nil
	})
}

func StoreCoordinates(db *sql.DB, lc map[string]*types.Coordinates, logger logging.Logger) error {
	if len(lc) == 0 {
		return nil
	}
	
	logger.Infof("Inserting %d location coordinates into database", len(lc))
	
	return transaction(db, func (tx *sql.Tx) error {
		sw := watch.NewStopWatch()
		
		vals := ""
		
		for n, c := range lc {
			if c == nil {
				continue
			}
			// TODO Make robust towards injection!
			val := fmt.Sprintf("\n('%s', %f, %f)", escapeSingleQuotes(n), c.Lat, c.Lng)
			if len(vals) > 0 {
				vals += ","
			}
			vals += val
		}
		
		// Need check because map may contain only 'nil' values.
		if vals != "" {
			insertStmt := "INSERT INTO coordinates VALUES" + vals
			logger.Debugf("Executing query %s", insertStmt)
			if _, err := tx.Exec(insertStmt); err != nil {
				return err
			}
		}
		
		logger.Infof("Inserted %d location coordinate pairs in %d ms", len(lc), sw.TotalElapsedTimeMillis())
		return nil
	})
}
