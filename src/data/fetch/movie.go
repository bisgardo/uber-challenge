package fetch

import (
	"src/data/types"
	"src/logging"
	"appengine"
	"appengine/urlfetch"
	"io/ioutil"
	"encoding/json"
	"strconv"
	"log"
	"strings"
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

func FetchFromUrl(url string, ctx appengine.Context, logger logging.Logger) ([]types.Movie, error) {
	logger.Infof("Fetching from URL '%s'", url)
	bytes, err := fetchBytes(url, ctx)
	if err != nil {
		return nil, err
	}
	logger.Infof("Fetched %d bytes", len(bytes))
	es, err := fetchEntries(bytes)
	if err != nil {
		return nil, err
	}
	logger.Infof("Resolved %d entries", len(es))
	ms := entriesToMovies(es)
	logger.Infof("Resolved %d movies", len(ms))
	return ms, nil
}

func FetchFromFile(filename string) ([]types.Movie, error) {
	log.Println("Fetching from file: " + filename)
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	
	es, err := fetchEntries(bytes)
	if err != nil {
		return nil, err
	}
	
	return entriesToMovies(es), nil
}

func fetchEntries(bytes []byte) (es []entry, err error) {
	err = json.Unmarshal(bytes, &es)
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

func entriesToMovies(es []entry) []types.Movie {
	// Read entries into map indexed by the movie title.
	m := make(map[string]*types.Movie)
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
	ms := make([]types.Movie, 0, len(m))
	for _, mov := range m {
		ms = append(ms, *mov)
	}
	
	return ms
}

func entryToMovie(e entry) (m types.Movie) {
	// "Location"/"Fun fact" is added in `entryToLocation` below.
	m.Title = cleaned(e.Title)
	
	cleanedActor1 := cleaned(e.Actor_1)
	cleanedActor2 := cleaned(e.Actor_2)
	cleanedActor3 := cleaned(e.Actor_3)
	
	if cleanedActor1 != "" {
		m.Actors = append(m.Actors, e.Actor_1)
	}
	if cleanedActor2 != "" {
		m.Actors = append(m.Actors, e.Actor_2)
	}
	if cleanedActor3 != "" {
		m.Actors = append(m.Actors, e.Actor_3)
	}
	
	m.Director = cleaned(e.Director)
	m.ProductionCompany = cleaned(e.Production_company)
	
	cleanedReleaseYear := cleaned(e.Release_year)
	if cleanedReleaseYear != "" {
		m.ReleaseYear, _ = strconv.Atoi(cleanedReleaseYear)
	}
	
	m.Writer = cleaned(e.Writer)
	
	return
}

func entryToLocation(e entry) (l types.Location) {
	l.Name = cleaned(e.Locations)
	l.FunFact = cleaned(e.Fun_facts)
	return
}

func cleaned(s string) string {
	ts := strings.TrimSpace(s)
	if ts == "N/A" {
		return ""
	}
	return ts
}