package commands

import (
	"fmt"

	"github.com/zachmann/mytoken/internal/client/config"
)

// atCommand is a type for holding and handling the AT command
type atCommand struct {
	generalOptions
	Scopes    []string `long:"scope" description:"Request the passed scope. Can be used multiple times."`
	Audiences []string `long:"aud" description:"Request the passed audience. Can be used multiple times."`
}

// Execute implements the flags.Commander interface
func (atc *atCommand) Execute(args []string) error {
	var comment string
	if len(args) > 0 {
		comment = args[0]
	}
	mytoken := config.Get().Mytoken
	at, err := mytoken.GetAccessToken(atc.SuperToken, atc.Scopes, atc.Audiences, comment)
	if err != nil {
		return err
	}
	fmt.Println(at)
	return nil
}
