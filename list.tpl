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
	    <div>
	        {{ if .OutputLogs }}
	        <ul>
            {{ range .Logs }}
                <li>{{.}}</li>
            {{end}}
            </ul>
            {{ end }}
        </div>
		<form action="update" method="post"><button>Update</button> (crashes on concurrent calls because requests then hit multiple instances)</form>
		<ul>
			{{range .Movies}}
				<li>
					{{if .Title}}<b>{{.Title}}</b>{{else}}<i>[No title]<i>{{end}}
					{{if .Writer}}<i>Written by </i> {{.Writer}}.{{end}}
					{{$actors := join .Actors}}
					{{if $actors}}<i>Actor(s):</i> {{$actors}}.{{end}}
					<ul>
					{{range .Locations}}
						<li>
							{{.Name}}
							{{if .FunFact}}{{parenthesize .FunFact}}{{end}}
						</li>
					{{end}}
					</ul>
				</li>
			{{end}}
		</ul>
		<p>[TODO 'favicon.ico' stolen from...]</p>
  </body>
</html>
