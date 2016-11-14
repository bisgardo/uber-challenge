package uber_challenge

import (
	"src/data"
	"src/data/types"
	"src/data/sqldb"
	"src/data/fetch"
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

var recordingLogger logging.RecordingLogger

var recordedLog []string
var recordedError error

func init() {
	var logger logging.InitLogger
	recordingLogger.Wrap(&logger)
	defer recordingLogger.Unwrap()
	
	recordingLogger.Infof("Spinning up instance with ID '%s'", appengine.InstanceID())
	
	if err := OpenInit(&recordingLogger); err != nil {
		recordInitUpdate(err)
		panic(err)
	}
	
	recordInitUpdate(nil)
	
	http.HandleFunc("/", render(front, false))
	http.HandleFunc("/movie", render(movies, false))
	http.HandleFunc("/data", dataJson)
	http.HandleFunc("/movie/", render(movie, false))
	http.HandleFunc("/update", render(update, true))
	http.HandleFunc("/ping", render(func(w http.ResponseWriter, r *http.Request, logger logging.Logger) error {
		sw := watch.NewStopWatch()
		if err := db.Ping(); err != nil {
			return err
		}
		
		ctx := appengine.NewContext(r)
		version := appengine.VersionID(ctx)
		args := &struct {
			Clock   string
			Time    int64
			Version string
		}{sw.InitTime.String(), sw.TotalElapsedTimeMillis(), version}
		
		if err := tpl.Render(w, tpl.Ping, args); err != nil {
			return err
		}
		return nil
	}, false))
	
	// TODO Make "raw data dump" page.
	
	// TODO Add pages for movie, actor, ...
	
	http.HandleFunc("/status", renderStatus)
}

func recordInitUpdate(err error) {
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
		db, err = sql.Open("mysql", LocalDbSourceName())
	} else {
		logger.Infof("Running in production mode")
		db, err = sql.Open("mysql", CloudDbSourceName())
	}
	return err
}

func OpenInit(logger logging.Logger) error {
	if err := Open(logger); err != nil {
		return err
	}
	if _, err := data.Init(db, JsonFileName(), logger); err != nil {
		return err
	}
	return nil
}

func render(renderer func(w http.ResponseWriter, r *http.Request, logger logging.Logger) error, record bool) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := appengine.NewContext(r)
		recordingLogger.Wrap(ctx)
		defer recordingLogger.Unwrap()
		
		if err := renderer(w, r, &recordingLogger); err != nil {
			if record {
				recordInitUpdate(err)
			}
			ctx.Errorf("ERROR: %+v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		if record {
			recordInitUpdate(nil)
		}
	}
}

func front(w http.ResponseWriter, r *http.Request, logger logging.Logger) error {
	ctx := appengine.NewContext(r)
	version := appengine.VersionID(ctx)
	args := &struct {
		Version string
	}{version}
	
	return tpl.Render(w, tpl.About, args)
}

func movie(w http.ResponseWriter, r *http.Request, logger logging.Logger) error {
	defer func() {
		logger.Infof("Clearing recording logger")
		recordingLogger.Clear()
	}()
	
	path := r.URL.Path
	i := strings.LastIndex(path, "/")
	id := path[i + 1:]
	
	mId, err := strconv.Atoi(id)
	if err != nil {
		return errors.New(fmt.Sprintf("Invalid movie ID '%s'", id))
	}
	
	logger.Infof("Rendering movie with ID %d", mId)
	
	m, err := sqldb.LoadMovie(db, mId)
	if err != nil {
		http.Error(w, fmt.Sprintf("Movie with ID %d not found", mId), http.StatusNotFound)
		return nil
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
	
	ctx := appengine.NewContext(r)
	version := appengine.VersionID(ctx)
	args := &struct {
		Subtitle string
		Movie    *types.Movie
		Info     *MovieInfo
		Version  string
	}{info.Title, &m, &info, version}
	
	if infoJson, err := sqldb.LoadMovieInfoJson(db, m.Title, logger); infoJson != "" && err == nil {
		// Only attempt to parse JSON if it was loaded successfully
		if err := json.Unmarshal([]byte(infoJson), &info); err != nil {
			logger.Errorf(err.Error())
		}
	}
	
	return tpl.Render(w, tpl.Movie, args)
}

func movies(w http.ResponseWriter, r *http.Request, logger logging.Logger) error {
	defer recordingLogger.Clear()
	
	logger.Infof("Rendering movie list page")
	
	// Check if database is initialized and load from file if it isn't.
	initialized, err := data.Init(db, JsonFileName(), logger)
	if initialized {
		recordInitUpdate(err)
	}
	if err != nil {
		return err
	}
	
	ms, err := sqldb.LoadMovies(db, logger, true)
	if err != nil {
		return err
	}
	
	// TODO Block until geo locations and movie data are fetched (if we change 'update' into not blocking...).
	
	logger.Infof("Clearing recording logger")
	
	ctx := appengine.NewContext(r)
	version := appengine.VersionID(ctx)
	
	args := &struct {
		Logs       []string
		Movies     []types.IdMoviePair
		OutputLogs bool
		Version    string
	}{recordingLogger.Entries, ms, true, version}
	
	if err := tpl.Render(w, tpl.Movies, args); err != nil {
		return err
	}
	
	return nil
}

func dataJson(w http.ResponseWriter, r *http.Request) {
	//time.Sleep(1000000000)
	
	ctx := appengine.NewContext(r)
	
	// TODO Only load movies...
	ms, err := sqldb.LoadMovies(db, ctx, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	
	if err := json.NewEncoder(w).Encode(ms); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func update(w http.ResponseWriter, r *http.Request, logger logging.Logger) error {
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
	
	ms, err := fetch.FetchFromUrl(ServiceUrl(), ctx, logger)
	if err != nil {
		return err
	}
	if err := sqldb.InitTablesAndStoreMovies(db, ms, logger); err != nil {
		return err
	}
	
	// Fetch movie data.
	// TODO Parallelize and use memcached (with expiration) instead of SQL.
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
	
	// TODO Fetch geo locations.
	
	http.Redirect(w, r, "", http.StatusFound)
	return nil
}

func renderStatus(w http.ResponseWriter, r *http.Request) {
	if err := status(w, r); err != nil {
		ctx := appengine.NewContext(r)
		ctx.Errorf(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func status(w http.ResponseWriter, r *http.Request) error {
	ctx := appengine.NewContext(r)
	recordingLogger.Wrap(ctx)
	defer recordingLogger.Unwrap()
	
	recordingLogger.Infof("Rendering status page")
	
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
		Version          string
	}{sw.InitTime.String(), dt, mc, mt, ac, at, lc, lt, rc, rt, cc, ct, ic, it, recordedError, recordedLog, version}
	
	return tpl.Render(w, tpl.Status, args)
}
