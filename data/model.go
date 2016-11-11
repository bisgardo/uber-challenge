package data

import (
	"sort"
	"database/sql"
	"logging"
)

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
	
	logger.Infof("Creating table 'omdb' unless it already exists")
	_, err = tx.Exec(
		`CREATE TABLE IF NOT EXISTS omdb (
			movie_title VARCHAR(255) PRIMARY KEY,
			data        TEXT
		)`,
	)
	if err != nil {
		return err
	}
	
	return nil
}

func InsertMovies(tx *sql.Tx, ms []Movie, logger logging.Logger) error {
	// Insert movies, locations, and actors.
	as := make(map[string]int64)
	for _, m := range ms {
		// Inserting movie.
		logger.Debugf("Inserting movie '%v'", m.Title)
		mId, err := Insert(
			tx,
			"INSERT INTO movies VALUES (?, ?, ?, ?, ?, ?, ?)",
			nil, m.Title, m.Writer, m.Director, m.Distributor, m.ProductionCompany, m.ReleaseYear,
		)
		if err != nil {
			return err
		}
		//logger.Debugf("Movie inserted with ID %d", mId)
		
		// Inserting locations for movie.
		for _, l := range m.Locations {
			logger.Debugf("Inserting location '%v' for movie ID %d", l, mId)
			_, err := Insert(
				tx,
				"INSERT INTO locations VALUES (?, ?, ?, ?)",
				nil, mId, l.Name, l.FunFact,
			)
			if err != nil {
				return err
			}
		}
		
		// Inserting actors in movie.
		for _, a := range m.Actors {
			aId, exists := as[a]
			if !exists {
				// Inserting actor.
				logger.Debugf("Inserting actor '%v'", a)
				aId, err = Insert(tx, "INSERT INTO actors VALUES (?, ?)", nil, a)
				//logger.Debugf("Actor inserted with ID %d", aId)
				if err != nil {
					return err
				}
				as[a] = aId
			}
			
			// Inserting actor relation to movie.
			logger.Debugf("Inserting actor-movie relation %d-%d", mId, aId)
			_, err = Insert(tx, "INSERT INTO movies_actors VALUES (?, ?)", mId, aId)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (db *LocationDb) LoadMovies(filename string, logger logging.Logger) (*[]Movie, bool, error) {
	// Check if database is initialized and load from file if it isn't.
	initialized, err := db.Init(filename, logger)
	if err != nil {
		return nil, initialized, err
	}
	
	var ms []Movie
	
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
		
		// TODO Set capacity of `ms`.
		// TODO (Optionally) optimize by querying *all* locations and insert based on the map...
		
		for mId, m := range idMovieMap {
			if err := loadLocations(tx, mId, &m.Locations, logger); err != nil {
				return err
			}
			
			ms = append(ms, *m)
		}
		
		return nil
	})
	
	sort.Sort(ByTitle(ms))
	return &ms, initialized, err
}

func loadLocations(tx *sql.Tx, mId int64, ls *[]Location, logger logging.Logger) error {
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
		if err := InsertMovies(tx, ms, logger); err != nil {
			return err
		}
		return nil
	})
}
