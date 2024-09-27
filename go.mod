module github.com/oidc-mytoken/server

go 1.22.0

toolchain go1.22.5

require (
	github.com/Songmu/prompter v0.5.1
	github.com/TwiN/gocache/v2 v2.2.2
	github.com/arran4/golang-ical v0.3.1
	github.com/coreos/go-oidc/v3 v3.11.0
	github.com/evanphx/json-patch/v5 v5.9.0
	github.com/fatih/structs v1.1.0
	github.com/gliderlabs/ssh v0.3.7
	github.com/go-resty/resty/v2 v2.15.3
	github.com/go-sql-driver/mysql v1.8.1
	github.com/gofiber/fiber/v2 v2.52.5
	github.com/gofiber/template/mustache/v2 v2.0.12
	github.com/gofrs/uuid v4.4.0+incompatible
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/ip2location/ip2location-go v8.3.0+incompatible
	github.com/jinzhu/copier v0.4.0
	github.com/jmoiron/sqlx v1.4.0
	github.com/jordan-wright/email v4.0.1-0.20210109023952-943e75fe5223+incompatible
	github.com/lestrrat-go/jwx v1.2.30
	github.com/oidc-mytoken/api v0.11.2-0.20240426092102-fa4d583a79ad
	github.com/oidc-mytoken/lib v0.7.1
	github.com/oidc-mytoken/utils v0.1.3-0.20230731143919-ea5b78243e5d
	github.com/pires/go-proxyproto v0.7.0
	github.com/pkg/errors v0.9.1
	github.com/redis/go-redis/v9 v9.6.1
	github.com/sethvargo/go-limiter v1.0.0
	github.com/sirupsen/logrus v1.9.3
	github.com/urfave/cli/v2 v2.3.1-0.20211205195634-e8d81738896c
	github.com/valyala/fasthttp v1.56.0
	github.com/vmihailenco/msgpack/v5 v5.4.1
	github.com/zachmann/go-oidfed v0.1.1-0.20240830095406-169de417a975
	golang.org/x/crypto v0.27.0
	golang.org/x/mod v0.21.0
	golang.org/x/oauth2 v0.23.0
	golang.org/x/term v0.24.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/andybalholm/brotli v1.1.0 // indirect
	github.com/anmitsu/go-shlex v0.0.0-20200514113438-38f4b401e2be // indirect
	github.com/cbroglie/mustache v1.4.0 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.1 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-jose/go-jose/v4 v4.0.2 // indirect
	github.com/goccy/go-json v0.10.3 // indirect
	github.com/gofiber/template v1.8.3 // indirect
	github.com/gofiber/utils v1.1.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/lestrrat-go/backoff/v2 v2.0.8 // indirect
	github.com/lestrrat-go/blackmagic v1.0.2 // indirect
	github.com/lestrrat-go/httpcc v1.0.1 // indirect
	github.com/lestrrat-go/iter v1.0.2 // indirect
	github.com/lestrrat-go/option v1.0.1 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/philhofer/fwd v1.1.2 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/tinylib/msgp v1.1.8 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	golang.org/x/exp v0.0.0-20220722155223-a9213eeb770e // indirect
	golang.org/x/net v0.29.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	tideland.dev/go/slices v0.2.0 // indirect
)

replace github.com/urfave/cli/v2 => github.com/zachmann/cli/v2 v2.3.1-0.20211220102037-d619fd40a704
