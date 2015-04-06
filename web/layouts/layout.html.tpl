<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">

	<title>{{template "pagetitle" .}}</title>

    <link href="/static/css/docs.css" rel="stylesheet">
    <link href="/static/css/kesho.css" rel="stylesheet">

    <script src="/static/js/zepto.min.js"></script>
</head>
<body>
<header class="masthead">
	<div class="container">
		<a href="/" class="masthead-logo">
			<span class="mega-octicon octicon-package"></span>
			kesho
		</a>

		<nav class="masthead-nav">
			<a href="/"><i class="icono-home"></i> </a>
			{{if not .loggedin}}
			<a href="/auth/register">Register</a>
			<a href="/auth/login"><i class="fa fa-sign-in"></i> Login</a>
			{{else}}

						<a href="/auth/logout">
							<i class="fa fa-sign-out"></i> Logout
						</a>
			{{end}}
		</nav>
	</div>
</header>

<div class="container">
	<div class="colums docs-layout">
			{{with .flash_success}}<div class="alert alert-success">{{.}}</div>{{end}}
			{{with .flash_error}}<div class="alert alert-danger">{{.}}</div>{{end}}
			{{template "yield" .}}
			{{template "authboss" .}}
	</div>
</div>
</body>
</html>
{{define "pagetitle"}}{{end}}
{{define "yield"}}{{end}}
{{define "authboss"}}{{end}}