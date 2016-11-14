{{ define "content" }}

<h1>Ping</h1>
<h2>Time</h2>
<table>
	<tr><td>Clock</td><td>{{ .Clock }}</td></tr>
	<tr><td>Database ping time</td><td>{{ .Time }} ms</td></tr>
</table>

{{ end }}
