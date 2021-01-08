package main

import (
	"fmt"
	"os"

	"github.com/zachmann/mytoken/internal/client/config"
	"github.com/zachmann/mytoken/internal/httpClient"
	"github.com/zachmann/mytoken/pkg/mytokenlib"
)

func main() {
	config.Init()
	httpClient.Init("")

	mytoken, err := mytokenlib.NewMytokenInstance(config.Get().URL)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	st := os.Getenv("ST_AT")
	at, err := mytoken.GetAccessToken(st, nil, nil, "testAT")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(at)
}
