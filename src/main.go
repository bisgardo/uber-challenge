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
	logger := logging.NewRecordingLogger(&logging.InitLogger{}, true)
	
	logger.Infof("Spinning up instance with ID '%s'", appengine.InstanceID())
	
	err := openInit(logger)
	recordInitUpdate(err, logger)
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

func openInit(logger *logging.RecordingLogger) error {
	if err := Open(logger); err != nil {
		return err
	}
	if _, err := data.Init(db, jsonFileName, logger); err != nil {
		return err
	}
	return nil
}

func recordInitUpdate(err error, recordingLogger *logging.RecordingLogger) {
	recordedError = err
	entries := recordingLogger.Entries
	entriesCopy := make([]string, len(entries))
	copy(entriesCopy, entries)
	recordedLog = entriesCopy
}

func Open(logger logging.Logger) error {
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

func render(renderer func(w http.ResponseWriter, r *http.Request, logger *logging.RecordingLogger) error) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := appengine.NewContext(r)
		logger := logging.NewRecordingLogger(ctx, false)
		
		// Check if database is initialized and load from file if it isn't.
		initialized, err := data.Init(db, jsonFileName, logger)
		if initialized {
			recordInitUpdate(err, logger)
		}
		
		if err == nil {
			err = renderer(w, r, logger)
		}
		
		if err != nil {
			ctx.Errorf("ERROR: %+v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func front(w http.ResponseWriter, r *http.Request, logger *logging.RecordingLogger) error {
	ctx := appengine.NewContext(r)
	version := appengine.VersionID(ctx)
	args := &struct {
		Logs    []string
		Version string
	}{logger.Entries, version}
	
	return tpl.Render(w, tpl.About, args)
}

func movie(w http.ResponseWriter, r *http.Request, logger *logging.RecordingLogger) error {
	path := r.URL.Path
	i := strings.LastIndex(path, "/")
	id := path[i + 1:]
	
	mId, err := strconv.Atoi(id)
	if err != nil {
		return errors.New(fmt.Sprintf("Invalid movie ID '%s'", id))
	}
	
	logger.Infof("Rendering movie with ID %d", mId)
	
	m, err := sqldb.LoadMovie(db, int64(mId), logger)
	if err != nil {
		http.Error(w, fmt.Sprintf("Movie with ID %d not found", mId), http.StatusNotFound)
		return nil
	}
	
	logger.Infof("Loading coordinates")
	coords, err := sqldb.LoadCoordinates(db, m.Locations, logger)
	
	missingCoords := make(map[string]*types.Coordinates)
	for _, location := range m.Locations {
		name := location.Name
		if _, exists := coords[name]; !exists {
			missingCoords[name] = nil
		}
	}
	
	// Load missing coordinates.
	ctx := appengine.NewContext(r)
	delayFunc := func (count int) int { return 50 * count }
	fetch.FetchMissingLocationNames(missingCoords, delayFunc, mapsApiKey, ctx, logger)
	
	// Store missing coordinates.
	if err := sqldb.StoreCoordinates(db, missingCoords, logger); err != nil {
		return err
	}
	
	// Add missing coordinates to `coords`.
	for n, c := range missingCoords {
		if c != nil {
			coords[n] = *c
		}
	}
	
	// Set coordinates on locations.
	for i := range m.Locations {
		location := &m.Locations[i]
		location.Coordinates = coords[location.Name]
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
	
	info.Title = m.Title
	info.Actors = strings.Join(m.Actors, ", ")
	info.Writer = m.Writer
	info.Director = m.Director
	info.Released = strconv.Itoa(m.ReleaseYear)
	
	version := appengine.VersionID(ctx)
	args := &struct {
		Logs    []string
		Subtitle string
		Movie    *types.Movie
		Info     *MovieInfo
		Version  string
	}{logger.Entries, info.Title, &m, &info, version}
	
	if infoJson, err := sqldb.LoadMovieInfoJson(db, m.Title, logger); infoJson != "" && err == nil {
		// Only attempt to parse JSON if it was loaded successfully
		if err := json.Unmarshal([]byte(infoJson), &info); err != nil {
			logger.Errorf(err.Error())
		}
	}
	
	return tpl.Render(w, tpl.Movie, args)
}

func movies(w http.ResponseWriter, r *http.Request, logger *logging.RecordingLogger) error {
	logger.Infof("Rendering movie list page")
	
	ms, err := sqldb.LoadMovies(db, logger)
	if err != nil {
		return err
	}
	
	ctx := appengine.NewContext(r)
	version := appengine.VersionID(ctx)
	
	args := &struct {
		Logs       []string
		Movies     []types.IdMoviePair
		OutputLogs bool
		Version    string
	}{logger.Entries, ms, true, version}
	
	if err := tpl.Render(w, tpl.Movies, args); err != nil {
		return err
	}
	
	return nil
}

// TODO Have one endpoint with only data needed for autocomplete and one with *all* data.

func renderDataJson(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	
	ms, err := sqldb.LoadMovies(db, ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	
	if err := json.NewEncoder(w).Encode(ms); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func renderUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	logger := logging.NewRecordingLogger(ctx, false)
	
	var err error
	defer recordInitUpdate(err, logger)
	err = update(w, r, logger)
	if err != nil {
		ctx.Errorf("ERROR: %+v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func update(w http.ResponseWriter, r *http.Request, logger *logging.RecordingLogger) error {
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
	
	ms, err := fetch.FetchFromUrl(config.ServiceUrl(), ctx, logger)
	if err != nil {
		return err
	}
	if err := sqldb.InitTablesAndStoreMovies(db, ms, logger); err != nil {
		return err
	}
	
	// Fetch movie data.
	// TODO This information should be fetched on demand (as location data is) or also fetched on initialization.
	mi, err := sqldb.LoadMovieInfoJsons(db, logger)
	if err != nil {
		return err
	}
	
	movieTitleInfo := make(map[string]string)
	for _, m := range ms {
		t := m.Title
		if _, exists := mi[t]; exists {
			// Info already in DB.
			continue
		}
		
		infoJson, err := fetch.FetchMovieInfo(t, ctx, logger)
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
		
		movieTitleInfo[t] = infoJson
	}
	
	// Store movie data.
	if err := sqldb.StoreMovieInfo(db, movieTitleInfo, logger); err != nil {
		return err
	}
	
	http.Redirect(w, r, "", http.StatusFound)
	return nil
}

func renderStatus(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	logger := logging.NewRecordingLogger(ctx, false)
	if err := status(w, r, logger); err != nil {
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
	
	ctx := appengine.NewContext(r)
	version := appengine.VersionID(ctx)
	
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
		Logs             []string
		Version          string
	}{sw.InitTime.String(), dt, mc, mt, ac, at, lc, lt, rc, rt, cc, ct, ic, it, recordedError, recordedLog, logger.Entries, version}
	
	return tpl.Render(w, tpl.Status, args)
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
	
	ctx := appengine.NewContext(r)
	version := appengine.VersionID(ctx)
	args := &struct {
		Logs    []string
		Clock   string
		Time    int64
		Version string
	}{nil, sw.InitTime.String(), sw.TotalElapsedTimeMillis(), version}
	
	if err := tpl.Render(w, tpl.Ping, args); err != nil {
		return err
	}
	return nil
}
