package sqldb

import (
	"src/data/types"
	"src/logging"
	"src/watch"
	"sort"
	"database/sql"
)

func LoadMovie(db *sql.DB, id int64, logger logging.Logger) (types.Movie, error) {
	var m types.Movie
	err := transaction(db, func (tx *sql.Tx) error {
		row := tx.QueryRow("SELECT * FROM movies WHERE id = ?", id)
		
		var dummy int
		err := row.Scan(
			&dummy,
			&m.Title,
			&m.Writer,
			&m.Director,
			&m.Distributor,
			&m.ProductionCompany,
			&m.ReleaseYear,
		)
		if err != nil {
			return err
		}
		
		if err := LoadLocations(tx, id, &m.Locations, logger); err != nil {
			return err
		}
		
		if err := LoadActors(tx, id, &m.Actors, logger); err != nil {
			return err
		}
		
		return nil
	})
	return m, err
}

func LoadLocations(tx *sql.Tx, mId int64, ls *[]types.Location, logger logging.Logger) error {
	logger.Debugf("Querying locations for movie %d", mId)
	
	rows, err := tx.Query("SELECT name, fun_fact FROM locations AS l WHERE l.movie_id = ?", mId)
	if err != nil {
		return err
	}
	
	return forEachRow(rows, func (rows *sql.Rows) error {
		var l types.Location
		err := rows.Scan(&l.Name, &l.FunFact)
		if err != nil {
			return err
		}
		
		*ls = append(*ls, l)
		return nil
	})
}

func LoadActors(tx *sql.Tx, mId int64, as *[]string, logger logging.Logger) error {
	logger.Debugf("Querying actors for movie %d", mId)
	
	rows, err := tx.Query(
		"SELECT a.name FROM actors AS a, movies_actors AS r WHERE a.id = r.actor_id AND r.movie_id = ?",
		mId,
	)
	if err != nil {
		return err
	}
	
	return forEachRow(rows, func (rows *sql.Rows) error {
		var a string
		err := rows.Scan(&a)
		if err != nil {
			return err
		}
		
		*as = append(*as, a)
		return nil
	})
}

func LoadMovies(db *sql.DB, logger logging.Logger) ([]types.IdMoviePair, error) {
	var ms []types.IdMoviePair
	
	err := transaction(db, func (tx *sql.Tx) error {
		logger.Debugf("Querying movies")
		
		// Loading all movies.
		rows, err := tx.Query("SELECT id, title, writer, director, distributor, production_company, release_year FROM movies")
		if err != nil {
			return err
		}
		
		idMovieMap := make(map[int64]*types.Movie)
		err = forEachRow(rows, func (rows *sql.Rows) error {
			var mId int64
			var m types.Movie
			
			if err := rows.Scan(&mId, &m.Title, &m.Writer, &m.Director, &m.Distributor, &m.ProductionCompany, &m.ReleaseYear); err != nil {
				return err
			}
			
			idMovieMap[mId] = &m
			return nil
		})
		if err != nil {
			return err
		}
		
		// Load all locations.
		if err := LoadAllLocations(tx, idMovieMap, logger); err != nil {
			return err
		}
		
		// Load all actors.
		if err := LoadAllActors(tx, idMovieMap, logger); err != nil {
			return err
		}
		
		ms = make([]types.IdMoviePair, 0, len(idMovieMap))
		for mId, m := range idMovieMap {
			ms = append(ms, types.IdMoviePair{Id: mId, Movie: *m})
		}
		
		// TODO Load actors...
		
		return nil
	})
	if err != nil {
		return nil, err
	}
	
	sort.Sort(types.ByTitle(ms))
	return ms, nil
}

func LoadAllLocations(tx *sql.Tx, idMovieMap map[int64]*types.Movie, logger logging.Logger) error {
	logger.Debugf("Querying all locations")
	
	rows, err := tx.Query("SELECT movie_id, name, fun_fact FROM locations")
	if err != nil {
		return err
	}
	
	return forEachRow(rows, func (rows *sql.Rows) error {
		var mId int64
		var l types.Location
		err := rows.Scan(&mId, &l.Name, &l.FunFact)
		if err != nil {
			return err
		}
		
		m, exists := idMovieMap[mId]
		if !exists {
			panic("Unexpected location ID...")
		}
		m.Locations = append(m.Locations, l)
		return nil
	})
}

func LoadAllActors(tx *sql.Tx, idMovieMap map[int64]*types.Movie, logger logging.Logger) error {
	logger.Debugf("Querying all actors")
	
	var rows *sql.Rows
	var err error
	rows, err = tx.Query("SELECT id, name FROM actors")
	if err != nil {
		return err
	}
	
	idActorMap := make(map[int64]string)
	err = forEachRow(rows, func (rows *sql.Rows) error {
		var id int64
		var a string
		if err := rows.Scan(&id, &a); err != nil {
			return err
		}
		
		idActorMap[id] = a
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
		var mId int64
		var aId int64
		if err := rows.Scan(&mId, &aId); err != nil {
			return err
		}
		
		m, exists := idMovieMap[mId]
		if !exists {
			panic("Unexpected movie ID...")
		}
		
		a, exists := idActorMap[aId]
		if !exists {
			panic("Unexpected actor ID...")
		}
		
		m.Actors = append(m.Actors, a)
		return nil
	})
	if err != nil {
		return err
	}
	
	return nil
}

func LoadMovieInfoJson(db *sql.DB, title string, logger logging.Logger) (string, error) {
	sw := watch.NewStopWatch()
	
	var info string
	err := transaction(db, func (tx *sql.Tx) error {
		row := tx.QueryRow("SELECT info_json FROM movie_info WHERE movie_title = ?", title)
		return row.Scan(&info)
	})
	
	if err != nil {
		logger.Infof("Loaded info for movie '%s' in %d ms", title, sw.TotalElapsedTimeMillis())
	}
	
	return info, err
}

func LoadMovieInfoJsons(db *sql.DB, logger logging.Logger) (map[string]string, error) {
	// TODO Parallelize (if the API allows it) and consider using memcached (with expiration) instead of SQL.
	
	sw := watch.NewStopWatch()
	movieInfo := make(map[string]string)
	err := transaction(db, func (tx *sql.Tx) error {
		rows, err := tx.Query("SELECT movie_title, info_json FROM movie_info")
		if err != nil {
			return err
		}
		
		return forEachRow(rows, func (rows *sql.Rows) error {
			var t string
			var i string
			if err := rows.Scan(&t, &i); err != nil {
				return err
			}
			movieInfo[t] = i
			return nil
		})
	})
	
	if err == nil {
		logger.Infof("Fetched info for %d movies in %d ms", len(movieInfo), sw.TotalElapsedTimeMillis())
	}
	
	return movieInfo, err
}

func LoadCoordinates(db *sql.DB, locations []types.Location, logger logging.Logger) (map[string]types.Coordinates, error) {
	sw := watch.NewStopWatch()
	
	names := make([]interface{}, 0, len(locations))
	for _, l := range locations {
		name := l.Name
		names = append(names, name)
	}
	
	res := make(map[string]types.Coordinates)
	
	err := transaction(db, func (tx *sql.Tx) error {
		// Construct string with format "(?, ?, ..., ?)".
		prpStmtStr := fancyRepeat("(", "?", len(locations), ", ", ")")
		
		stmt := "SELECT location_name, lat, lng FROM coordinates WHERE location_name IN " + prpStmtStr
		logger.Infof("Executing query '%s'", stmt)
		
		rows, err := tx.Query(stmt, names...)
		if err != nil {
			return err
		}
		
		return forEachRow(rows, func (rows *sql.Rows) error {
			var n string
			var lat float32
			var lng float32
			err := rows.Scan(&n, &lat, &lng)
			if err != nil {
				return err
			}
			
			res[n] = types.Coordinates{Lat: lat, Lng: lng}
			return nil
		})
	})
	
	if err == nil {
		logger.Infof("Fetched %d coordinated locations in %d ms", len(res), sw.TotalElapsedTimeMillis())
	}
	
	return res, err
}
