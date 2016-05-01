# FREEMED SERVER

[![Build Status](https://secure.travis-ci.org/freemed/freemed-server.png)](http://travis-ci.org/freemed/freemed-server)
[![Coverage Status](https://coveralls.io/repos/freemed/freemed-server/badge.svg?branch=master)](https://coveralls.io/r/freemed/freemed-server?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/freemed/freemed-server)](https://goreportcard.com/report/github.com/freemed/freemed-server)
[![GoDoc](https://godoc.org/github.com/freemed/freemed-server?status.png)](https://godoc.org/github.com/freemed/freemed-server)
[![Join the chat at https://gitter.im/gin-gonic/gin](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/freemed/freemed-server?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

Refactoring of **FreeMED** in Golang / Gin.

Uses:

 * [Go](https://golang.org/): Efficient programming language
 * [Gin](https://github.com/gin-gonic/gin/): Web framework for Go
 * [GORP](http://github.com/go-gorp/gorp): Db access layer
 * [Go-MySQL-Driver](http://github.com/go-sql-driver/mysql): MySQL driver
 * [go-redis](https://github.com/go-redis/redis): Redis driver
 * [jQuery](https://jquery.com): Fast Javascript framework
 * [Bootstrap](http://getbootstrap.com): Responsive framework
 * [Bootstrap-Switch](http://www.bootstrap-switch.org): Switches for Bootstrap
 * [Knockout](http://knockoutjs.com/): MVVM UI toolkit
 * [Knockout.Mapping](https://github.com/SteveSanderson/knockout.mapping): Automatic JS object mapping for Knockout
 * [toastr](https://github.com/CodeSeven/toastr): Toaster widget

Code in this repository can be run against a valid FreeMED 0.9.x series database with no modifications.

## Architectural changes from FreeMED 0.9.x

 * **Redis Sessions**. Sessions are stored in Redis, to decrease load on the MySQL server. (TODO: Move to Redis cluster for full redundancy)
 * **Authentication**. Switched from cookies to renewable ``Bearer`` Authorization headers.

## Other CC/Opensource Resources

 * Background image : [CC BY-SA 2.0](https://commons.wikimedia.org/wiki/Category:Medical#/media/File:Laptop_and_stethoscope_%286123892769%29.jpg)

