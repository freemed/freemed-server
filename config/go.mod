module github.com/freemed/freemed-server/config

go 1.24

toolchain go1.24.0

replace (
	github.com/freemed/freemed-server => ../
	github.com/freemed/freemed-server/common => ../common
)

require gopkg.in/yaml.v2 v2.4.0

require (
	github.com/kr/pretty v0.3.1 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)
