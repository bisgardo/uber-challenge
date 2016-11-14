{{ define "layout" }}

<html>
<head>
	<title>SF Movies {{ if has_field . "Subtitle" }} | {{ .Subtitle }} {{ end }}</title>
	
	<link rel="stylesheet" href="https://cdn.jsdelivr.net/foundation/6.2.4/foundation.min.css">
	<script src="https://code.jquery.com/jquery-3.1.1.min.js" integrity="sha256-hVVnYaiADRTO2PzUGmuLJr8BLUSjGIZsDYGmIJLv2b8=" crossorigin="anonymous"></script>
	<script src="https://cdn.jsdelivr.net/foundation/6.2.4/foundation.min.js"></script>
</head>
<body>

<div class="row">
	<div class="large-12 columns">
		<div class="top-bar">
			<div class="top-bar-left">
				<ul class="menu">
					<li class="menu-text">SF Movies</li>
					<li><a href="/">About</a></li>
					<li><a href="/movie">Movie list</a></li>
					<li><a href="/status">Status</a></li>
				</ul>
			</div>
			<div class="top-bar-right">
				<ul class="menu">
					<li><input type="search" placeholder="Search"></li>
					<li><button type="button" class="button">Search</button></li>
				</ul>
			</div>
		</div>
		
		{{ template "content" . }}
		<hr>
		<footer class="text-left">
			<small><a href="/ping">Ping</a></small>
		</footer>
		<footer class="text-right">
			<small><code>favicon.ico</code> thief-stolen from <a href="https://www.omdbapi.com/">OMDB</a></small>
		</footer>
	</div>
</div>
</body>
{{ end }}
