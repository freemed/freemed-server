module github.com/freemed/freemed-server/api

go 1.24.0

replace (
	github.com/freemed/freemed-server => ../
	github.com/freemed/freemed-server/common => ../common
	github.com/freemed/freemed-server/config => ../config
	github.com/freemed/freemed-server/model => ../model
	github.com/freemed/gokogiri => ../../gokogiri
	github.com/freemed/gokogiri/help => ../../gokogiri/help
	github.com/freemed/gokogiri/xpath => ../../gokogiri/xpath
	github.com/freemed/ratago/xslt => ../../ratago/xslt
	github.com/freemed/remitt-server/common => ../../remitt-server/common
	github.com/freemed/remitt-server/config => ../../remitt-server/config
	github.com/freemed/remitt-server/model => ../../remitt-server/model
)

require (
	github.com/freemed/freemed-server/common v0.0.0-20250416130701-c7ae75b9c375
	github.com/freemed/freemed-server/config v0.0.0-20250416130701-c7ae75b9c375
	github.com/freemed/freemed-server/model v0.0.0-20250416130701-c7ae75b9c375
	github.com/gin-gonic/gin v1.10.0
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/appleboy/gin-jwt/v2 v2.10.3 // indirect
	github.com/bytedance/sonic v1.13.2 // indirect
	github.com/bytedance/sonic/loader v0.2.4 // indirect
	github.com/cloudwego/base64x v0.1.5 // indirect
	github.com/freemed/gokogiri/help v0.0.0-20250402180648-1e651eb8ffcd // indirect
	github.com/freemed/gokogiri/util v0.0.0-20250402180648-1e651eb8ffcd // indirect
	github.com/freemed/gokogiri/xml v0.0.0-20250402180648-1e651eb8ffcd // indirect
	github.com/freemed/gokogiri/xpath v0.0.0-20250402180648-1e651eb8ffcd // indirect
	github.com/freemed/ratago/xslt v0.0.0-20250203231425-016f1ea48158 // indirect
	github.com/freemed/remitt-server/common v0.0.0-20250408203712-6db141480e4c // indirect
	github.com/freemed/remitt-server/config v0.0.0-20250408203712-6db141480e4c // indirect
	github.com/freemed/remitt-server/model v0.0.0-20250408203712-6db141480e4c // indirect
	github.com/gabriel-vasile/mimetype v1.4.9 // indirect
	github.com/gin-contrib/sse v1.1.0 // indirect
	github.com/go-gorp/gorp v2.2.0+incompatible // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.26.0 // indirect
	github.com/go-sql-driver/mysql v1.9.2 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/cpuid/v2 v2.2.10 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/mattes/migrate v3.0.1+incompatible // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.12 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	golang.org/x/arch v0.16.0 // indirect
	golang.org/x/crypto v0.37.0 // indirect
	golang.org/x/net v0.39.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/text v0.24.0 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/bsm/ratelimit.v1 v1.0.0-20170922094635-f56db5e73a5e // indirect
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df // indirect
	gopkg.in/redis.v3 v3.6.4 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gorm.io/driver/mysql v1.5.7 // indirect
	gorm.io/gorm v1.26.0 // indirect
)
