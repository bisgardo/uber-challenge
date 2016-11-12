package uber_challenge

import "io/ioutil"

func JsonFileName() string {
	return "res/data/wwmu-gmzc.json";
}

func ServiceUrl() string {
	// Using 'http' instead of 'https' because App Engine will otherwise complain about the SSL certificate being invalid.
	return "http://data.sfgov.org/resource/wwmu-gmzc.json";
}

func LocalDbSourceName() string {
	bytes, err := ioutil.ReadFile("res/data-source-name")
	if err != nil {
		panic(err)
	}
	return string(bytes)
}

func CloudDbSourceName() string {
	return "root@cloudsql(uber-challenge-148819:europe-west1:movie-locations)/locations"
}
