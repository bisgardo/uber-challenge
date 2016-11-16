package sqldb

import (
	"src/data/types"
	"src/logging"
	"src/watch"
	"database/sql"
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

func StoreMovies(tx *sql.Tx, ms []types.Movie, logger logging.Logger) error {
	if len(ms) == 0 {
		return nil
	}
	
	logger.Infof("Inserting %d movies into database", len(ms))
	
	sw := watch.NewStopWatch()
	
	// Batch insert movies.
	movieBI := NewBulkInserter(7)
	for _, m := range ms {
		movieBI.Add(nil, m.Title, m.Writer, m.Director, m.Distributor, m.ProductionCompany, m.ReleaseYear)
	}
	
	if _, err := movieBI.Exec(tx, "movies", nil); err != nil {
		return err
	}
	
	logger.Infof("Inserted %d movies in %d ms", len(ms), sw.ElapsedTimeMillis(true))
	
	// Query movies in order to get their IDs.
	movieTitleIdMap, err := movieTitleIdMap(tx)
	if err != nil {
		return err
	}
	
	// Bulk insert locations.
	locationBI := NewBulkInserter(4)
	locationCount := 0
	for _, m := range ms {
		mId := movieTitleIdMap[m.Title]
		for _, l := range m.Locations {
			locationBI.Add(nil, mId, l.Name, l.FunFact)
			locationCount++
		}
	}
	
	if _, err := locationBI.Exec(tx, "locations", nil); err != nil {
		return err
	}
	
	logger.Infof("Inserted %d locations in %d ms", locationCount, sw.ElapsedTimeMillis(true))
	
	// Bulk insert actors.
	
	actorsBI := NewBulkInserter(2)
	as := make(map[string]bool)
	for _, m := range ms {
		for _, a := range m.Actors {
			if _, exists := as[a]; !exists {
				as[a] = true
				actorsBI.Add(nil, a)
			}
		}
	}
	
	if _, err := actorsBI.Exec(tx, "actors", nil); err != nil {
		return err
	}
	
	logger.Infof("Inserted %d actors in %d ms", len(as), sw.ElapsedTimeMillis(true))
	
	// Query actors in order to get their IDs.
	actorIdMap, err := actorIdMap(tx)
	if err != nil {
		return err
	}
	
	// Bulk insert movie-actor relations.
	relationBI := NewBulkInserter(2)
	movieActorCount := 0
	for _, m := range ms {
		mId := movieTitleIdMap[m.Title]
		for _, a := range m.Actors {
			aId := actorIdMap[a]
			relationBI.Add(mId, aId)
			movieActorCount++
		}
	}
	
	if _, err := relationBI.Exec(tx, "movies_actors", nil); err != nil {
		return err
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
		
		bi := NewBulkInserter(2)
		
		for t, i := range movieInfo {
			bi.Add(t, i)
		}
		
		if _, err := bi.Exec(tx, "movie_info", nil); err != nil {
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
		
		bi := NewBulkInserter(3)
		
		for n, c := range lc {
			if c == nil {
				continue
			}
			
			bi.Add(n, c.Lat, c.Lng)
		}
		
		if _, err := bi.Exec(tx, "coordinates", nil); err != nil {
			return err
		}
		
		logger.Infof("Inserted %d location coordinate pairs in %d ms", len(lc), sw.TotalElapsedTimeMillis())
		return nil
	})
}
