module github.com/freemed/freemed-server/config

go 1.22

toolchain go1.22.2

replace (
	github.com/freemed/freemed-server => ../
	github.com/freemed/freemed-server/common => ../common
)

require gopkg.in/yaml.v2 v2.4.0
