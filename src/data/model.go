package data

import (
	"src/logging"
	"src/watch"
	"sort"
	"database/sql"
	"fmt"
	"strings"
)

type IdMoviePair struct {
	Id    int64
	Movie Movie
}

// TODO Turn bulk insertion stuff into nice utility (that is safe regarding injection and missing data) and replace current solutions with it...

// TODO Replace '*' in selects with explicit column names for robustness.

func InitTables(tx *sql.Tx, logger logging.Logger) error {
	// TODO Store writer and director in (renamed) actors table and add role to relation.
	
	var err error
	
	logger.Infof("Creating table 'movies' unless it already exists")
	_, err = tx.Exec(
		`CREATE TABLE IF NOT EXISTS movies (
			id                 INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			title              VARCHAR(255),
			writer             VARCHAR(255),
			director           VARCHAR(255),
			distributor        VARCHAR(255),
			production_company VARCHAR(255),
			release_year       INT UNSIGNED
		)`,
	)
	if err != nil {
		return err
	}
	
	logger.Infof("Creating table 'locations' unless it already exists")
	_, err = tx.Exec(
		`CREATE TABLE IF NOT EXISTS locations (
			id       INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			movie_id INT UNSIGNED,
			name     VARCHAR(255),
			fun_fact TEXT,
			
			FOREIGN KEY (movie_id) REFERENCES movies(id)
		)`,
	)
	if err != nil {
		return err
	}
	
	logger.Infof("Creating table 'actors' unless it already exists")
	_, err = tx.Exec(
		`CREATE TABLE IF NOT EXISTS actors (
			id   INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255)
		)`,
	)
	if err != nil {
		return err
	}
	
	logger.Infof("Creating table 'movies_actors' unless it already exists")
	_, err = tx.Exec(
		`CREATE TABLE IF NOT EXISTS movies_actors (
			movie_id INT UNSIGNED,
			actor_id INT UNSIGNED,
			
			PRIMARY KEY (movie_id, actor_id),
			FOREIGN KEY (movie_id) REFERENCES movies(id),
			FOREIGN KEY (actor_id) REFERENCES actors(id)
		)`,
	)
	if err != nil {
		return err
	}
	
	logger.Infof("Clearing table 'movies_actors'")
	_, err = tx.Exec("DELETE FROM movies_actors")
	if err != nil {
		return err
	}
	
	logger.Infof("Clearing table 'actors'")
	_, err = tx.Exec("DELETE FROM actors")
	if err != nil {
		return err
	}
	
	logger.Infof("Clearing table 'locations'")
	_, err = tx.Exec("DELETE FROM locations")
	if err != nil {
		return err
	}
	
	logger.Infof("Clearing table 'movies'")
	_, err = tx.Exec("DELETE FROM movies")
	if err != nil {
		return err
	}
	
	logger.Infof("Creating table 'coordinates' unless it already exists")
	_, err = tx.Exec(
		`CREATE TABLE IF NOT EXISTS coordinates (
			location_name VARCHAR(255) PRIMARY KEY,
			lat           FLOAT(10, 6) NOT NULL,
			lng           FLOAT(10, 6) NOT NULL
		)`,
	)
	if err != nil {
		return err
	}
	
	logger.Infof("Creating table 'movie_info' unless it already exists")
	_, err = tx.Exec(
		`CREATE TABLE IF NOT EXISTS movie_info (
			movie_title VARCHAR(255) PRIMARY KEY,
			info_json   TEXT
		)`,
	)
	if err != nil {
		return err
	}
	
	return nil
}

//func InsertMovies(tx *sql.Tx, ms []Movie, logger logging.Logger) error {
//	// Insert movies, locations, and actors.
//	as := make(map[string]int64)
//	for _, m := range ms {
//		// Inserting movie.
//		logger.Debugf("Inserting movie '%v'", m.Title)
//		mId, err := Insert(
//			tx,
//			"INSERT INTO movies VALUES (?, ?, ?, ?, ?, ?, ?)",
//			nil, m.Title, m.Writer, m.Director, m.Distributor, m.ProductionCompany, m.ReleaseYear,
//		)
//		if err != nil {
//			return err
//		}
//		//logger.Debugf("Movie inserted with ID %d", mId)
//		
//		// Inserting locations for movie.
//		for _, l := range m.Locations {
//			logger.Debugf("Inserting location '%v' for movie ID %d", l, mId)
//			_, err := Insert(
//				tx,
//				"INSERT INTO locations VALUES (?, ?, ?, ?)",
//				nil, mId, l.Name, l.FunFact,
//			)
//			if err != nil {
//				return err
//			}
//		}
//		
//		// Inserting actors in movie.
//		for _, a := range m.Actors {
//			aId, exists := as[a]
//			if !exists {
//				// Inserting actor.
//				logger.Debugf("Inserting actor '%v'", a)
//				aId, err = Insert(tx, "INSERT INTO actors VALUES (?, ?)", nil, a)
//				//logger.Debugf("Actor inserted with ID %d", aId)
//				if err != nil {
//					return err
//				}
//				as[a] = aId
//			}
//			
//			// Inserting actor relation to movie.
//			logger.Debugf("Inserting actor-movie relation %d-%d", mId, aId)
//			_, err = Insert(tx, "INSERT INTO movies_actors VALUES (?, ?)", mId, aId)
//			if err != nil {
//				return err
//			}
//		}
//	}
//	return nil
//}
//
//func InsertMoviesOptimized(tx *sql.Tx, ms []Movie, logger logging.Logger) error {
//	// Insert movies, locations, and actors.
//	as := make(map[string]int64)
//	relations := make(map[int64]int64)
//	insertLocationsVals := ""
//	insertMoviesActorsVals := ""
//	locationCount := 0
//	movieActorCount := 0
//	for _, m := range ms {
//		// Inserting movie.
//		logger.Debugf("Inserting movie '%v'", m.Title)
//		mId, err := Insert(
//			tx,
//			"INSERT INTO movies VALUES (?, ?, ?, ?, ?, ?, ?)",
//			nil, m.Title, m.Writer, m.Director, m.Distributor, m.ProductionCompany, m.ReleaseYear,
//		)
//		if err != nil {
//			return err
//		}
//		//logger.Debugf("Movie inserted with ID %d", mId)
//		
//		for _, l := range m.Locations {
//			insertLocationsVal := fmt.Sprintf(
//				"(NULL, %d, '%s', '%s')",
//				mId,
//				// TODO Make robust towards injection!
//				escapeSingleQuotes(l.Name),
//				escapeSingleQuotes(l.FunFact),
//			)
//			if len(insertLocationsVals) > 0 {
//				insertLocationsVals += ", "
//			}
//			insertLocationsVals += insertLocationsVal
//			
//			locationCount++
//		}
//		
//		// Inserting actors in movie.
//		for _, a := range m.Actors {
//			aId, exists := as[a]
//			if !exists {
//				// Inserting actor.
//				logger.Debugf("Inserting actor '%s'", a)
//				aId, err = Insert(tx, "INSERT INTO actors VALUES (?, ?)", nil, a)
//				//logger.Debugf("Actor inserted with ID %d", aId)
//				if err != nil {
//					return err
//				}
//				as[a] = aId
//			}
//			
//			relations[mId] = aId
//			
//			insertMoviesActorsVal := fmt.Sprintf("(%d, %d)", mId, aId)
//			if len(insertMoviesActorsVals) > 0 {
//				insertMoviesActorsVals += ", "
//			}
//			insertMoviesActorsVals += insertMoviesActorsVal
//			
//			movieActorCount++;
//		}
//	}
//	
//	// Bulk inserting locations.
//	insertLocationsStmt := "INSERT INTO locations VALUES " + insertLocationsVals
//	logger.Debugf("Inserting %d locations", locationCount)
//	logger.Debugf("Executing query %s", insertLocationsStmt)
//	if _, err := tx.Exec(insertLocationsStmt); err != nil {
//		return err
//	}
//	
//	// Bulk inserting movie-actor relations.
//	insertMoviesActorsStmt := "INSERT INTO movies_actors VALUES " + insertMoviesActorsVals
//	logger.Debugf("Inserting %d actor-movie relations", movieActorCount)
//	logger.Debugf("Executing query %s", insertMoviesActorsStmt)
//	if _, err := tx.Exec(insertMoviesActorsStmt); err != nil {
//		return err
//	}
//	
//	return nil
//}

func escapeSingleQuotes(s string) string {
	return strings.Replace(s, "'", "\\'", -1)
}

func InsertMoviesMoreOptimized(tx *sql.Tx, ms []Movie, logger logging.Logger) error {
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
	
	insertMoviesStmt := "INSERT INTO movies VALUES" + insertMoviesVals
	//logger.Debugf("Executing query %s", insertMoviesStmt)
	if _, err := tx.Exec(insertMoviesStmt); err != nil {
		return err
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
	
	// TODO Make robust towards there being no data.
	insertLocationsStmt := "INSERT INTO locations VALUES" + insertLocationsVals
	//logger.Debugf("Executing query %s", insertLocationsStmt)
	if _, err := tx.Exec(insertLocationsStmt); err != nil {
		return err
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
	
	// TODO Make robust towards there being no data.
	insertActorsStmt := "INSERT INTO actors VALUES" + insertActorsVals
	//logger.Debugf("Executing query %s", insertActorsStmt)
	if _, err := tx.Exec(insertActorsStmt); err != nil {
		return err
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
	
	// TODO Make robust towards there being no data.
	insertMovieActorStmt := "INSERT INTO movies_actors VALUES" + insertMovieActorVals
	//logger.Debugf("Executing query %s", insertMovieActorStmt)
	if _, err := tx.Exec(insertMovieActorStmt); err != nil {
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
	err = ForEachRow(rows, func (rows *sql.Rows) error {
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
	err = ForEachRow(rows, func (rows *sql.Rows) error {
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

func LoadMovies(db *LocationDb, filename string, logger logging.Logger, optimize bool) ([]IdMoviePair, bool, error) {
	// Check if database is initialized and load from file if it isn't.
	initialized, err := db.Init(filename, logger)
	if err != nil {
		return nil, initialized, err
	}
	
	var ms []IdMoviePair
	
	err = db.transaction(func (tx *sql.Tx) error {
		logger.Debugf("Querying movies")
		rows, err := tx.Query("SELECT * FROM movies")
		if err != nil {
			return err
		}
		
		// TODO Extract function that returns/fills the map.
		idMovieMap := make(map[int64]*Movie)
		err = ForEachRow(rows, func (rows *sql.Rows) error {
			var mId int64
			var m Movie
			
			if err := rows.Scan(&mId, &m.Title, &m.Writer, &m.Director, &m.Distributor, &m.ProductionCompany, &m.ReleaseYear); err != nil {
				return err
			}
			
			idMovieMap[mId] = &m
			return nil
		})
		if err != nil {
			return err
		}
		
		ms = make([]IdMoviePair, 0, len(idMovieMap))
		if optimize {
			if err := LoadAllLocations(tx, idMovieMap, logger); err != nil {
				return err
			}
			for mId, m := range idMovieMap {
				ms = append(ms, IdMoviePair{mId, *m})
			}
		} else {
			for mId, m := range idMovieMap {
				if err := LoadLocations(tx, mId, &m.Locations, logger); err != nil {
					return err
				}
				
				ms = append(ms, IdMoviePair{mId, *m})
			}
		}
		
		return nil
	})
	
	sort.Sort(ByTitle(ms))
	return ms, initialized, err
}

func LoadLocations(tx *sql.Tx, mId int64, ls *[]Location, logger logging.Logger) error {
	logger.Debugf("Querying locations for movie %d", mId)
	
	rows, err := tx.Query("SELECT * FROM locations AS l WHERE l.movie_id = ?", mId)
	if err != nil {
		return err
	}
	
	return ForEachRow(rows, func (rows *sql.Rows) error {
		var dummy int64
		var l Location
		err := rows.Scan(&dummy, &dummy, &l.Name, &l.FunFact)
		if err != nil {
			return err
		}
		
		*ls = append(*ls, l)
		return nil
	})
}

func LoadAllLocations(tx *sql.Tx, idMovieMap map[int64]*Movie, logger logging.Logger) error {
	logger.Debugf("Querying all locations")
	
	rows, err := tx.Query("SELECT * FROM locations")
	if err != nil {
		return err
	}
	
	return ForEachRow(rows, func (rows *sql.Rows) error {
		var lId int64
		var mId int64
		var l Location
		err := rows.Scan(&lId, &mId, &l.Name, &l.FunFact)
		if err != nil {
			return err
		}
		
		m := idMovieMap[mId]
		m.Locations = append(m.Locations, l)
		return nil
	})
}

//func LoadPositionNames(tx *sql.Tx) ([]string, error) {
//	rows, err := tx.Query("SELECT location_name FROM coordinates")
//	if err != nil {
//		return nil, err
//	}
//	
//	var names []string
//	err = ForEachRow(rows, func (rows *sql.Rows) error {
//		var name string
//		rows.Scan(&name)
//		names = append(names, name)
//		return nil
//	})
//	
//	return names, err
//}

func StoreMovies(db *LocationDb, ms []Movie, logger logging.Logger) error {
	return db.transaction(func (tx *sql.Tx) error {
		if err := InitTables(tx, logger); err != nil {
			return err
		}
		if err := InsertMoviesMoreOptimized(tx, ms, logger); err != nil {
			return err
		}
		return nil
	})
}

func LoadMovieInfoJson(db *LocationDb, title string, logger logging.Logger) (string, error) {
	sw := watch.NewStopWatch()
	
	var info string
	err := db.transaction(func (tx *sql.Tx) error {
		row := tx.QueryRow("SELECT info_json FROM movie_info WHERE movie_title = ?", title)
		return row.Scan(&info)
	})
	
	if err != nil {
		logger.Infof("Loaded info for movie '%s' in %d ms", title, sw.TotalElapsedTimeMillis())
	}
	
	return info, err
}

func LoadMovieInfoJsons(db *LocationDb, logger logging.Logger) (map[string]string, error) {
	sw := watch.NewStopWatch()
	movieInfo := make(map[string]string)
	err := db.transaction(func (tx *sql.Tx) error {
		rows, err := tx.Query("SELECT * FROM movie_info")
		if err != nil {
			return err
		}
		
		return ForEachRow(rows, func (rows *sql.Rows) error {
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


func StoreMovieInfo(db *LocationDb, movieInfo map[string]string, logger logging.Logger) error {
	if len(movieInfo) == 0 {
		return nil
	}
	
	logger.Infof("Inserting %d movie infos into database", len(movieInfo))
	
	return db.transaction(func (tx *sql.Tx) error {
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
		
		// TODO Make robust towards there being no data.
		insertStmt := "INSERT INTO movie_info VALUES" + vals
		logger.Debugf("Executing query %s", insertStmt)
		_, err := tx.Exec(insertStmt)
		return err
	})
}

func LoadMovie(db *LocationDb, id int) (Movie, error) {
	var m Movie
	err := db.transaction(func (tx *sql.Tx) error {
		row := tx.QueryRow("SELECT * FROM movies WHERE id = ?", id)
		
		var dummy int
		return row.Scan(
			&dummy,
			&m.Title,
			&m.Writer,
			&m.Director,
			&m.Distributor,
			&m.ProductionCompany,
			&m.ReleaseYear,
		)
	})
	return m, err
}
