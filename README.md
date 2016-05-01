# FREEMED SERVER

[![Build Status](https://secure.travis-ci.org/freemed/freemed-server.png)](http://travis-ci.org/freemed/freemed-server)

[![GoDoc](https://godoc.org/github.com/freemed/freemed-server?status.png)](https://godoc.org/github.com/freemed/freemed-server)

Refactoring of **FreeMED** in Golang / Martini.

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

