{{ define "content" }}

<h1>All movies</h1>

<form action="/update" method="post">
	<button class="button">Update</button>
	<p>
		(NOTE: Not yet concurrency-safe because requests hit multiple instances)
	</p>
</form>
<ul>
	{{ range . }}
		<li>
			{{ $m := .Movie}}
			<a href="/movie/{{.Id}}">{{ if $m.Title }}<b>{{ $m.Title }}</b>{{ else }}<i>[No title]</i>{{ end }}</a>
			{{ if $m.Writer}}<i>Written by </i> {{ $m.Writer }}.{{end}}
			{{ $actors := join $m.Actors }}
			{{ if $actors }}<i>Actor(s):</i> {{ $actors }}.{{ end }}
			<ul>
				{{ range $m.Locations }}
					<li>
						{{ .Name }}
						{{ if .FunFact }}{{ parenthesize .FunFact }}{{ end }}
					</li>
				{{ end }}
			</ul>
		</li>
	{{ end }}
</ul>

{{ end }}
