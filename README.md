# FREEMED SERVER

[![Build Status](https://github.com/freemed/freemed-server/actions/workflows/go.yml/badge.svg)](https://github.com/freemed/freemed-server/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/freemed/freemed-server)](https://goreportcard.com/report/github.com/freemed/freemed-server)
[![codecov](https://codecov.io/gh/freemed/freemed-server/branch/master/graph/badge.svg)](https://codecov.io/gh/freemed/freemed-server)
[![GoDoc](https://godoc.org/github.com/freemed/freemed-server?status.png)](https://godoc.org/github.com/freemed/freemed-server)
[![Join the chat at https://gitter.im/freemed/freemed-server](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/freemed/freemed-server?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

Refactoring of **FreeMED** in Golang / Gin.

The backend uses:
 * [Go](https://golang.org/): Efficient programming language
 * [Gin](https://github.com/gin-gonic/gin/): Web framework for Go
 * [Gin-JWT](https://github.com/appleboy/gin-jwt): JWT middleware for Gin
 * [GORM](https://gorm.io): DB access layer
 * [Go-MySQL-Driver](http://github.com/go-sql-driver/mysql): MySQL driver
 * [go-redis](https://github.com/go-redis/redis): Redis driver
 * [lumberjack](https://github.com/natefinch/lumberjack): Rolling logger
 * [manners](https://github.com/braintree/manners): Graceful http/https serving

The frontend uses:
 * [jQuery](https://jquery.com): Fast Javascript framework
 * [Popper](https://popper.js.org/): Pop-over management, required by bootstrap
 * [Bootstrap](http://getbootstrap.com): Responsive framework
 * [Bootstrap-Switch](http://www.bootstrap-switch.org): Switches for Bootstrap
 * [Knockout](http://knockoutjs.com/): MVVM UI toolkit
 * [Knockout.Mapping](https://github.com/SteveSanderson/knockout.mapping): Automatic JS object mapping for Knockout
 * [Select2](https://select2.org/): Extensible select widgets
 * [toastr](https://github.com/CodeSeven/toastr): Toaster widget
 
Code in this repository can be run against a valid FreeMED 0.9.x series database with no modifications.

## Caveats

 * MySQL's `ONLY_FULL_GROUP_BY` needs to be disabled -- at least until the queries have been rewritten to no longer require it. This can be temporarily accomplished with `SET GLOBAL sql_mode=(SELECT REPLACE(@@sql_mode,'ONLY_FULL_GROUP_BY',''));`, but should have the actual MySQL server configuration adjusted for it in production systems.

## Architectural changes from FreeMED 0.9.x

 * **Redis Sessions**. Sessions are stored in Redis, to decrease load on the MySQL server. (TODO: Move to Redis cluster for full redundancy)
 * **Authentication**. Switched from cookies to renewable ``Bearer`` Authorization headers.
 * **UI Architecture**. Switched from GWT pre-generated specific javascript to simple jQuery frontend with Bootstrap using RESTful API.

## Other CC/Opensource Resources

 * Background image : [CC BY-SA 2.0](https://commons.wikimedia.org/wiki/File:Laptop_and_stethoscope_\(6123892769\).jpg)

