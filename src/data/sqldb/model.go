package sqldb

import (
	"src/logging"
	"database/sql"
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
