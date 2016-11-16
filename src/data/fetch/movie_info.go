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

func FetchMovieInfo(title string, ctx appengine.Context, logger logging.Logger) (string, error) {
	// Sanitize movie title.
	regex, err := regexp.Compile("(?i)\\s*(-|,|season).*")
	if err != nil {
		panic(err)
	}
	
	sanitizedTitle := regex.ReplaceAllString(title, "")
	
	u := "http://www.omdbapi.com/?y=&plot=short&r=json&t=" + url.QueryEscape(sanitizedTitle)
	
	sw := watch.NewStopWatch()
	
	logger.Infof("Fetching info for movie '%s' ('%s') from URL '%s'", title, sanitizedTitle, u)
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
