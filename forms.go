package main

type LoginForm struct {
	UserName string `formam:"username" valid:"Required;AlphaNumeric"`
	Password string `formam:"password" valid:"Required"`
}
