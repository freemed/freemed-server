module github.com/freemed/freemed-server/config

go 1.19

replace (
	github.com/freemed/freemed-server => ../
	github.com/freemed/freemed-server/common => ../common
)

require gopkg.in/yaml.v2 v2.4.0
