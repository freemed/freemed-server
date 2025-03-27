module github.com/freemed/freemed-server

go 1.24.0

toolchain go1.24.0

replace (
	github.com/freemed/freemed-server => ./
	github.com/freemed/freemed-server/api => ./api
	github.com/freemed/freemed-server/common => ./common
	github.com/freemed/freemed-server/config => ./config
	github.com/freemed/freemed-server/model => ./model
	github.com/freemed/gokogiri => ../gokogiri
	github.com/freemed/gokogiri/help => ../gokogiri/help
	github.com/freemed/gokogiri/xpath => ../gokogiri/xpath
	github.com/freemed/ratago/xslt => ../ratago/xslt
	github.com/freemed/remitt-server => ../remitt-server
	github.com/freemed/remitt-server/client => ../remitt-server/client
	github.com/freemed/remitt-server/common => ../remitt-server/common
	github.com/freemed/remitt-server/config => ../remitt-server/config
	github.com/freemed/remitt-server/model => ../remitt-server/model

	github.com/ugorji/go => github.com/ugorji/go/codec v1.1.7
)
