module github.com/freemed/freemed-server/cmd/freemed-server

go 1.25.0

replace (
	github.com/freemed/freemed-server => ../../
	github.com/freemed/freemed-server/api => ../../api
	github.com/freemed/freemed-server/common => ../../common
	github.com/freemed/freemed-server/config => ../../config
	github.com/freemed/freemed-server/model => ../../model
	github.com/freemed/ratago/xslt => ../../../ratago/xslt
	github.com/freemed/remitt-server => ../../../remitt-server
	github.com/freemed/remitt-server/client => ../../../remitt-server/client
	github.com/freemed/remitt-server/common => ../../../remitt-server/common
	github.com/freemed/remitt-server/config => ../../../remitt-server/config
	github.com/freemed/remitt-server/model => ../../../remitt-server/model

	github.com/ugorji/go => github.com/ugorji/go/codec v1.1.7
)

require (
	github.com/appleboy/gin-jwt/v2 v2.10.3
	github.com/freemed/freemed-server v0.0.0-00010101000000-000000000000
	github.com/freemed/freemed-server/api v0.0.0-20240506234320-3b301527e988
	github.com/freemed/freemed-server/common v0.0.0-20250416130701-c7ae75b9c375
	github.com/freemed/freemed-server/config v0.0.0-20260815062657-6692d8561a7b
	github.com/freemed/freemed-server/model v0.0.0-20250416130701-c7ae75b9c375
	github.com/gin-gonic/contrib v0.0.0-20221130124618-7e01895a63f2
	github.com/gin-gonic/gin v1.12.0
	github.com/google/uuid v1.6.0
	golang.org/x/crypto v0.55.0
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
)

require (
	filippo.io/edwards25519 v1.2.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bytedance/gopkg v0.1.4 // indirect
	github.com/bytedance/sonic v1.15.2 // indirect
	github.com/bytedance/sonic/loader v0.5.2 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cloudwego/base64x v0.1.7 // indirect
	github.com/gabriel-vasile/mimetype v1.4.15 // indirect
	github.com/gin-contrib/sse v1.1.1 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.30.3 // indirect
	github.com/go-sql-driver/mysql v1.10.0 // indirect
	github.com/goccy/go-json v0.10.6 // indirect
	github.com/goccy/go-yaml v1.19.2 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.2 // indirect
	github.com/golang-migrate/migrate/v4 v4.19.1 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/jung-kurt/gofpdf v1.16.2 // indirect
	github.com/klauspost/cpuid/v2 v2.4.0 // indirect
	github.com/leodido/go-urn v1.5.0 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pelletier/go-toml/v2 v2.4.3 // indirect
	github.com/prometheus/client_golang v1.24.1 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.70.1 // indirect
	github.com/prometheus/procfs v0.21.1 // indirect
	github.com/quic-go/qpack v0.6.0 // indirect
	github.com/quic-go/quic-go v0.61.0 // indirect
	github.com/richardlehane/mscfb v1.0.7 // indirect
	github.com/richardlehane/msoleps v1.0.6 // indirect
	github.com/tiendc/go-deepcopy v1.7.2 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.3.2 // indirect
	github.com/xuri/efp v0.0.1 // indirect
	github.com/xuri/excelize/v2 v2.11.0 // indirect
	github.com/xuri/nfp v0.0.2-0.20250530014748-2ddeb826f9a9 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	go.mongodb.org/mongo-driver/v2 v2.8.0 // indirect
	golang.org/x/arch v0.30.0 // indirect
	golang.org/x/net v0.58.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.41.0 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
	gopkg.in/bsm/ratelimit.v1 v1.0.0-20170922094635-f56db5e73a5e // indirect
	gopkg.in/redis.v3 v3.6.4 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
