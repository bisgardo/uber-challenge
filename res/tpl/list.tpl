{{ .LogsCommentBegin }}
    {{ range .Logs }}
		{{.}}
	{{end}}
{{ .LogsCommentEnd }}
<html>
	<head>
		<title>Movies</title>
	</head>
	<body>
		<form action="/update" method="post"><button>Update</button> (crashes on concurrent calls because requests then hit multiple instances)</form>
		<ul>
			{{range .Movies}}
				<li>
					{{ $m := .Movie}}
					<a href="/movie/{{.Id}}">{{if $m.Title}}<b>{{$m.Title}}</b>{{else}}<i>[No title]</i>{{end}}</a>
					{{if $m.Writer}}<i>Written by </i> {{$m.Writer}}.{{end}}
					{{$actors := join $m.Actors}}
					{{if $actors}}<i>Actor(s):</i> {{$actors}}.{{end}}
					<ul>
					{{range $m.Locations}}
						<li>
							{{.Name}}
							{{if .FunFact}}{{parenthesize .FunFact}}{{end}}
						</li>
					{{end}}
					</ul>
				</li>
			{{end}}
		</ul>
		
		<footer>[TODO 'favicon.ico' stolen from...]</footer>
  </body>
</html>
