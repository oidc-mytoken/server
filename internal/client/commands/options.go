package commands

import (
	"os"

	"github.com/jessevdk/go-flags"

	"github.com/zachmann/mytoken/internal/client/config"
)

type configOptions struct {
	Config func(filename flags.Filename) `long:"config" value-name:"FILE" default:"" description:"Use FILE as the config file instead of the default one."`
}

// options holds all the command line commands and their options
var options struct {
	ConfigOptions configOptions
	AT            atCommand
	Revoke        revokeCommand

	ST struct {
		TransferCode string `long:"TC" description:"Use the passed transfer code to exchange it into a super token"`
		OIDCFlow     string `long:"oidc" choice:"auth" choice:"device" optional:"true" optional-value:"auth" description:"Use the passed OpenID Connect flow to create a super token"`

		Scopes               []string `long:"scope" description:"Request the passed scope. Can be used multiple times"`
		Audiences            []string `long:"aud" description:"Request the passed audience. Can be used multiple times"`
		Capabilities         []string `long:"capability" choice:"AT" choice:"create_supertoken" description:"Request the passed capabilities. Can be used multiple times"`                   //TODO
		SubtokenCapabilities []string `long:"subtoken-capability" choice:"AT" choice:"create_supertoken" description:"Request the passed subtoken capabilities. Can be used multiple times"` //TODO
		Restrictions         string

		TokenType  string `long:"token-type" choice:"short" choice:"transfer" choice:"token" default:"token" description:"The type of the returned token. Can only be used if --store is not set"`
		StoreToken bool   `short:"s" long:"store" description:"If set the super token is stored by the mytoken client so it can be used later"`
		GPGKey     string `long:"gpg-key" value-name:"KEY" description:"When storing this token use KEY for encryption"`
	} `command:"ST" description:"Obtain a new mytoken super token."`
}

var parser *flags.Parser

func init() {
	options.ConfigOptions.Config = func(filename flags.Filename) {
		if len(filename) > 0 {
			config.Load(string(filename))
		} else {
			config.LoadDefault()
		}
	}

	parser = flags.NewNamedParser("mytoken", flags.Default)
	parser.AddGroup("Config Options", "", &options.ConfigOptions)
	parser.AddCommand("AT", "Obtain access token", "Obtain a new OpenID Connect access token", &options.AT)
	parser.AddCommand("revoke", "Revoke super token", "Revoke a mytoken super token", &options.Revoke)
}

// Parse parses the command line options and calls the specified command
func Parse() {
	_, err := parser.Parse()
	if err != nil {
		os.Exit(1)
	}
}
