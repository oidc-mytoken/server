package commands

import (
	"fmt"

	"github.com/zachmann/mytoken/internal/client/config"
)

type revokeCommand struct {
	generalOptions
	Recursive bool `short:"r" long:"recursive" description:"If set also all subtokens are revoked"`
}

// Execute implements the flags.Commander interface
func (rc *revokeCommand) Execute(args []string) error {
	mytoken := config.Get().Mytoken
	provider, superToken := rc.Check()
	err := mytoken.Revoke(superToken, provider.Issuer, rc.Recursive)
	if err != nil {
		return err
	}
	fmt.Println("Token revoked")
	return nil
}
