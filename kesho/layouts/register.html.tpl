	<div class="one-fourth column">
		<div class="menu docs-menu">
			<div class="menu-item selected">Registration</div>
				<form method="POST">
					{{$pid := .primaryID}}
					<div class="menu-item {{with .errs}}{{with $errlist := index . $pid}}has-error{{end}}{{end}}">
						<input type="text" class="form-control" name="{{.primaryID}}" placeholder="{{title .primaryID}}" value="{{.primaryIDValue}}" />
						{{with .errs}}{{with $errlist := index . $pid}}{{range $errlist}}<span class="menu-item flash flash-error">{{.}}</span>{{end}}{{end}}{{end}}
					</div>
					<div class="menu-item {{with .errs}}{{with $errlist := index . "password"}}has-error{{end}}{{end}}">
						<input type="password" class="form-control" name="password" placeholder="Password" value="{{.password}}" />
						{{with .errs}}{{with $errlist := index . "Password"}}{{range $errlist}}<span class="help-block">{{.}}</span>{{end}}{{end}}{{end}}
					</div>
					<div class="menu-item {{with .errs}}{{with $errlist := index . "confirm_password"}}has-error{{end}}{{end}}">
						<input type="password" class="form-control" name="confirm_password" placeholder="Confirm Password" value="{{.confirmPassword}}" />
						{{with .errs}}{{with $errlist := index . "confirm_password"}}{{range $errlist}}<span class="help-block">{{.}}</span>{{end}}{{end}}{{end}}
					</div>
					<input type="hidden" name="{{.xsrfName}}" value="{{.xsrfToken}}" />
						<div class="menu-item">
							<button class="btn btn-primary btn-block" type="submit">Register</button>
						</div>						
				</form>
		</div>
		</div>