{{ define "content" }}

<h1>{{ .Movie.Title }}</h1>

<ul class="tabs" data-tabs id="movie-tabs">
	<li class="tabs-title is-active"><a href="#tab-map" aria-selected="true">Map</a></li>
	<li class="tabs-title"><a href="#tab-info">Info</a></li>
</ul>

<div class="tabs-content" data-tabs-content="movie-tabs">
	<div class="tabs-panel is-active" id="tab-map">
		<div class="row">
			<div class="medium-8 columns">
				<div id="map" style="width:100%;height:600px"></div>
			</div>
			<div class="medium-4 columns">
				<h5>{{ len .Movie.Locations }} location(s)</h5>
				<div style="height:600px;overflow:auto">
					{{ range .Movie.Locations }}
						<div class="callout location" data-name="{{ .Name }}" data-lat="{{ .Coordinates.Lat }}" data-lng="{{ .Coordinates.Lng }}">
							{{ .Name }}
							{{ if .FunFact }}
								<hr>
								<em>
									{{ .FunFact }}
								</em>
							{{ end }}
						</div>
					{{ end }}
				</div>
			</div>
		</div>
	</div>
	<div class="tabs-panel" id="tab-info">
		<div class="row">
			<div class="row">
				<div class="medium-4 columns">
				</div>
				<div class="medium-8 columns">
					<h3>Details (from <a href="https://www.omdbapi.com/">OMDB</a>)</h3>
				</div>
			</div>
			<div class="medium-4 columns">
				{{ $info := .Info }}
				{{ $poster := $info.Poster }}
				{{ if ne $poster "N/A" }}<img src="{{ $poster }}"/>{{ end }}
			</div>
			<div class="medium-8 columns">
				<table>
					<tr>
						<td>Title</td>
						<td>{{ field $info.Title }}</td>
					</tr>
					<tr>
						<td>IMDB ID</td>
						<td>{{ if $info.ImdbID }} <a href="http://www.imdb.com/title/{{ $info.ImdbID }}">{{ $info.ImdbID }}</a> {{ else }} <i>N/A</i> {{ end }}</td>
					</tr>
					<tr>
						<td>IMDB Rating</td>
						<td>{{ field $info.ImdbRating }} ({{ field $info.ImdbVotes }} votes)</td>
					</tr>
					<tr>
						<td>Metascore</td>
						<td>{{ field $info.Metascore }}</td>
					</tr>
					<tr>
						<td>Genre</td>
						<td>{{ field $info.Genre }}</td>
					</tr>
					<tr>
						<td>Plot</td>
						<td>{{ field $info.Plot }}</td>
					</tr>
					<tr>
						<td>Writer</td>
						<td>{{ field $info.Writer }}</td>
					</tr>
					<tr>
						<td>Director</td>
						<td>{{ field $info.Director }}</td>
					</tr>
					<tr>
						<td>Actors</td>
						<td>{{ field $info.Actors }}</td>
					</tr>
					<tr>
						<td>Language</td>
						<td>{{ field $info.Language }}</td>
					</tr>
					<tr>
						<td>Country</td>
						<td>{{ field $info.Country }}</td>
					</tr>
					<tr>
						<td>Awards</td>
						<td>{{ field $info.Awards }}</td>
					</tr>
					<tr>
						<td>Year</td>
						<td>{{ field $info.Year }}</td>
					</tr>
					<tr>
						<td>Released</td>
						<td>{{ field $info.Released }}</td>
					</tr>
					<tr>
						<td>Runtime</td>
						<td>{{ field $info.Runtime }}</td>
					</tr>
					<tr>
						<td>Rated</td>
						<td>{{ field $info.Rated }}</td>
					</tr>
				</table>
			</div>
		</div>
	</div>
</div>

<script src="/location-map.js"></script>
<script async defer src="https://maps.googleapis.com/maps/api/js?key={{ maps_api_key }}&callback=initMap"></script>

{{ end }}
