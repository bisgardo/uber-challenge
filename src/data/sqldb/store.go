package sqldb

import (
	"src/data/types"
	"src/logging"
	"src/watch"
	"database/sql"
)

func InitTablesAndStoreMovies(db *sql.DB, movies []types.Movie, log logging.Logger) error {
	return transaction(db, func (tx *sql.Tx) error {
		if err := InitTables(tx, log); err != nil {
			return err
		}
		if err := StoreMovies(tx, movies, log); err != nil {
			return err
		}
		return nil
	})
}

func StoreMovies(tx *sql.Tx, movies []types.Movie, log logging.Logger) error {
	if len(movies) == 0 {
		return nil
	}
	
	log.Infof("Inserting %d movies into database", len(movies))
	
	sw := watch.NewStopWatch()
	
	// Batch insert movies.
	movieInserter := NewBulkInserter(7)
	for _, movie := range movies {
		movieInserter.Add(nil, movie.Title, movie.Writer, movie.Director, movie.Distributor, movie.ProductionCompany, movie.ReleaseYear)
	}
	
	if _, err := movieInserter.Exec(tx, "movies", nil); err != nil {
		return err
	}
	
	log.Infof("Inserted %d movies in %d ms", len(movies), sw.ElapsedTimeMillis(true))
	
	// Query movies in order to get their IDs.
	movieTitleIdMap, err := loadMovieTitleIdMap(tx)
	if err != nil {
		return err
	}
	
	// Bulk insert locations.
	locationInserter := NewBulkInserter(4)
	locationCount := 0
	for _, movie := range movies {
		id := movieTitleIdMap[movie.Title]
		for _, loc := range movie.Locations {
			locationInserter.Add(nil, id, loc.Name, loc.FunFact)
			locationCount++
		}
	}
	
	if _, err := locationInserter.Exec(tx, "locations", nil); err != nil {
		return err
	}
	
	log.Infof("Inserted %d locations in %d ms", locationCount, sw.ElapsedTimeMillis(true))
	
	// Bulk insert actors.
	
	actorInserter := NewBulkInserter(2)
	actorNames := make(map[string]bool)
	for _, movie := range movies {
		for _, actorName := range movie.Actors {
			if _, exists := actorNames[actorName]; !exists {
				actorNames[actorName] = true
				actorInserter.Add(nil, actorName)
			}
		}
	}
	
	if _, err := actorInserter.Exec(tx, "actors", nil); err != nil {
		return err
	}
	
	log.Infof("Inserted %d actors in %d ms", len(actorNames), sw.ElapsedTimeMillis(true))
	
	// Query actors in order to get their IDs.
	actorIdMap, err := loadActorIdMap(tx)
	if err != nil {
		return err
	}
	
	// Bulk insert movie-actor relations.
	movieActorInserter := NewBulkInserter(2)
	movieActorCount := 0
	for _, movie := range movies {
		movieId := movieTitleIdMap[movie.Title]
		for _, actorName := range movie.Actors {
			actorId := actorIdMap[actorName]
			movieActorInserter.Add(movieId, actorId)
			movieActorCount++
		}
	}
	
	if _, err := movieActorInserter.Exec(tx, "movies_actors", nil); err != nil {
		return err
	}
	
	log.Infof("Inserted %d movie-actor relations in %d ms", movieActorCount, sw.ElapsedTimeMillis(true))
	
	log.Infof("Database updated in %d ms", sw.TotalElapsedTimeMillis())
	
	return nil
}

func loadMovieTitleIdMap(tx *sql.Tx) (map[string]int64, error) {
	rows, err := tx.Query("SELECT title, id FROM movies")
	if err != nil {
		return nil, err
	}
	
	movieTitleIdMap := make(map[string]int64)
	err = forEachRow(rows, func (rows *sql.Rows) error {
		var title string
		var id int64
		if err := rows.Scan(&title, &id); err != nil {
			return err
		}
		movieTitleIdMap[title] = id
		return nil
	})
	return movieTitleIdMap, err
}

func loadActorIdMap(tx *sql.Tx) (map[string]int64, error) {
	rows, err := tx.Query("SELECT name, id FROM actors")
	if err != nil {
		return nil, err
	}
	
	actorIdMap := make(map[string]int64)
	err = forEachRow(rows, func (rows *sql.Rows) error {
		var actorName string
		var id int64
		if err := rows.Scan(&actorName, &id); err != nil {
			return err
		}
		actorIdMap[actorName] = id
		return nil
	})
	return actorIdMap, err
}

func StoreMovieInfo(db *sql.DB, movieInfo map[string]string, log logging.Logger) error {
	if len(movieInfo) == 0 {
		return nil
	}
	
	sw := watch.NewStopWatch()
	
	log.Infof("Inserting %d movie infos into database", len(movieInfo))
	
	err := transaction(db, func (tx *sql.Tx) error {
		inserter := NewBulkInserter(2)
		
		for t, i := range movieInfo {
			inserter.Add(t, i)
		}
		
		if _, err := inserter.Exec(tx, "movie_info", nil); err != nil {
			return err
		}
		
		return nil
	})
	if err != nil {
		return err
	}
	
	log.Infof("Inserted %d movie infos in %d ms", len(movieInfo), sw.TotalElapsedTimeMillis())
	return nil
}

func StoreCoordinates(db *sql.DB, lc map[string]*types.Coordinates, log logging.Logger) error {
	if len(lc) == 0 {
		return nil
	}
	
	sw := watch.NewStopWatch()
	
	log.Infof("Inserting %d location coordinates into database", len(lc))
	
	err := transaction(db, func (tx *sql.Tx) error {
		inserter := NewBulkInserter(3)
		
		for n, c := range lc {
			if c == nil {
				continue
			}
			
			inserter.Add(n, c.Lat, c.Lng)
		}
		
		_, err := inserter.Exec(tx, "coordinates", nil)
		return err
	})
	if err != nil {
		return err
	}
	
	log.Infof("Inserted %d location coordinate pairs in %d ms", len(lc), sw.TotalElapsedTimeMillis())
	return nil
}
