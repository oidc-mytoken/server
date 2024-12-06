package main

import (
	server "github.com/oidc-mytoken/server/internal/notifier/server"
)

func main() {
	loadConfig()
	server.InitStandalone(conf.Email)
}
