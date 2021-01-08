package main

import (
	"github.com/zachmann/mytoken/internal/httpClient"
	"github.com/zachmann/mytoken/internal/server/config"
	"github.com/zachmann/mytoken/internal/server/db"
	configurationEndpoint "github.com/zachmann/mytoken/internal/server/endpoints/configuration"
	"github.com/zachmann/mytoken/internal/server/jws"
	"github.com/zachmann/mytoken/internal/server/oidc/authcode"
	server2 "github.com/zachmann/mytoken/internal/server/server"
	loggerUtils "github.com/zachmann/mytoken/internal/server/utils/logger"
)

func main() {
	config.Load()
	loggerUtils.Init()
	server2.Init()
	configurationEndpoint.Init()
	authcode.Init()
	if err := db.Connect(); err != nil {
		panic(err)
	}
	jws.LoadKey()
	httpClient.Init(config.Get().IssuerURL)

	server2.Start()
}
