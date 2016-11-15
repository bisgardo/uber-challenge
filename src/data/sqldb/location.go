package sqldb

import (
	"src/data/fetch"
	"src/data/types"
	"src/config"
	"src/logging"
	"time"
	"sync"
	"appengine"
)

func LocationNames(coords map[string]*types.Coordinates, delay func (int) int, ctx appengine.Context, logger logging.Logger) {
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
			
			cs, err := fetch.FetchLocationCoordinates(config.MapsApiKey(), name, ctx, logger)
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
