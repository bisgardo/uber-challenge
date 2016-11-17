package fetch

import (
	"src/logging"
	"src/watch"
	"appengine"
	"appengine/urlfetch"
	"io/ioutil"
	"net/url"
	"regexp"
)

func FetchMovieInfo(title string, ctx appengine.Context, log logging.Logger) (string, error) {
	// Sanitize movie title.
	// TODO Should cache compiled regex.
	regex, err := regexp.Compile("(?i)\\s*(-|,|season).*")
	if err != nil {
		panic(err)
	}
	
	sanitizedTitle := regex.ReplaceAllString(title, "")
	
	uri := "http://www.omdbapi.com/?y=&plot=short&r=json&t=" + url.QueryEscape(sanitizedTitle)
	
	sw := watch.NewStopWatch()
	
	log.Infof("Fetching info for movie '%s' ('%s') from URL '%s'", title, sanitizedTitle, uri)
	client := urlfetch.Client(ctx)
	resp, err := client.Get(uri)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	
	log.Infof("Fetched %d bytes in %d ms", len(bytes), sw.TotalElapsedTimeMillis())
	
	return string(bytes), nil
}
