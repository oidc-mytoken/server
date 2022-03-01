module github.com/oidc-mytoken/server

go 1.16

require (
	github.com/Songmu/prompter v0.5.0
	github.com/coreos/go-oidc/v3 v3.1.0
	github.com/fatih/structs v1.1.0
	github.com/gliderlabs/ssh v0.3.3
	github.com/go-resty/resty/v2 v2.7.0
	github.com/go-sql-driver/mysql v1.6.0
	github.com/gofiber/fiber/v2 v2.28.0
	github.com/gofiber/helmet/v2 v2.2.6
	github.com/gofiber/template v1.6.22
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/ip2location/ip2location-go v8.3.0+incompatible
	github.com/jinzhu/copier v0.3.5
	github.com/jmoiron/sqlx v1.3.4
	github.com/klauspost/compress v1.14.2 // indirect
	github.com/lestrrat-go/jwx v1.2.18
	github.com/oidc-mytoken/api v0.5.1
	github.com/oidc-mytoken/lib v0.3.3
	github.com/pires/go-proxyproto v0.6.1
	github.com/pkg/errors v0.9.1
	github.com/satori/go.uuid v1.2.0
	github.com/sirupsen/logrus v1.8.1
	github.com/urfave/cli/v2 v2.3.1-0.20211205195634-e8d81738896c
	github.com/valyala/fasthttp v1.33.0
	golang.org/x/crypto v0.0.0-20220112180741-5e0467b6c7ce
	golang.org/x/mod v0.5.1
	golang.org/x/oauth2 v0.0.0-20211104180415-d3ed0bb246c8
	golang.org/x/sys v0.0.0-20220114195835-da31bd327af9 // indirect
	golang.org/x/term v0.0.0-20210503060354-a79de5458b56
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

replace github.com/urfave/cli/v2 => github.com/zachmann/cli/v2 v2.3.1-0.20211220102037-d619fd40a704
