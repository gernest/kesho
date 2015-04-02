<div class="one-fourth column">
	<div class="menu docs-menu">
		<div class="menu-item selected">
			login
		</div>
		{{if .error}}
		<div class="menu-item">
					<div class="flash flash-error">{{.error}}</div>
		</div>
		{{end}}
		<form method="POST">
			<div class="menu-item">
				<input type="text" class="form-control" name="{{.primaryID}}" placeholder="{{title .primaryID}}" value="{{.primaryIDValue}}">
			</div>
			<div class="menu-item">
				<input  type="password" class="form-control" name="password" placeholder="Password">
			</div>
			{{if .showRemember}}
			<div class="menu-item">
				<input type="checkbox" name="rm" value="true"> Remember Me
			</div>
			{{end}}
			<input type="hidden" name="{{.xsrfName}}" value="{{.xsrfToken}}" />
				<div class="menu-item">
					<button class="btn btn-primary btn-block" type="submit">Login</button>
				</div>
			{{if .showRecover}}
			<div class="row">
				<div class="col-md-offset-1 col-md-10">
					<a class="btn btn-link btn-block" href="{{mountpathed "recover"}}">Recover Account</a>
				</div>
			</div>
			{{end}}
		</form>
	</div>	
</div>
				
			
