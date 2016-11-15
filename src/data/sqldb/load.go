package sqldb

import (
	"src/data/types"
	"src/logging"
	"src/watch"
	"sort"
	"database/sql"
)

// TODO Replace '*' in selects with explicit column names for robustness.

func LoadMovies(db *sql.DB, logger logging.Logger) ([]types.IdMoviePair, error) {
	var ms []types.IdMoviePair
	
	err := transaction(db, func (tx *sql.Tx) error {
		logger.Debugf("Querying movies")
		rows, err := tx.Query("SELECT * FROM movies")
		if err != nil {
			return err
		}
		
		// TODO Extract function that returns/fills the map.
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
		
		ms = make([]types.IdMoviePair, 0, len(idMovieMap))
		if err := LoadAllLocations(tx, idMovieMap, logger); err != nil {
			return err
		}
		for mId, m := range idMovieMap {
			ms = append(ms, types.IdMoviePair{mId, *m})
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

func LoadLocations(tx *sql.Tx, mId int64, ls *[]types.Location, logger logging.Logger) error {
	logger.Debugf("Querying locations for movie %d", mId)
	
	rows, err := tx.Query("SELECT * FROM locations AS l WHERE l.movie_id = ?", mId)
	if err != nil {
		return err
	}
	
	return forEachRow(rows, func (rows *sql.Rows) error {
		var dummy int64
		var l types.Location
		err := rows.Scan(&dummy, &dummy, &l.Name, &l.FunFact)
		if err != nil {
			return err
		}
		
		*ls = append(*ls, l)
		return nil
	})
}

func LoadAllLocations(tx *sql.Tx, idMovieMap map[int64]*types.Movie, logger logging.Logger) error {
	logger.Debugf("Querying all locations")
	
	rows, err := tx.Query("SELECT * FROM locations")
	if err != nil {
		return err
	}
	
	return forEachRow(rows, func (rows *sql.Rows) error {
		var lId int64
		var mId int64
		var l types.Location
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
	sw := watch.NewStopWatch()
	movieInfo := make(map[string]string)
	err := transaction(db, func (tx *sql.Tx) error {
		rows, err := tx.Query("SELECT * FROM movie_info")
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
		
		// TODO Load actors too?
		return LoadLocations(tx, id, &m.Locations, logger)
	})
	return m, err
}
