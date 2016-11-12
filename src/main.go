package uber_challenge

import (
	"src/data"
	"src/logging"
	"net/http"
	"html/template"
	"strings"
	"appengine"
	"time"
	_ "github.com/go-sql-driver/mysql"
)

var db *data.LocationDb

var recordingLogger logging.RecordingLogger

var recordedLog []string
var recordedError error

func init() {
	var logger logging.InitLogger
	recordingLogger.Wrap(&logger)
	defer recordingLogger.Unwrap()
	
	recordingLogger.Infof("Spinning up instance with ID '%v'", appengine.InstanceID())
	
	if err := OpenInit(&recordingLogger); err != nil {
		recordInitUpdate(err)
		panic(err)
	}
	
	recordInitUpdate(nil)
	
	http.HandleFunc("/", root)
	http.HandleFunc("/update", update)
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()
		// TODO *Just* ping database.
		template.Must(template.New("").Parse("<html>Clock: {{.}}</html>")).Execute(w, t.String())
	})
	
	// TODO Make "raw data dump" page.
	
	// TODO Add pages for movie, actor, ...
	
	http.HandleFunc("/status", status)
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

func root(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	recordingLogger.Wrap(ctx)
	defer recordingLogger.Unwrap()
	defer recordingLogger.Clear()
	
	if err := renderMovies(w, &recordingLogger); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func renderMovies(w http.ResponseWriter, logger logging.Logger) error {
	recordingLogger.Infof("Rendering movie list page")
	
	ms, initialized, err := db.LoadMovies(JsonFileName(), logger)
	if initialized {
		recordInitUpdate(err)
	}
	if err != nil {
		return err
	}
	
	// TODO Block until geo locations and movie data are fetched
	
	// TODO Add links to pages for movie, actor, ...
	
	m := &struct {
		Logs             *[]string
		Movies           *[]data.Movie
		OutputLogs       bool
		LogsCommentBegin template.HTML
		LogsCommentEnd   template.HTML
	}{&recordingLogger.Entries, ms, true, "<!-- <LOGS>", "</LOGS> -->"}
	
	if err := listTemplate.Execute(w, m); err != nil {
		return err
	}
	
	return nil
}

func update(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Cannot " + r.Method + " '/load'", http.StatusMethodNotAllowed)
		return
	}
	
	// TODO Need to add timestamp(s) to DB for locking to work across instances.
	
	ctx := appengine.NewContext(r)
	recordingLogger.Wrap(ctx)
	defer recordingLogger.Unwrap()
	
	data.InitUpdateMutex.Lock()
	defer data.InitUpdateMutex.Unlock()
	
	logger := &recordingLogger
	logger.Infof("Update started")
	
	ms := data.FetchFromUrl(ServiceUrl(), ctx, logger)
	err := data.StoreMovies(db, ms, logger)
	
	recordInitUpdate(err)
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// TODO Fetch geo locations and movie data.
	
	logger.Infof("Update completed")
	
	http.Redirect(w, r, "", http.StatusFound)
}

func status(w http.ResponseWriter, r *http.Request) {
	if err := renderStatus(w, r); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func renderStatus(w http.ResponseWriter, r *http.Request) error {
	ctx := appengine.NewContext(r)
	recordingLogger.Wrap(ctx)
	defer recordingLogger.Unwrap()
	
	recordingLogger.Infof("Rendering status page")
	
	sw := NewStopWatch()
	
	mc := 0
	ac := 0
	lc := 0
	rc := 0
	cc := 0
	oc := 0
	
	mt := int64(0)
	at := int64(0)
	lt := int64(0)
	rt := int64(0)
	ct := int64(0)
	ot := int64(0)
	
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
		
		oc, err = data.QuerySingleInt(db.SqlDb, "SELECT COUNT(*) FROM omdb")
		if err != nil {
			return err
		}
		ot = sw.ElapsedTimeMillis(true)
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
		OmdbCount        int
		OmdbTime         int64
		RecordedErr      error
		RecordedLog      []string
	}{sw.InitTime.String(), dt, mc, mt, ac, at, lc, lt, rc, rt, cc, ct, oc, ot, recordedError, recordedLog}
	
	return statusTemplate.Execute(w, cs)
}

var listTemplate = template.Must(
	// Custom string functions needed because this appears to be the only way to prevent the template engine from
	// littering the output with disruptive whitespace. Also, duplication of template name was necessary for the `Funcs`
	// call to work.
	template.New("list.tpl").Funcs(template.FuncMap{
		"join": func(ss []string) string {
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
		"parenthesize": func(s string) string {
			return "(" + s + ")"
		},
	}).ParseFiles("res/tpl/list.tpl"),
)

var statusTemplate = template.Must(template.ParseFiles("res/tpl/status.tpl"))
