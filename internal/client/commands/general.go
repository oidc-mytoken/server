package commands

// generalOptions holds command line options that can be used with all commands
type generalOptions struct {
	Provider   string `short:"p" long:"provider" description:"The name or issuer url of the OpenID provider that should be used."`
	Name       string `long:"name" description:"The name of the super token that should be used."`
	SuperToken string `long:"ST" description:"The passed super token is used instead of a stored one."`
}
