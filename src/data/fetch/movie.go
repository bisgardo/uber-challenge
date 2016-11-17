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

func FetchFromUrl(url string, ctx appengine.Context, log logging.Logger) ([]types.Movie, error) {
	log.Infof("Fetching from URL '%s'", url)
	bytes, err := fetchBytes(url, ctx)
	if err != nil {
		return nil, err
	}
	log.Infof("Fetched %d bytes", len(bytes))
	entries, err := fetchEntries(bytes)
	if err != nil {
		return nil, err
	}
	log.Infof("Resolved %d entries", len(entries))
	movies := entriesToMovies(entries)
	log.Infof("Resolved %d movies", len(movies))
	return movies, nil
}

func FetchFromFile(fileName string) ([]types.Movie, error) {
	log.Println("Fetching from file: " + fileName)
	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	
	entries, err := fetchEntries(bytes)
	if err != nil {
		return nil, err
	}
	
	return entriesToMovies(entries), nil
}

func fetchEntries(bytes []byte) (entries []entry, err error) {
	err = json.Unmarshal(bytes, &entries)
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

func entriesToMovies(entries []entry) []types.Movie {
	// Read entries into map indexed by the movie title.
	titleMovieMap := make(map[string]*types.Movie)
	for _, entry := range entries {
		// Parse location data and skip entry if it's empty.
		loc := entryToLocation(entry)
		if loc.Name == "" {
			continue
		}
		
		// Allocate new movie entry if it doesn't exist.
		title := entry.Title
		movie, exists := titleMovieMap[title]
		if !exists {
			m := entryToMovie(entry)
			movie = &m
			titleMovieMap[title] = movie;
		}
		
		// Add location to entry.
		movie.Locations = append(movie.Locations, loc)
	}
	
	// Extract map values to slice...
	movies := make([]types.Movie, 0, len(titleMovieMap))
	for _, movie := range titleMovieMap {
		movies = append(movies, *movie)
	}
	
	return movies
}

func entryToMovie(entry entry) (movie types.Movie) {
	// "Location"/"Fun fact" is added in `entryToLocation` below.
	movie.Title = cleaned(entry.Title)
	
	cleanedActor1 := cleaned(entry.Actor_1)
	cleanedActor2 := cleaned(entry.Actor_2)
	cleanedActor3 := cleaned(entry.Actor_3)
	
	if cleanedActor1 != "" {
		movie.Actors = append(movie.Actors, entry.Actor_1)
	}
	if cleanedActor2 != "" {
		movie.Actors = append(movie.Actors, entry.Actor_2)
	}
	if cleanedActor3 != "" {
		movie.Actors = append(movie.Actors, entry.Actor_3)
	}
	
	movie.Director = cleaned(entry.Director)
	movie.ProductionCompany = cleaned(entry.Production_company)
	
	cleanedReleaseYear := cleaned(entry.Release_year)
	if cleanedReleaseYear != "" {
		movie.ReleaseYear, _ = strconv.Atoi(cleanedReleaseYear)
	}
	
	movie.Writer = cleaned(entry.Writer)
	
	return
}

func entryToLocation(entry entry) (loc types.Location) {
	loc.Name = cleaned(entry.Locations)
	loc.FunFact = cleaned(entry.Fun_facts)
	return
}

func cleaned(str string) string {
	ts := strings.TrimSpace(str)
	if ts == "N/A" {
		return ""
	}
	return ts
}
