<html>
    <p>
        Clock: {{.Clock}}
    </p>
    <p>
        Render time: {{.Time}} ms
    </p>
    
    <table>
        <tr><td>#Movies</td><td>{{.MoviesCount}}</td><td>({{.MoviesTime}} ms)</td></tr>
        <tr><td>#Actors</td><td>{{.ActorsCount}}</td><td>({{.ActorsTime}} ms)</td></tr>
        <tr><td>#Locations</td><td>{{.LocationsCount}}</td><td>({{.LocationsTime}} ms)</td></tr>
        <tr><td>#Movie-actor relations</td><td>{{.MovieActorsCount}}</td><td>({{.MovieActorsTime}} ms)</td></tr>
        <tr><td>#Location coordinates</td><td>{{.CoordinatesCount}}</td><td>({{.CoordinatesTime}} ms)</td></tr>
        <tr><td>#Cached OMDB movie lookups</td><td>{{.OmdbCount}}</td><td>({{.OmdbTime}} ms)</td></tr>
    </table>
    
    <p>
        Init/update err: {{ .RecordedErr }}
    </p>
    <p>
        Init/update log:
        <ul>
            {{ range .RecordedLog }}
                <li>{{ . }}</li>
            {{ end }}
        </ul>
    </p>
</html>
