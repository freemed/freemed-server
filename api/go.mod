module github.com/freemed/freemed-server/api

go 1.20

replace (
	github.com/freemed/freemed-server => ../
	github.com/freemed/freemed-server/common => ../common
	github.com/freemed/freemed-server/config => ../config
	github.com/freemed/freemed-server/model => ../model
	github.com/freemed/gokogiri => ../../gokogiri
	github.com/freemed/gokogiri/help => ../../gokogiri/help
	github.com/freemed/gokogiri/xpath => ../../gokogiri/xpath
	github.com/freemed/remitt-server/common => ../../remitt-server/common
	github.com/freemed/remitt-server/config => ../../remitt-server/config
	github.com/freemed/remitt-server/model => ../../remitt-server/model
)

require (
	github.com/freemed/freemed-server/common v0.0.0-00010101000000-000000000000
	github.com/freemed/freemed-server/model v0.0.0-00010101000000-000000000000
	github.com/gin-gonic/gin v1.9.1
)

require (
	github.com/appleboy/gin-jwt/v2 v2.9.1 // indirect
	github.com/bytedance/sonic v1.10.0-rc2 // indirect
	github.com/chenzhuoyu/base64x v0.0.0-20230717121745-296ad89f973d // indirect
	github.com/chenzhuoyu/iasm v0.9.0 // indirect
	github.com/freemed/freemed-server/config v0.0.0-00010101000000-000000000000 // indirect
	github.com/freemed/gokogiri/help v0.0.0-20230628164547-0f93de0487ac // indirect
	github.com/freemed/gokogiri/util v0.0.0-20230628164547-0f93de0487ac // indirect
	github.com/freemed/gokogiri/xml v0.0.0-20230628164547-0f93de0487ac // indirect
	github.com/freemed/gokogiri/xpath v0.0.0-20230628164547-0f93de0487ac // indirect
	github.com/freemed/ratago/xslt v0.0.0-20230628164800-7a1df60588d2 // indirect
	github.com/freemed/remitt-server/common v0.0.0-20230221125127-71aa32d472a1 // indirect
	github.com/freemed/remitt-server/config v0.0.0-20230221125127-71aa32d472a1 // indirect
	github.com/freemed/remitt-server/model v0.0.0-00010101000000-000000000000 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.2 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-gorp/gorp v2.2.0+incompatible // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.14.1 // indirect
	github.com/go-sql-driver/mysql v1.7.1 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/cpuid/v2 v2.2.5 // indirect
	github.com/leodido/go-urn v1.2.4 // indirect
	github.com/mattes/migrate v3.0.1+incompatible // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pelletier/go-toml/v2 v2.0.9 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.11 // indirect
	golang.org/x/arch v0.4.0 // indirect
	golang.org/x/crypto v0.17.0 // indirect
	golang.org/x/net v0.12.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/bsm/ratelimit.v1 v1.0.0-20170922094635-f56db5e73a5e // indirect
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df // indirect
	gopkg.in/gorp.v1 v1.7.2 // indirect
	gopkg.in/redis.v3 v3.6.4 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
