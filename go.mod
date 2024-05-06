module github.com/freemed/freemed-server

go 1.22

toolchain go1.22.2

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

require (
	github.com/appleboy/gin-jwt/v2 v2.9.2
	github.com/braintree/manners v0.0.0-20160418043613-82a8879fc5fd
	github.com/freemed/freemed-server/api v0.0.0-20230213182720-f119baf49957
	github.com/freemed/freemed-server/common v0.0.0-20240127174730-c59888f28f54
	github.com/freemed/freemed-server/config v0.0.0-20240127174730-c59888f28f54
	github.com/freemed/freemed-server/model v0.0.0-20240127174730-c59888f28f54
	github.com/gin-gonic/contrib v0.0.0-20221130124618-7e01895a63f2
	github.com/gin-gonic/gin v1.9.1
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/bytedance/sonic v1.11.6 // indirect
	github.com/chenzhuoyu/base64x v0.0.0-20230717121745-296ad89f973d // indirect
	github.com/chenzhuoyu/iasm v0.9.1 // indirect
	github.com/freemed/gokogiri/help v0.0.0-20230628164547-0f93de0487ac // indirect
	github.com/freemed/gokogiri/util v0.0.0-20230628164547-0f93de0487ac // indirect
	github.com/freemed/gokogiri/xml v0.0.0-20230628164547-0f93de0487ac // indirect
	github.com/freemed/gokogiri/xpath v0.0.0-20230628164547-0f93de0487ac // indirect
	github.com/freemed/ratago v0.0.0-20191105200024-660929a3e119 // indirect
	github.com/freemed/remitt-server/common v0.0.0-20240315224111-230a584acffa // indirect
	github.com/freemed/remitt-server/config v0.0.0-20240315224111-230a584acffa // indirect
	github.com/freemed/remitt-server/model v0.0.0-20240315224111-230a584acffa // indirect
	github.com/gabriel-vasile/mimetype v1.4.3 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-gorp/gorp v2.2.0+incompatible // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.20.0 // indirect
	github.com/go-sql-driver/mysql v1.8.1 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/cpuid/v2 v2.2.7 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/mattes/migrate v3.0.1+incompatible // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pelletier/go-toml/v2 v2.2.2 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	golang.org/x/arch v0.8.0 // indirect
	golang.org/x/crypto v0.23.0 // indirect
	golang.org/x/net v0.25.0 // indirect
	golang.org/x/sys v0.20.0 // indirect
	golang.org/x/text v0.15.0 // indirect
	google.golang.org/protobuf v1.34.1 // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/bsm/ratelimit.v1 v1.0.0-20170922094635-f56db5e73a5e // indirect
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df // indirect
	gopkg.in/gorp.v1 v1.7.2 // indirect
	gopkg.in/redis.v3 v3.6.4 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gorm.io/driver/mysql v1.5.6 // indirect
	gorm.io/gorm v1.25.10 // indirect
)
