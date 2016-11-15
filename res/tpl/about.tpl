{{ define "content" }}

<h1>SF Movies</h1>

<p>
	This site is a deployed instance of a project in "Uber Coding Challenge" about listing locations in San Fransisco
	where movies have been filmed. The source code and implementation details of the project may be found on
	<a href="https://github.com/halleknast/uber-challenge">GitHub</a>.
</p>

<h2>Usage</h2>

<p>
	To view the filming locations of some movie, browse or search the drop-down in the upper right corner. An
	alternative is to find it on the <a href="/movie">all movies</a> page, which also contains a button for updating the
	database.
</p>
<p>
	Movie pages have two tabs: one (default) for listing the filming locations and displaying them on a map. The map
	markers will animate whenever the mouse hovers the box for their locations (and vice versa). Boxes whose location
	coordinates could not be found are colored yellow. If the application could find it, the other tab contains a poster
	and some info about the movie.
</p>
<p>
	Note that due to the poor quality of the movie/location data, both movie info and coordinates may be wrong or
	missing. The application is able to handle some cases, but not all.
</p>
<p>
	The <a href="/status">status</a> page shows some info about the application that is probably only useful during
	development. The same goes with the <a href="/ping">ping</a> page linked in the bottom of the layout. Besides it is
	the <a href="/data">Data</a> link for retrieving (most of) the database in JSON format.
</p>

{{ end }}
