package main

import (
	"github.com/zachmann/mytoken/internal/client/commands"
	"github.com/zachmann/mytoken/internal/client/utils/logger"
	"github.com/zachmann/mytoken/internal/httpClient"
)

func main() {
	logger.Init()
	httpClient.Init("")
	commands.Parse()
}
