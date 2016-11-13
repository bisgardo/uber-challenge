package uber_challenge

import (
	"src/data"
	"src/logging"
	"src/watch"
	"strings"
	"time"
	"appengine"
	"appengine/urlfetch"
	"net/http"
	"net/url"
	"html/template"
	"io/ioutil"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
	"encoding/json"
	"fmt"
	"errors"
)

var db *data.LocationDb

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
		t := time.Now()
		if err := db.SqlDb.Ping(); err != nil {
			return err
		}
		if err := pingTemplate.Execute(w, t.String()); err != nil {
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
		db, err = data.Open("mysql", LocalDbSourceName())
		
	} else {
		logger.Infof("Running in production mode")
		db, err = data.Open("mysql", CloudDbSourceName())
	}
	return err
}

func OpenInit(logger logging.Logger) error {
	if err := Open(logger); err != nil {
		return err
	}
	if _, err := db.Init(JsonFileName(), logger); err != nil {
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
	
	m, err := data.LoadMovie(db, mId)
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
	
	model := &struct {
		Movie *data.Movie
		Info  *MovieInfo
	}{&m, &info}
	
	if infoJson, err := data.LoadMovieInfoJson(db, m.Title, logger); infoJson != "" && err == nil {
		// Only attempt to parse JSON if it was loaded successfully
		if err := json.Unmarshal([]byte(infoJson), &info); err != nil {
			logger.Errorf(err.Error())
		}
	}
	
	logger.Infof("%+v", model)
	
	return movieTemplate.Execute(w, model)
}

func movies(w http.ResponseWriter, _ *http.Request, logger logging.Logger) error {
	defer recordingLogger.Clear()
	
	recordingLogger.Infof("Rendering movie list page")
	
	ms, initialized, err := data.LoadMovies(db, JsonFileName(), logger, true)
	if initialized {
		recordInitUpdate(err)
	}
	if err != nil {
		return err
	}
	
	// TODO Block until geo locations and movie data are fetched (if we change 'update' into not blocking...).
	
	logger.Infof("Clearing recording logger")
	
	m := &struct {
		Logs             []string
		Movies           []data.IdMoviePair
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
	
	ms, err := data.FetchFromUrl(ServiceUrl(), ctx, logger)
	if err != nil {
		return err
	}
	if err := data.StoreMovies(db, ms, logger); err != nil {
		return err
	}
	
	// Fetch movie data.
	// TODO Parallelize and use memcached (with expiration) instead of SQL.
	mi, err := data.LoadMovieInfoJsons(db, logger)
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
	if err := data.StoreMovieInfo(db, movieTitleInfo, logger); err != nil {
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
	
	initialized, err := db.IsInitialized()
	if err != nil {
		return err
	}
	
	if initialized {
		mc, err = data.QuerySingleInt(db.SqlDb, "SELECT COUNT(*) FROM movies")
		if err != nil {
			return err
		}
		mt = sw.ElapsedTimeMillis(true)
		
		ac, err = data.QuerySingleInt(db.SqlDb, "SELECT COUNT(*) FROM actors")
		if err != nil {
			return err
		}
		at = sw.ElapsedTimeMillis(true)
		
		lc, err = data.QuerySingleInt(db.SqlDb, "SELECT COUNT(*) FROM locations")
		if err != nil {
			return err
		}
		lt = sw.ElapsedTimeMillis(true)
		
		rc, err = data.QuerySingleInt(db.SqlDb, "SELECT COUNT(*) FROM movies_actors")
		if err != nil {
			return err
		}
		rt = sw.ElapsedTimeMillis(true)
		
		cc, err = data.QuerySingleInt(db.SqlDb, "SELECT COUNT(*) FROM coordinates")
		if err != nil {
			return err
		}
		ct = sw.ElapsedTimeMillis(true)
		
		ic, err = data.QuerySingleInt(db.SqlDb, "SELECT COUNT(*) FROM movie_info")
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

var pingTemplate = template.Must(template.New("").Parse("<html>Clock: {{.}}</html>"))

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
