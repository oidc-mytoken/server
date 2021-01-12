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
	ST            stCommand
	AT            atCommand
	Revoke        revokeCommand
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
