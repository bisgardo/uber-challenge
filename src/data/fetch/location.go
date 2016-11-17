package fetch

import (
	"src/logging"
	"sync"
	"time"
	"fmt"
	"net/url"
	"src/watch"
	"appengine/urlfetch"
	"io/ioutil"
	"encoding/json"
	"errors"
	"strings"
	"appengine"
	"src/data/types"
)

func FetchLocationCoordinates(mapsApiKey string, locName string, ctx appengine.Context, logger logging.Logger) (types.Coordinates, error) {
	uri := fmt.Sprintf(
		"https://maps.googleapis.com/maps/api/geocode/json?address=%s,San+Fransisco,+CA&key=%s",
		url.QueryEscape(locName),
		mapsApiKey,
	)
	
	sw := watch.NewStopWatch()
	
	logger.Infof("Fetching coordinates of location '%s' from URL '%s'", locName, uri)
	client := urlfetch.Client(ctx)
	resp, err := client.Get(uri)
	if err != nil {
		return types.Coordinates{}, err
	}
	defer resp.Body.Close()
	
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return types.Coordinates{}, err
	}
	
	var res struct {
		Results []struct {
			Formatted_Address string
			Geometry          struct { Location types.Coordinates }
		}
		Status  string
	}
	if err := json.Unmarshal(bytes, &res); err != nil {
		return types.Coordinates{}, err
	}
	
	logger.Infof("Fetched %d bytes in %d ms", len(bytes), sw.TotalElapsedTimeMillis())
	
	if (len(res.Results) == 0) {
		return types.Coordinates{}, errors.New("Address not found")
	}
	
	result := res.Results[0]
	if res.Status != "OK" || result.Formatted_Address == "California, USA" {
		// Error or generic response. Look for nested address.
		leftParIdx := strings.Index(locName, "(")
		rightParIdx := strings.Index(locName, ")")
		
		if 0 <= leftParIdx && leftParIdx < rightParIdx {
			subLocation := strings.TrimSpace(locName[leftParIdx + 1 : rightParIdx])
			return FetchLocationCoordinates(mapsApiKey, subLocation, ctx, logger)
		}
		
		commaIdx := strings.Index(locName, ",")
		if 0 <= commaIdx && commaIdx < len(locName) {
			subLocation := strings.TrimSpace(locName[commaIdx + 1 : ])
			return FetchLocationCoordinates(mapsApiKey, subLocation, ctx, logger)
		}
		
		// Consider this a non-match.
		return types.Coordinates{}, errors.New("Address not found")
	}
	return result.Geometry.Location, nil
}

func FetchMissingLocationNames(coords map[string]*types.Coordinates, delay func (int) int, mapsApiKey string, ctx appengine.Context, logger logging.Logger) {
	// Fetch geo locations in parallel.
	mutex := &sync.Mutex{}
	ch := make(chan bool)
	
	count := 0
	for n := range coords {
		d := delay(count)
		
		go func (name string) {
			// Allow caller to block on this routine.
			defer func() { ch <- true }()
			
			time.Sleep(time.Duration(d) * time.Millisecond)
			
			cs, err := FetchLocationCoordinates(mapsApiKey, name, ctx, logger)
			if err != nil {
				logger.Infof("Coordinates could not be fetched for location %s", name)
				return
			}
			logger.Infof("Fetched coordinates (%f, %f) for location %s", cs.Lat, cs.Lng, name)
			
			mutex.Lock()
			coords[name] = &cs
			mutex.Unlock()
		}(n)
		count++
	}
	
	// Wait for all go-routines to complete.
	for i := 0; i < count; i++ {
		<- ch
	}
}
