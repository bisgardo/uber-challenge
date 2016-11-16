{{ define "layout" }}

<html>
{{ comment_begin }} BEGIN LOGS
	{{ range .Log }}
		{{ . }}
	{{ end }}
END LOGS {{ comment_end }}
<head>
	<title>SF Movies {{ if has_field . "Subtitle" }} | {{ .Subtitle }} {{ end }}</title>
	
	<link rel="stylesheet" href="https://cdn.jsdelivr.net/foundation/6.2.4/foundation.min.css">
	<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/chosen/1.6.2/chosen.css">
	
	<script src="https://code.jquery.com/jquery-3.1.1.min.js" integrity="sha256-hVVnYaiADRTO2PzUGmuLJr8BLUSjGIZsDYGmIJLv2b8=" crossorigin="anonymous"></script>
	<script src="https://cdn.jsdelivr.net/foundation/6.2.4/foundation.min.js"></script>
	<script src="https://cdnjs.cloudflare.com/ajax/libs/chosen/1.6.2/chosen.jquery.min.js"></script>
	<script src="/autocomplete.js"></script>
	
	<script>
		jQuery(function ($) {
			$(document).foundation();
		});
	</script>
</head>
<body>

<div class="row">
	<div class="large-12 columns">
		<div class="top-bar">
			<div class="top-bar-left">
				<ul class="menu">
					<li class="menu-text">SF Movies</li>
					<li><a href="/">About</a></li>
					<li><a href="/movie">All movies</a></li>
					<li><a href="/status">Status</a></li>
				</ul>
			</div>
			<div class="top-bar-right">
				<ul class="menu">
					<li>
						<span id="loading">loading movies...</span>
					</li>
				</ul>
			</div>
		</div>
		
		{{ template "content" .Data }}
		
		<hr>
		<small style="float:left">Version: {{ .Version }} &middot; <a href="/ping">Ping</a> &middot; <a href="/data">Data</a></small>
		<small style="float:right">Favicon thief-stolen from <a href="https://www.omdbapi.com/favicon.ico">OMDB</a></small>
	</div>
</div>
</body>

{{ end }}
