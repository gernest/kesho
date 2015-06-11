# kesho [![Build Status](https://travis-ci.org/gernest/kesho.svg?branch=master)](https://travis-ci.org/gernest/kesho)

A blogging platform in  Go. This is a experiment to work with the [authboss](https://github.com/go-authboss/authboss) authentication library, so far its good.

You can use this to learn about how to use authboss.

## Installation

You must have [golang](http://golang.org) installed and configured `GOPATH` also You should have added `$GOPATH/bin` to system paths.

Firts install godep

	go get github.com/tools/godep

Then clone this repository.

	git clone https://github.com/gernest/kesho

Cd to the cleoned repository

	cd kesho

Build

	godep go build

Run

	./kesho

Open your browser and point to `localhost:8080`` to view the app
