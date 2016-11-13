<html>
<head>
	<title>Movie {{ .Movie.Title }}</title>
</head>
<body>

<p>
	<a href="/movie">Back to list</a>
</p>

<h1>{{ .Movie.Title }}</h1>

<table>
	<tr>
		{{ $info := .Info }}
		<td>{{ if $info.Poster }}<img src="{{ $info.Poster }}"/>{{ end }}</td>
		<td>
			<table>
				<tr>
					<td colspan="2"><h3>IMDB</h3></td>
				</tr>
				<tr>
					<td>ID</td>
					<td>{{ if $info.ImdbID }} <a href="http://www.imdb.com/title/{{ $info.ImdbID }}">{{ $info.ImdbID }}</a> {{ else }} <i>N/A</i> {{ end }}</td>
				</tr>
				<tr>
					<td>Rating</td>
					<td>{{ field $info.ImdbRating }} ({{ field $info.ImdbVotes }} votes)</td>
				</tr>
				
				<tr>
					<td colspan="2"><h3>Details</h3></td>
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
					<td>Genre</td>
					<td>{{ field $info.Genre }}</td>
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
					<td>Plot</td>
					<td>{{ field $info.Plot }}</td>
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
					<td>Metascore</td>
					<td>{{ field $info.Metascore }}</td>
				</tr>
			</table>
		</td>
	</tr>
</table>

<footer>
	[TODO 'favicon.ico' stolen from...]
</footer>
</body>
</html>
