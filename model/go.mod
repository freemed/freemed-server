module github.com/freemed/freemed-server/model

go 1.25.0

replace (
	github.com/freemed/freemed-server => ../
	github.com/freemed/freemed-server/common => ../common
	github.com/freemed/freemed-server/config => ../config
	github.com/freemed/remitt-server/client => ../../remitt-server/client
	github.com/freemed/remitt-server/common => ../../remitt-server/common
	github.com/freemed/remitt-server/config => ../../remitt-server/config
	github.com/freemed/remitt-server/model => ../../remitt-server/model
)

require (
	github.com/freemed/freemed-server v0.0.0-00010101000000-000000000000
	github.com/freemed/freemed-server/common v0.0.0-20250416130701-c7ae75b9c375
	golang.org/x/crypto v0.37.0
)

require (
	github.com/appleboy/gin-jwt/v2 v2.10.3 // indirect
	github.com/bytedance/sonic v1.13.2 // indirect
	github.com/bytedance/sonic/loader v0.2.4 // indirect
	github.com/cloudwego/base64x v0.1.5 // indirect
	github.com/gabriel-vasile/mimetype v1.4.9 // indirect
	github.com/gin-contrib/sse v1.1.0 // indirect
	github.com/gin-gonic/gin v1.10.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.26.0 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.2 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/cpuid/v2 v2.2.10 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/nxadm/tail v1.4.8 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.12 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	golang.org/x/arch v0.16.0 // indirect
	golang.org/x/net v0.39.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.24.0 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
	gopkg.in/bsm/ratelimit.v1 v1.0.0-20170922094635-f56db5e73a5e // indirect
	gopkg.in/redis.v3 v3.6.4 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
