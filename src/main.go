package uber_challenge

import (
	"src/data"
	"src/data/types"
	"src/data/sqldb"
	"src/data/fetch"
	"src/logging"
	"src/watch"
	"appengine"
	"appengine/urlfetch"
	"net/http"
	"net/url"
	"html/template"
	"io/ioutil"
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
	
	http.HandleFunc("/", render(movies, false))
	http.HandleFunc("/movie", render(movies, false))
	http.HandleFunc("/movie/", render(movie, false))
	http.HandleFunc("/update", render(update, true))
	http.HandleFunc("/ping", render(func(w http.ResponseWriter, r *http.Request, logger logging.Logger) error {
		sw := watch.NewStopWatch()
		if err := db.Ping(); err != nil {
			return err
		}
		
		args := &struct {
			Clock string
			Time int64
		}{sw.InitTime.String(), sw.TotalElapsedTimeMillis()}
		
		if err := pingTemplate.Execute(w, args); err != nil {
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
	
	args := &struct {
		Movie *types.Movie
		Info  *MovieInfo
	}{&m, &info}
	
	if infoJson, err := sqldb.LoadMovieInfoJson(db, m.Title, logger); infoJson != "" && err == nil {
		// Only attempt to parse JSON if it was loaded successfully
		if err := json.Unmarshal([]byte(infoJson), &info); err != nil {
			logger.Errorf(err.Error())
		}
	}
	
	return movieTemplate.Execute(w, args)
}

func movies(w http.ResponseWriter, _ *http.Request, logger logging.Logger) error {
	defer recordingLogger.Clear()
	
	recordingLogger.Infof("Rendering movie list page")
	
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
	
	m := &struct {
		Logs             []string
		Movies           []types.IdMoviePair
		OutputLogs       bool
		LogsCommentBegin template.HTML
		LogsCommentEnd   template.HTML
	}{recordingLogger.Entries, ms, true, "<!-- <LOGS>", "</LOGS> -->"}
	
	if err := listTemplate.Execute(w, m); err != nil {
		return err
	}
	
	return nil
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
	if err := sqldb.StoreMovies(db, ms, logger); err != nil {
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
		
		infoJson, err := fetchMovieInfo(t, ctx, logger)
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

// TODO Move function...
func fetchMovieInfo(title string, ctx appengine.Context, logger logging.Logger) (string, error) {
	// TODO Sanitize title...
	
	u := "http://www.omdbapi.com/?y=&plot=short&r=json&t=" + url.QueryEscape(title)
	
	sw := watch.NewStopWatch()
	
	logger.Infof("Fetching info for movie '%s' from URL '%s'", title, u)
	client := urlfetch.Client(ctx)
	resp, err := client.Get(u)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	
	logger.Infof("Fetched %d bytes in %d ms", len(bytes), sw.TotalElapsedTimeMillis())
	
	return string(bytes), nil
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
		querySingleInt := func (db *sql.DB, sql string, args ...interface{}) (int, error) {
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
	
	cs := struct {
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
	
	return statusTemplate.Execute(w, cs)
}

// TODO Make convenience function for constructing template...

var pingTemplate = template.Must(template.New("").Parse("<html><p>Clock: {{.Clock}}<p>Ping+render time: {{.Time}} ms</html>"))

var movieTemplate = template.Must(
	template.New("movie.tpl").Funcs(template.FuncMap{
		"field": func (field string) template.HTML {
			if field == "" || field == "N/A" {
				return "<i>N/A</i>"
			}
			return template.HTML(field)
		},
	}).ParseFiles("res/tpl/movie.tpl"),
)

var listTemplate = template.Must(
	// Custom string functions needed because this appears to be the only way to prevent the template engine from
	// littering the output with disruptive whitespace. Also, duplication of template name was necessary for the `Funcs`
	// call to work.
	template.New("list.tpl").Funcs(template.FuncMap{
		"join": func (ss []string) string {
			switch len(ss) {
			case 0:
				return ""
			case 1:
				return ss[0]
			case 2:
				return ss[0] + " and " + ss[1]
			}
			return strings.Join(ss[:len(ss) - 1], ", ") + ", and " + ss[len(ss) - 1]
		},
		"parenthesize": func (s string) string {
			return "(" + s + ")"
		},
	}).ParseFiles("res/tpl/list.tpl"),
)

var statusTemplate = template.Must(template.ParseFiles("res/tpl/status.tpl"))
