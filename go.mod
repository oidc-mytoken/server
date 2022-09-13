module github.com/oidc-mytoken/server

go 1.16

require (
	github.com/Songmu/prompter v0.5.1
	github.com/coreos/go-oidc/v3 v3.4.0
	github.com/fatih/structs v1.1.0
	github.com/gliderlabs/ssh v0.3.5
	github.com/go-resty/resty/v2 v2.7.0
	github.com/go-sql-driver/mysql v1.6.0
	github.com/gofiber/fiber/v2 v2.37.1
	github.com/gofiber/helmet/v2 v2.2.16
	github.com/gofiber/template v1.7.1
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/ip2location/ip2location-go v8.3.0+incompatible
	github.com/jinzhu/copier v0.3.5
	github.com/jmoiron/sqlx v1.3.5
	github.com/lestrrat-go/jwx v1.2.25
	github.com/oidc-mytoken/api v0.8.0
	github.com/oidc-mytoken/lib v0.5.0
	github.com/pires/go-proxyproto v0.6.2
	github.com/pkg/errors v0.9.1
	github.com/satori/go.uuid v1.2.0
	github.com/sirupsen/logrus v1.9.0
	github.com/urfave/cli/v2 v2.3.1-0.20211205195634-e8d81738896c
	github.com/valyala/fasthttp v1.40.0
	golang.org/x/crypto v0.0.0-20220826181053-bd7e27e6170d
	golang.org/x/mod v0.5.1
	golang.org/x/oauth2 v0.0.0-20220822191816-0ebed06d0094
	golang.org/x/term v0.0.0-20220722155259-a9ba230a4035
	gopkg.in/yaml.v3 v3.0.1
)

replace github.com/urfave/cli/v2 => github.com/zachmann/cli/v2 v2.3.1-0.20211220102037-d619fd40a704
