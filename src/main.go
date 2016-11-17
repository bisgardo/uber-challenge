package app

import (
	"src/data"
	"src/data/types"
	"src/data/sqldb"
	"src/data/fetch"
	"src/config"
	"src/tpl"
	"src/logging"
	"src/watch"
	"appengine"
	"net/http"
	"database/sql"
	"encoding/json"
	"strings"
	"strconv"
	"fmt"
	"errors"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

var recordedLog []string
var recordedError error

var jsonFileName = config.JsonFileName()
var mapsApiKey = config.MapsApiKey()

func init() {
	log := logging.NewRecordingLogger(&logging.InitLogger{}, true)
	
	log.Infof("Spinning up instance with ID '%s'", appengine.InstanceID())
	
	err := openInit(log)
	recordInitUpdate(err, log)
	if err != nil {
		panic(err)
	}
	
	http.HandleFunc("/", render(front))
	http.HandleFunc("/movie", render(movies))
	http.HandleFunc("/movie/", render(movie))
	http.HandleFunc("/status", renderStatus)
	http.HandleFunc("/update", renderUpdate)
	http.HandleFunc("/ping", renderPing)
	http.HandleFunc("/data", renderDataJson)
	
	// TODO Make "raw data dump" page.
	// TODO Add pages for actor, ...
}

func openInit(log *logging.RecordingLogger) error {
	if err := openDb(log); err != nil {
		return err
	}
	if _, err := data.Init(db, jsonFileName, log); err != nil {
		return err
	}
	return nil
}

func recordInitUpdate(err error, log *logging.RecordingLogger) {
	recordedError = err
	entries := log.Entries
	entriesCopy := make([]string, len(entries))
	copy(entriesCopy, entries)
	recordedLog = entriesCopy
}

func openDb(logger logging.Logger) error {
	var err error
	if appengine.IsDevAppServer() {
		logger.Infof("Running in development mode")
		db, err = sql.Open("mysql", config.LocalDbSourceName())
	} else {
		logger.Infof("Running in production mode")
		db, err = sql.Open("mysql", config.CloudDbSourceName())
	}
	return err
}

func render(renderer func(w http.ResponseWriter, r *http.Request, log *logging.RecordingLogger) error) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := appengine.NewContext(r)
		log := logging.NewRecordingLogger(ctx, false)
		
		// Check if database is initialized and load from file if it isn't.
		initialized, err := data.Init(db, jsonFileName, log)
		if initialized {
			recordInitUpdate(err, log)
		}
		
		if err == nil {
			err = renderer(w, r, log)
		}
		
		if err != nil {
			ctx.Errorf("ERROR: %+v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func front(w http.ResponseWriter, r *http.Request, log *logging.RecordingLogger) error {
	ctx := appengine.NewContext(r)
	templateData := tpl.NewTemplateData(ctx, log, nil)
	return tpl.Render(w, tpl.About, templateData)
}

func movie(w http.ResponseWriter, r *http.Request, log *logging.RecordingLogger) error {
	path := r.URL.Path
	idx := strings.LastIndex(path, "/")
	idStr := path[idx + 1:]
	
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return errors.New(fmt.Sprintf("Invalid movie ID '%s'", idStr))
	}
	
	log.Infof("Rendering movie with ID %d", id)
	
	movie, err := sqldb.LoadMovie(db, int64(id), log)
	if err != nil {
		http.Error(w, fmt.Sprintf("Movie with ID %d not found", id), http.StatusNotFound)
		return nil
	}
	
	log.Infof("Loading coordinates")
	locNameCoordsMap, err := sqldb.LoadCoordinates(db, movie.Locations, log)
	
	missingCoords := make(map[string]*types.Coordinates)
	for _, loc := range movie.Locations {
		locName := loc.Name
		if _, exists := locNameCoordsMap[locName]; !exists {
			missingCoords[locName] = nil
		}
	}
	
	// Load missing coordinates.
	ctx := appengine.NewContext(r)
	delayFunc := func (count int) int { return 50 * count }
	fetch.FetchMissingLocationNames(missingCoords, delayFunc, mapsApiKey, ctx, log)
	
	// Store missing coordinates.
	if err := sqldb.StoreCoordinates(db, missingCoords, log); err != nil {
		return err
	}
	
	// Add missing coordinates to `coords`.
	for locName, locCoords := range missingCoords {
		if locCoords != nil {
			locNameCoordsMap[locName] = *locCoords
		}
	}
	
	// Set coordinates on locations.
	for i := range movie.Locations {
		loc := &movie.Locations[i]
		loc.Coordinates = locNameCoordsMap[loc.Name]
	}
	
	type MovieInfo struct {
		Title      string
		Year       string
		Rated      string
		Released   string
		Runtime    string
		Genre      string
		Director   string
		Writer     string
		Actors     string
		Plot       string
		Language   string
		Country    string
		Awards     string
		Poster     string
		Metascore  string
		ImdbRating string
		ImdbVotes  string
		ImdbID     string
	}
	
	var info MovieInfo
	
	info.Title = movie.Title
	info.Actors = strings.Join(movie.Actors, ", ")
	info.Writer = movie.Writer
	info.Director = movie.Director
	info.Released = strconv.Itoa(movie.ReleaseYear)
	
	args := &struct {
		Movie    *types.Movie
		Info     *MovieInfo
	}{&movie, &info}
	
	if infoJson, err := sqldb.LoadMovieInfoJson(db, movie.Title, log); infoJson != "" && err == nil {
		// Only attempt to parse JSON if it was loaded successfully
		if err := json.Unmarshal([]byte(infoJson), &info); err != nil {
			log.Errorf(err.Error())
		}
	}
	
	templateData := tpl.NewTemplateData(ctx, log, args)
	templateData.Subtitle = info.Title
	return tpl.Render(w, tpl.Movie, templateData)
}

func movies(w http.ResponseWriter, r *http.Request, log *logging.RecordingLogger) error {
	log.Infof("Rendering movie list page")
	
	movies, err := sqldb.LoadMovies(db, log)
	if err != nil {
		return err
	}
	
	ctx := appengine.NewContext(r)
	templateData := tpl.NewTemplateData(ctx, log, movies)
	templateData.Subtitle = "List"
	if err := tpl.Render(w, tpl.Movies, templateData); err != nil {
		return err
	}
	
	return nil
}

// TODO Have one optimized endpoint with only data needed for autocomplete and one with *all* data.

func renderDataJson(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	
	movies, err := sqldb.LoadMovies(db, ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	
	if err := json.NewEncoder(w).Encode(movies); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func renderUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	log := logging.NewRecordingLogger(ctx, false)
	
	var err error
	defer recordInitUpdate(err, log)
	err = update(w, r, log)
	if err != nil {
		ctx.Errorf("ERROR: %+v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func update(w http.ResponseWriter, r *http.Request, log *logging.RecordingLogger) error {
	var err error
	
	ctx := appengine.NewContext(r)
	
	if r.Method != "POST" {
		errMsg := "Cannot " + r.Method + " '/update'"
		ctx.Errorf(errMsg)
		http.Error(w, errMsg, http.StatusMethodNotAllowed)
		return nil
	}
	
	// TODO Add timestamp(s) to DB for locking to work across instances.
	
	data.InitUpdateMutex.Lock()
	defer data.InitUpdateMutex.Unlock()
	
	movies, err := fetch.FetchFromUrl(config.ServiceUrl(), ctx, log)
	if err != nil {
		return err
	}
	if err := sqldb.InitTablesAndStoreMovies(db, movies, log); err != nil {
		return err
	}
	
	// Fetch movie data.
	// TODO This information should be fetched on demand (as location data is) or also fetched on initialization.
	movieTitleInfoMap, err := sqldb.LoadMovieInfoJsons(db, log)
	if err != nil {
		return err
	}
	
	movieTitleInfo := make(map[string]string)
	for _, movie := range movies {
		movieTitle := movie.Title
		if _, exists := movieTitleInfoMap[movieTitle]; exists {
			// Info already in DB.
			continue
		}
		
		infoJson, err := fetch.FetchMovieInfo(movieTitle, ctx, log)
		if err != nil {
			return err
		}
		
		info := &struct {
			Response string
		}{}
		
		json.Unmarshal([]byte(infoJson), info)
		if info.Response != "True" {
			infoJson = "";
		}
		
		movieTitleInfo[movieTitle] = infoJson
	}
	
	// Store movie data.
	if err := sqldb.StoreMovieInfo(db, movieTitleInfo, log); err != nil {
		return err
	}
	
	http.Redirect(w, r, "", http.StatusFound)
	return nil
}

func renderStatus(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	log := logging.NewRecordingLogger(ctx, false)
	if err := status(w, r, log); err != nil {
		ctx.Errorf(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func status(w http.ResponseWriter, r *http.Request, logger *logging.RecordingLogger) error {
	logger.Infof("Rendering status page")
	
	sw := watch.NewStopWatch()
	
	mc := 0
	ac := 0
	lc := 0
	rc := 0
	cc := 0
	ic := 0
	
	mt := int64(0)
	at := int64(0)
	lt := int64(0)
	rt := int64(0)
	ct := int64(0)
	it := int64(0)
	
	initialized, err := data.IsInitialized(db)
	if err != nil {
		return err
	}
	
	if initialized {
		querySingleInt := func(db *sql.DB, sql string, args ...interface{}) (int, error) {
			row := db.QueryRow(sql, args...)
			var i int
			err := row.Scan(&i)
			return i, err
		}
		
		mc, err = querySingleInt(db, "SELECT COUNT(*) FROM movies")
		if err != nil {
			return err
		}
		mt = sw.ElapsedTimeMillis(true)
		
		ac, err = querySingleInt(db, "SELECT COUNT(*) FROM actors")
		if err != nil {
			return err
		}
		at = sw.ElapsedTimeMillis(true)
		
		lc, err = querySingleInt(db, "SELECT COUNT(*) FROM locations")
		if err != nil {
			return err
		}
		lt = sw.ElapsedTimeMillis(true)
		
		rc, err = querySingleInt(db, "SELECT COUNT(*) FROM movies_actors")
		if err != nil {
			return err
		}
		rt = sw.ElapsedTimeMillis(true)
		
		cc, err = querySingleInt(db, "SELECT COUNT(*) FROM coordinates")
		if err != nil {
			return err
		}
		ct = sw.ElapsedTimeMillis(true)
		
		ic, err = querySingleInt(db, "SELECT COUNT(*) FROM movie_info")
		if err != nil {
			return err
		}
		it = sw.ElapsedTimeMillis(true)
	}
	
	dt := sw.TotalElapsedTimeMillis()
	
	args := struct {
		Clock            string
		Time             int64
		MoviesCount      int
		MoviesTime       int64
		ActorsCount      int
		ActorsTime       int64
		LocationsCount   int
		LocationsTime    int64
		MovieActorsCount int
		MovieActorsTime  int64
		CoordinatesCount int
		CoordinatesTime  int64
		InfoCount        int
		InfoTime         int64
		RecordedErr      error
		RecordedLog      []string
	}{sw.InitTime.String(), dt, mc, mt, ac, at, lc, lt, rc, rt, cc, ct, ic, it, recordedError, recordedLog}
	
	ctx := appengine.NewContext(r)
	templateData := tpl.NewTemplateData(ctx, logger, args)
	templateData.Subtitle = "Status"
	return tpl.Render(w, tpl.Status, templateData)
}

func renderPing(w http.ResponseWriter, r *http.Request) {
	if err := ping(w, r); err != nil {
		ctx := appengine.NewContext(r)
		ctx.Errorf(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func ping(w http.ResponseWriter, r *http.Request) error {
	sw := watch.NewStopWatch()
	//err := db.Ping()
	row := db.QueryRow("SELECT 42")
	
	var _42 int
	if err := row.Scan(&_42); err != nil {
		return err
	}
	
	if _42 != 42 {
		return errors.New("Invalid response from DB")
	}
	
	args := &struct {
		Clock   string
		Time    int64
	}{sw.InitTime.String(), sw.TotalElapsedTimeMillis()}
	
	ctx := appengine.NewContext(r)
	log := logging.NewRecordingLogger(ctx, false)
	templateData := tpl.NewTemplateData(ctx, log, args)
	templateData.Subtitle = "Ping"
	if err := tpl.Render(w, tpl.Ping, templateData); err != nil {
		return err
	}
	return nil
}
