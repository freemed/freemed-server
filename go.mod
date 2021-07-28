module github.com/freemed/freemed-server

go 1.15

replace (
	github.com/freemed/freemed-server/api => ./api
	github.com/freemed/freemed-server/common => ./common
	github.com/freemed/freemed-server/config => ./config
	github.com/freemed/freemed-server/model => ./model
	github.com/freemed/remitt-server/client => ../remitt-server/client
	github.com/freemed/remitt-server/common => ../remitt-server/common
	github.com/freemed/remitt-server/config => ../remitt-server/config
	github.com/freemed/remitt-server/model => ../remitt-server/model
)

require (
	github.com/appleboy/gin-jwt/v2 v2.6.4
	github.com/braintree/manners v0.0.0-20160418043613-82a8879fc5fd
	github.com/freemed/freemed-server/api v0.0.0-00010101000000-000000000000
	github.com/freemed/freemed-server/common v0.0.0-00010101000000-000000000000
	github.com/freemed/freemed-server/config v0.0.0-00010101000000-000000000000
	github.com/freemed/freemed-server/model v0.0.0-00010101000000-000000000000
	github.com/gin-gonic/contrib v0.0.0-20201101042839-6a891bf89f19
	github.com/gin-gonic/gin v1.7.2
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
)
