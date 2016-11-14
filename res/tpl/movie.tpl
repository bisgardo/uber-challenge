{{ define "content" }}

<h1>{{ .Movie.Title }}</h1>



<table>
	<tr>
		{{ $info := .Info }}
		<td>{{ if $info.Poster }}<img src="{{ $info.Poster }}"/>{{ end }}</td>
		<td>
			<h3>Details (from <a href="https://www.omdbapi.com/">OMDB</a>)</h3>
			<table>
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
					<td>Year</td>
					<td>{{ field $info.Year }}</td>
				</tr>
				<tr>
					<td>Rated</td>
					<td>{{ field $info.Rated }}</td>
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
					<td>Director</td>
					<td>{{ field $info.Director }}</td>
				</tr>
				<tr>
					<td>Writer</td>
					<td>{{ field $info.Writer }}</td>
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
			</table>
		</td>
	</tr>
</table>

{{ end }}
