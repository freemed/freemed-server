module github.com/freemed/freemed-server

go 1.25.0

replace (
	github.com/freemed/freemed-server => ./
	github.com/freemed/freemed-server/api => ./api
	github.com/freemed/freemed-server/common => ./common
	github.com/freemed/freemed-server/config => ./config
	github.com/freemed/freemed-server/model => ./model
)

require (
	github.com/freemed/freemed-server/config v0.0.0-00010101000000-000000000000
	github.com/go-sql-driver/mysql v1.10.0
	github.com/golang-migrate/migrate/v4 v4.19.1
)

require (
	filippo.io/edwards25519 v1.2.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/rogpeppe/go-internal v1.16.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
