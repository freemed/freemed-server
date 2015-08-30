# FREEMED SERVER

[![Build Status](https://secure.travis-ci.org/freemed/freemed-server.png)](http://travis-ci.org/freemed/freemed-server)

[![GoDoc](https://godoc.org/github.com/freemed/freemed-server?status.png)](https://godoc.org/github.com/freemed/freemed-server)

Refactoring of **FreeMED** in Golang / Martini.

Uses:

 * [Go](https://golang.org/): Efficient programming language
 * [Martini](http://martini.codegangsta.io/): Web framework for Go
 * [GORP](http://github.com/go-gorp/gorp): Db access layer
 * [Go-MySQL-Driver](http://github.com/go-sql-driver/mysql): MySQL driver
 * [go-redis](https://github.com/go-redis/redis): Redis driver
 * [jQuery](https://jquery.com): Fast Javascript framework
 * [Bootstrap](http://getbootstrap.com): Responsive framework
 * [Bootstrap-Switch](http://www.bootstrap-switch.org): Switches for Bootstrap

Code in this repository can be run against a valid FreeMED 0.9.x series database with no modifications.

## Architectural changes from FreeMED 0.9.x

 * **Redis Sessions**. Sessions are stored in Redis, to decrease load on the MySQL server. (TODO: Move to Redis cluster for full redundancy)
 * **Authentication**. Switched from cookies to renewable ``Bearer`` Authorization headers.

