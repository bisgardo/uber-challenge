Uber challenge: San Francisco Movies
------------------------------------

This project is an implementation of the following "Uber Coding Challenge":

[Project description](https://github.com/uber/coding-challenge-tools/blob/master/coding_challenge.md):

> Create a service that shows on a map where movies have been filmed in San Francisco. The user should be able to filter
  the view using autocompletion search.
> 
> The [Film Locations data set](https://data.sfgov.org/Arts-Culture-and-Recreation-/Film-Locations-in-San-Francisco/yitu-d5am)
> is available on [DataSF](http://www.datasf.org/).

### Implementation

The project is implemented in Go and [deployed](https://uber-challenge-148819.appspot.com) on Google's App Engine as my
first experience with both of these technologies.

The data is intended to stored in Cloud SQL, which is an App Engine variant of MySQL.

The project was developed with Go App Engine SDK 1.9.40 (which includes Go 1.6.2) and MySQL 5.7.16 on a Windows 7 box.
At the time of this writing, Cloud SQL ("second generation") is based on MySQL 5.7.

The client side is implemented with the [Foundation](http://foundation.zurb.com/) front-end framework and jQuery and has
been tested in Chrome 54, Firefox 47, and Internet Explorer 11 on Windows 7.

### Life cycle

When the application starts up, it checks if the database is empty and initializes it with data from a cached file if it
is. When the `/update` endpoint is hit with a HTTP POST-request (e.g. using the button on the movie list page at
`/movie`), the database is cleared and reinitialized with fresh data from the data set link above. For robustness, the
"original" initialization also happens if the database is suddenly empty (i.e., we can delete and recreate it from the
console without restarting the application).

The sizes of the tables in the SQL database and the log/error of the last (re)initialization/update are accessible on
the "status" page.

### Features

See the ["About"](https://uber-challenge-148819.appspot.com/) page of the deployed application.

### Setup

0.  Install the Go App Engine SDK (has batteries included) using instructions found elsewhere.
1.  Create an App Engine trial account, give Google your credit card info, and make the sign of the cross.
2.  Create project in the console
3.  Create a Cloud SQL instance in the [web console](https://console.cloud.google.com). Make sure it's in the correct
    location (cannot be changed later) and make sure that your app is authorized to access it (which it should be
    automatically).
4.  Open the Cloud Shell (upper right corner of the console). Connect to the database using the command
    `gcloud beta sql connect [DB-instance-ID] --user=root`. Then create the database `locations` using the MySQL command
    `create database locations;`.
5.  After changing the `application` field of `app.yaml` to the ID of your project, deploy the app using the local
    command `goapp deploy`, executed from the root directory of the project. A browser window will open to log in to
    your Google account and then dump a file named `.appcfg_oauth2_tokens` in your home directory to remove the need for
    doing this again.

Running locally:

*   Install a MySQL server and create the database `locations` from the console using the command
    `create database locations;`.
*   Add a file named `data-source-name` in the `res` (resource) directory. The contents on the file should be a string
    of the format `root:[root-password]@/locations` without any newlines (it is assumed that you connect though the
    server's root user).
*   Obtain a API key to Google Maps and put it in a file (`res/maps-api-key`); also with no whitespace.
*   Start serving the app with the command `goapp serve`. If using IntelliJ, the
    [Go plugin](https://github.com/go-lang-plugin-org) can create a run configuration for doing this.

### Problems

During development, a number of problems were encountered and solved. These lessons learned have been written down in
[`problems.md`](https://github.com/halleknast/uber-challenge/blob/master/problems.md) for later retrieval.

### Missing features

Due to the fact that many new things had to be learned and dealt with to make this project in a limited amount of time,
the following features have been left unimplemented. While they are non-essential for a working prototype, they would
be needed for the project to be production-ready.

*   Testing: The whole system has been manually tested to work as intended, but should of course be covered by a proper
    test suite before new features are implemented. Also, types and functions should be properly documented.
*   There is currently nothing to prevent multiple app instances from updating the database simultaneously. While
    critical parts are done transactionally, weird inconsistencies have been observed. The task should be performed by a
    batch job and be limited in how often it can execute.
*   Geolocations are currently fetched and cached (concurrently) on demand when a movie is loaded. Because of timing
    constraints, it cannot be done for all movies at once (on update) - even with concurrent requests. This fetching
    should be performed in such a way that redundant queries to the Geolocation API are minimized and uniqueness
    constraint violations in the database avoided. Also, negative lookups are not cached.
*   Movie info data should expire such that at least ratings are updated once in a while. Also, the data is currently
    not loaded on initialization and thus requires an "update" action to be performed.

### Other ideas for future work

*   Figure out why Cloud SQL is so slow (simply pinging the database with a trivial query such as `SELECT 42` takes more
    than 100 ms as can be seen on the "ping" page) and try using other storage strategies if it can't be improved.
*   The quality of the data set linked above is quite bad. It could help a lot if users were able add the coordinates of
    a location (e.g. by giving an address), and possibly other pieces of data as well.
*   Add pages for people (actors, writers, and directors) and list the movies that they participated in and the
    locations of these movies.
*   Enable users to find movie locations near some location (e.g. their physical location).
*   Add an element of sightseeing: Show route (e.g. with directions) for a number of locations. This could be all
    locations of a movie, one location of some number of movies, or something else. The user could be able to order the
    locations using drag/drop or the system could compute a shortest route or something. And of course order an Uber car
    for this purpose.
*   Add a social element: Users could add trivia about the locations, times and images of a location appearing in a
    movie, appearing actors, pictures and ratings of locations, recommendations of tours, etc.
*   The main data structure `Movie` (defined in `types.go`) is used in all layers of the application; from parsing JSON
    to rendering in view templates. The data in this structure is "patched" with the data needed by a given view. While
    this approach works well for the project at its current complexity, it probably doesn't scale well.
