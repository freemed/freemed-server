module github.com/freemed/freemed-server/model

go 1.15

replace (
	github.com/freemed/freemed-server => ../
	github.com/freemed/freemed-server/common => ../common
	github.com/freemed/freemed-server/config => ../config
	github.com/freemed/remitt-server/client => ../../remitt-server/client
	github.com/freemed/remitt-server/common => ../../remitt-server/common
)

require (
	github.com/go-sql-driver/mysql v1.5.0
	github.com/lib/pq v1.8.0 // indirect
	github.com/mattn/go-sqlite3 v1.14.5 // indirect
	github.com/ziutek/mymysql v1.5.4 // indirect
	gopkg.in/gorp.v1 v1.7.2
)
