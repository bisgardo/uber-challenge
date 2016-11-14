{{ define "content" }}

<h1>Status</h1>

<h2>Time</h2>
<table>
	<tr><td>Clock</td><td>{{ .Clock }}</td></tr>
	<tr><td>Render time</td><td>{{ .Time }} ms</td></tr>
</table>

<h2>Database</h2>
<table>
	<tr>
		<th>Table</th>
		<th>Size</th>
		<th>Query time</th>
	</tr>
	<tr>
		<td>#Movies</td>
		<td>{{ .MoviesCount }}</td>
		<td>({{ .MoviesTime }} ms)</td>
	</tr>
	<tr>
		<td>#Actors</td>
		<td>{{ .ActorsCount }}</td>
		<td>({{ .ActorsTime }} ms)</td>
	</tr>
	<tr>
		<td>#Locations</td>
		<td>{{ .LocationsCount }}</td>
		<td>({{ .LocationsTime }} ms)</td>
	</tr>
	<tr>
		<td>#Movie-actor relations</td>
		<td>{{ .MovieActorsCount }}</td>
		<td>({{ .MovieActorsTime }} ms)</td>
	</tr>
	<tr>
		<td>#Location coordinates</td>
		<td>{{ .CoordinatesCount }}</td>
		<td>({{ .CoordinatesTime }} ms)</td>
	</tr>
	<tr>
		<td>#Cached movie info lookups</td>
		<td>{{ .InfoCount }}</td>
		<td>({{ .InfoTime }} ms)</td>
	</tr>
</table>

<h2>Init/update</h2>
<h3>Error</h3>
<pre>{{ .RecordedErr }}</pre>
<h3>Log</h3>
<ul>
	{{ range .RecordedLog }}
	<li>{{ . }}</li>
	{{ end }}
</ul>

{{ end }}
