package data

import (
	"io/ioutil"
	"encoding/json"
	"appengine/urlfetch"
	"strconv"
	"log"
	"logging"
	"appengine"
)

type entry struct {
	Actor_1            string
	Actor_2            string
	Actor_3            string
	
	Director           string
	Locations          string
	Fun_facts          string
	Production_company string
	Release_year       string
	Title              string
	Writer             string
}

type Movie struct {
	Title             string
	Locations         []Location
	Actors            []string
	Director          string
	Distributor       string
	Writer            string
	ProductionCompany string
	ReleaseYear       int
}

type Location struct {
	Name    string
	FunFact string
}

// Comparator for sorting movie list.
type ByTitle []Movie

func (ms ByTitle) Len() int {
	return len(ms)
}
func (ms ByTitle) Swap(i, j int) {
	ms[i], ms[j] = ms[j], ms[i]
}
func (ms ByTitle) Less(i, j int) bool {
	return ms[i].Title < ms[j].Title
}

func FetchFromUrl(url string, ctx appengine.Context, logger logging.Logger) []Movie {
	logger.Infof("Fetching from URL '%v'", url)
	bytes, err := fetchBytes(url, ctx)
	if err != nil {
		panic(err)
	}
	logger.Infof("Fetched %d bytes", len(bytes))
	es := fetchEntries(bytes)
	logger.Infof("Resolved %d entries", len(es))
	ms := entriesToMovies(es)
	logger.Infof("Resolved %d movies", len(ms))
	return ms
}

func FetchFromFile(filename string) []Movie {
	log.Println("Fetching from file: " + filename)
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	
	es := fetchEntries(bytes)
	return entriesToMovies(es)
}

func fetchEntries(bytes []byte) (es []entry) {
	err := json.Unmarshal(bytes, &es)
	if err != nil {
		panic(err)
	}
	return
}

func fetchBytes(url string, ctx appengine.Context) ([]byte, error) {
	client := urlfetch.Client(ctx)
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	return ioutil.ReadAll(resp.Body)
}

func entriesToMovies(es []entry) []Movie {
	// Read entries into map indexed by the movie title.
	m := make(map[string]*Movie)
	for _, e := range es {
		// Parse location data and skip entry if it's empty.
		l := entryToLocation(e)
		if l.Name == "" {
			continue
		}
		
		// Allocate new movie entry if it doesn't exist.
		title := e.Title
		mov, exists := m[title]
		if !exists {
			tmp := entryToMovie(e)
			mov = &tmp
			m[title] = mov;
		}
		
		// Add location to entry.
		mov.Locations = append(mov.Locations, l)
	}
	
	// Extract map values to slice...
	ms := make([]Movie, 0, len(m))
	for _, mov := range m {
		ms = append(ms, *mov)
	}
	
	return ms
}

func entryToMovie(e entry) (m Movie) {
	// Location/fun fact is added in `entryToLocation`.
	if defined(e.Title) {
		m.Title = e.Title
	}
	if defined(e.Actor_1) {
		m.Actors = append(m.Actors, e.Actor_1)
	}
	if defined(e.Actor_2) {
		m.Actors = append(m.Actors, e.Actor_2)
	}
	if defined(e.Actor_3) {
		m.Actors = append(m.Actors, e.Actor_3)
	}
	if defined(e.Director) {
		m.Director = e.Director
	}
	if defined(e.Production_company) {
		m.ProductionCompany = e.Production_company
	}
	if defined(e.Release_year) {
		m.ReleaseYear, _ = strconv.Atoi(e.Release_year)
	}
	if defined(e.Writer) {
		m.Writer = e.Writer
	}
	return
}

func entryToLocation(e entry) (l Location) {
	if defined(e.Locations) {
		l.Name = e.Locations
	}
	if defined(e.Fun_facts) {
		l.FunFact = e.Fun_facts
	}
	return
}

func defined(s string) bool {
	return s != "" && s != "N/A"
}
