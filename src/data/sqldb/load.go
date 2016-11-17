package sqldb

import (
	"src/data/types"
	"src/logging"
	"src/watch"
	"sort"
	"database/sql"
)

func LoadMovie(db *sql.DB, id int64, log logging.Logger) (types.Movie, error) {
	var movie types.Movie
	err := transaction(db, func (tx *sql.Tx) error {
		row := tx.QueryRow("SELECT title, writer, director, distributor, production_company, release_year FROM movies WHERE id = ?", id)
		
		err := row.Scan(
			&movie.Title,
			&movie.Writer,
			&movie.Director,
			&movie.Distributor,
			&movie.ProductionCompany,
			&movie.ReleaseYear,
		)
		if err != nil {
			return err
		}
		
		if err := LoadLocations(tx, id, &movie.Locations, log); err != nil {
			return err
		}
		
		if err := LoadActors(tx, id, &movie.Actors, log); err != nil {
			return err
		}
		
		return nil
	})
	return movie, err
}

func LoadLocations(tx *sql.Tx, id int64, locs *[]types.Location, log logging.Logger) error {
	log.Debugf("Querying locations for movie %d", id)
	
	rows, err := tx.Query("SELECT name, fun_fact FROM locations AS l WHERE l.movie_id = ?", id)
	if err != nil {
		return err
	}
	
	return forEachRow(rows, func (rows *sql.Rows) error {
		var loc types.Location
		err := rows.Scan(&loc.Name, &loc.FunFact)
		if err != nil {
			return err
		}
		
		*locs = append(*locs, loc)
		return nil
	})
}

func LoadActors(tx *sql.Tx, movieId int64, actors *[]string, log logging.Logger) error {
	log.Debugf("Querying actors for movie %d", movieId)
	
	rows, err := tx.Query(
		"SELECT a.name FROM actors AS a, movies_actors AS r WHERE a.id = r.actor_id AND r.movie_id = ?",
		movieId,
	)
	if err != nil {
		return err
	}
	
	return forEachRow(rows, func (rows *sql.Rows) error {
		var actorName string
		err := rows.Scan(&actorName)
		if err != nil {
			return err
		}
		
		*actors = append(*actors, actorName)
		return nil
	})
}

func LoadMovies(db *sql.DB, log logging.Logger) ([]types.IdMoviePair, error) {
	var movies []types.IdMoviePair
	
	err := transaction(db, func (tx *sql.Tx) error {
		log.Debugf("Querying movies")
		
		// Loading all movies.
		rows, err := tx.Query("SELECT id, title, writer, director, distributor, production_company, release_year FROM movies")
		if err != nil {
			return err
		}
		
		idMovieMap := make(map[int64]*types.Movie)
		err = forEachRow(rows, func (rows *sql.Rows) error {
			var id int64
			var movie types.Movie
			
			err := rows.Scan(
				&id,
				&movie.Title,
				&movie.Writer,
				&movie.Director,
				&movie.Distributor,
				&movie.ProductionCompany,
				&movie.ReleaseYear,
			);
			if err != nil {
				return err
			}
			
			idMovieMap[id] = &movie
			return nil
		})
		if err != nil {
			return err
		}
		
		// Load all locations.
		if err := LoadAllLocations(tx, idMovieMap, log); err != nil {
			return err
		}
		
		// Load all actors.
		if err := LoadAllActors(tx, idMovieMap, log); err != nil {
			return err
		}
		
		movies = make([]types.IdMoviePair, 0, len(idMovieMap))
		for mId, m := range idMovieMap {
			movies = append(movies, types.IdMoviePair{Id: mId, Movie: *m})
		}
		
		return nil
	})
	if err != nil {
		return nil, err
	}
	
	sort.Sort(types.ByTitle(movies))
	return movies, nil
}

func LoadAllLocations(tx *sql.Tx, idMovieMap map[int64]*types.Movie, log logging.Logger) error {
	log.Debugf("Querying all locations")
	
	rows, err := tx.Query("SELECT movie_id, name, fun_fact FROM locations")
	if err != nil {
		return err
	}
	
	return forEachRow(rows, func (rows *sql.Rows) error {
		var id int64
		var loc types.Location
		err := rows.Scan(&id, &loc.Name, &loc.FunFact)
		if err != nil {
			return err
		}
		
		movie, exists := idMovieMap[id]
		if !exists {
			panic("Unexpected location ID...")
		}
		movie.Locations = append(movie.Locations, loc)
		return nil
	})
}

func LoadAllActors(tx *sql.Tx, idMovieMap map[int64]*types.Movie, log logging.Logger) error {
	log.Debugf("Querying all actors")
	
	var rows *sql.Rows
	var err error
	rows, err = tx.Query("SELECT id, name FROM actors")
	if err != nil {
		return err
	}
	
	idActorMap := make(map[int64]string)
	err = forEachRow(rows, func (rows *sql.Rows) error {
		var id int64
		var actorName string
		if err := rows.Scan(&id, &actorName); err != nil {
			return err
		}
		
		idActorMap[id] = actorName
		return nil
	})
	if err != nil {
		return err
	}
	
	rows, err = tx.Query("SELECT movie_id, actor_id FROM movies_actors")
	if err != nil {
		return err
	}
	
	err = forEachRow(rows, func (rows *sql.Rows) error {
		var movieId int64
		var actorId int64
		if err := rows.Scan(&movieId, &actorId); err != nil {
			return err
		}
		
		movie, exists := idMovieMap[movieId]
		if !exists {
			panic("Unexpected movie ID...")
		}
		
		actorName, exists := idActorMap[actorId]
		if !exists {
			panic("Unexpected actor ID...")
		}
		
		movie.Actors = append(movie.Actors, actorName)
		return nil
	})
	if err != nil {
		return err
	}
	
	return nil
}

func LoadMovieInfoJson(db *sql.DB, title string, log logging.Logger) (string, error) {
	sw := watch.NewStopWatch()
	
	var info string
	err := transaction(db, func (tx *sql.Tx) error {
		row := tx.QueryRow("SELECT info_json FROM movie_info WHERE movie_title = ?", title)
		return row.Scan(&info)
	})
	
	if err != nil {
		log.Infof("Loaded info for movie '%s' in %d ms", title, sw.TotalElapsedTimeMillis())
	}
	
	return info, err
}

func LoadMovieInfoJsons(db *sql.DB, log logging.Logger) (map[string]string, error) {
	// TODO Parallelize (if the API allows it) and consider using memcached (with expiration) instead of SQL.
	
	sw := watch.NewStopWatch()
	movieInfo := make(map[string]string)
	err := transaction(db, func (tx *sql.Tx) error {
		rows, err := tx.Query("SELECT movie_title, info_json FROM movie_info")
		if err != nil {
			return err
		}
		
		return forEachRow(rows, func (rows *sql.Rows) error {
			var movieTitle string
			var infoJson string
			if err := rows.Scan(&movieTitle, &infoJson); err != nil {
				return err
			}
			movieInfo[movieTitle] = infoJson
			return nil
		})
	})
	
	if err == nil {
		log.Infof("Fetched info for %d movies in %d ms", len(movieInfo), sw.TotalElapsedTimeMillis())
	}
	
	return movieInfo, err
}

func LoadCoordinates(db *sql.DB, locs []types.Location, log logging.Logger) (map[string]types.Coordinates, error) {
	sw := watch.NewStopWatch()
	
	locNames := make([]interface{}, 0, len(locs))
	for _, loc := range locs {
		locName := loc.Name
		locNames = append(locNames, locName)
	}
	
	locCoords := make(map[string]types.Coordinates)
	
	err := transaction(db, func (tx *sql.Tx) error {
		// Construct string with format "(?, ?, ..., ?)".
		prpStmtStr := fancyRepeat("(", "?", len(locs), ", ", ")")
		
		stmt := "SELECT location_name, lat, lng FROM coordinates WHERE location_name IN " + prpStmtStr
		log.Infof("Executing query '%s'", stmt)
		
		rows, err := tx.Query(stmt, locNames...)
		if err != nil {
			return err
		}
		
		return forEachRow(rows, func (rows *sql.Rows) error {
			var locName string
			var lat float32
			var lng float32
			err := rows.Scan(&locName, &lat, &lng)
			if err != nil {
				return err
			}
			
			locCoords[locName] = types.Coordinates{Lat: lat, Lng: lng}
			return nil
		})
	})
	
	if err == nil {
		log.Infof("Fetched %d coordinated locations in %d ms", len(locCoords), sw.TotalElapsedTimeMillis())
	}
	
	return locCoords, err
}
