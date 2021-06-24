module github.com/freemed/freemed-server/api

go 1.15

replace (
	github.com/freemed/freemed-server => ../
	github.com/freemed/freemed-server/common => ../common
	github.com/freemed/freemed-server/config => ../config
	github.com/freemed/freemed-server/model => ../model
	github.com/freemed/remitt-server/common => ../../remitt-server/common
	github.com/freemed/remitt-server/config => ../../remitt-server/config
	github.com/freemed/remitt-server/model => ../../remitt-server/model
)

require (
	github.com/freemed/freemed-server/common v0.0.0-00010101000000-000000000000
	github.com/freemed/freemed-server/model v0.0.0-00010101000000-000000000000
	github.com/gin-gonic/gin v1.7.2
)
