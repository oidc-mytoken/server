package commands

import (
	"fmt"
	"os"

	"github.com/zachmann/mytoken/internal/client/config"
	"github.com/zachmann/mytoken/internal/client/utils/cryptutils"
	"github.com/zachmann/mytoken/internal/server/supertoken/capabilities"
	restrictions "github.com/zachmann/mytoken/internal/server/supertoken/restrictions"
	"github.com/zachmann/mytoken/pkg/model"
)

func init() {
	options.ST.CommonSTOptions = &CommonSTOptions{}
	options.ST.Store.CommonSTOptions = options.ST.CommonSTOptions
	st, _ := parser.AddCommand("ST", "Obtain super token", "Obtain a new mytoken super token", &options.ST)
	st.SubcommandsOptional = true
	for _, o := range st.Options() {
		if o.LongName == "capability" {
			o.Choices = capabilities.AllCapabilities.Strings()
		}
		if o.LongName == "subtoken-capability" {
			o.Choices = capabilities.AllCapabilities.Strings()
		}
	}
}

type stCommand struct {
	Store stStoreCommand `command:"store" description:"Store the obtained super token encrypted instead of returning it. This way the super token can be easily used with mytoken."`

	*CommonSTOptions

	TokenType string `long:"token-type" choice:"short" choice:"transfer" choice:"token" default:"token" description:"The type of the returned token. Can only be used if token is not stored."`
}

type CommonSTOptions struct {
	generalOptions
	TransferCode string `long:"TC" description:"Use the passed transfer code to exchange it into a super token"`
	OIDCFlow     string `long:"oidc" choice:"auth" choice:"device" choice:"default" optional:"true" optional-value:"default" description:"Use the passed OpenID Connect flow to create a super token"`

	Scopes               []string `long:"scope" description:"Request the passed scope. Can be used multiple times"`
	Audiences            []string `long:"aud" description:"Request the passed audience. Can be used multiple times"`
	Capabilities         []string `long:"capability" default:"default" description:"Request the passed capabilities. Can be used multiple times"` //TODO
	SubtokenCapabilities []string `long:"subtoken-capability" description:"Request the passed subtoken capabilities. Can be used multiple times"` //TODO
	Restrictions         string
}

type stStoreCommand struct {
	*CommonSTOptions
	PositionalArgs struct {
		StoreName string `positional-arg-name:"NAME" description:"Store the obtained super token under NAME. It can be used later by referencing NAME."`
	} `positional-args:"true" required:"true"`
	GPGKey   string `short:"k" long:"gpg-key" value-name:"KEY" description:"Use KEY for encryption instead of the default key"`
	Password bool   `long:"password" description:"Use a password for encrypting the token instead of a gpg key."`
}

// Execute implements the flags.Commander interface
func (stc *stCommand) Execute(args []string) error {
	if len(stc.Capabilities) > 0 && stc.Capabilities[0] == "default" {
		stc.Capabilities = config.Get().DefaultTokenCapabilities.Returned
	}
	st, err := obtainST(stc.CommonSTOptions, "", model.NewResponseType(stc.TokenType))
	if err != nil {
		return err
	}
	fmt.Println(st)
	return nil
}

func obtainST(args *CommonSTOptions, name string, responseType model.ResponseType) (string, error) {
	mytoken := config.Get().Mytoken
	if len(args.TransferCode) > 0 {
		return mytoken.GetSuperTokenByTransferCode(args.TransferCode)
	}
	provider, err := args.generalOptions.checkProvider()
	if err != nil {
		return "", err
	}
	tokenName := name
	prefix := config.Get().TokenNamePrefix
	if len(name) > 0 && len(prefix) > 0 {
		tokenName = fmt.Sprintf("%s:%s", prefix, name)
	}
	var r restrictions.Restrictions = nil
	c := capabilities.NewCapabilities(args.Capabilities)
	sc := capabilities.NewCapabilities(args.SubtokenCapabilities)
	if len(args.OIDCFlow) > 0 {
		if args.OIDCFlow == "default" {
			args.OIDCFlow = config.Get().DefaultOIDCFlow
		}
		switch args.OIDCFlow {
		case "auth":
			return mytoken.GetSuperTokenByAuthorizationFlow(provider.Issuer, r, c, sc, responseType, tokenName,
				func(authorizationURL string) error {
					fmt.Println(authorizationURL)
					return nil
				},
				func(interval int64, iteration int) {
					if iteration == 0 {
						fmt.Fprint(os.Stderr, "Starting polling ... ")
						return
					}
					if int64(iteration)%(30/interval) == 0 { // every 30s
						fmt.Fprint(os.Stderr, ".")
					}
				},
				func() {
					fmt.Fprintln(os.Stderr)
					fmt.Fprintln(os.Stderr, "success")
				},
			)
		case "device":
			return "", fmt.Errorf("Not yet implemented")
		default:
			return "", fmt.Errorf("Unknown oidc flow. Implementation error.")
		}
	}
	stGrant, err := args.generalOptions.checkToken(provider.Issuer)
	if err != nil {
		return "", err
	}
	return mytoken.GetSuperTokenBySuperToken(stGrant, provider.Issuer, r, c, sc, responseType, tokenName)
}

// Execute implements the flags.Commander interface
func (sstc *stStoreCommand) Execute(args []string) error {
	if len(sstc.Capabilities) > 0 && sstc.Capabilities[0] == "default" {
		sstc.Capabilities = config.Get().DefaultTokenCapabilities.Stored
	}
	provider, err := sstc.CommonSTOptions.generalOptions.checkProvider()
	if err != nil {
		return err
	}
	st, err := obtainST(sstc.CommonSTOptions, sstc.PositionalArgs.StoreName, model.ResponseTypeToken)
	if err != nil {
		return err
	}
	gpgKey := sstc.GPGKey
	if sstc.Password {
		gpgKey = ""
	} else {
		if len(gpgKey) == 0 {
			gpgKey = provider.GPGKey
		}
	}
	var encryptedToken string
	if len(gpgKey) == 0 {
		encryptedToken, err = cryptutils.EncryptPassword(st)
	} else {
		encryptedToken, err = cryptutils.EncryptGPG(st, gpgKey)
	}
	if err != nil {
		return err
	}
	saveEncryptedToken(encryptedToken, provider.Issuer, sstc.PositionalArgs.StoreName, gpgKey)
	return nil
}

func saveEncryptedToken(token, issuer, name, gpgKey string) error {
	tokens, err := config.LoadTokens()
	if err != nil {
		return err
	}
	for _, t := range tokens[issuer] {
		if t.Name == name {
			t.Token = token
			t.GPGKey = gpgKey
			return config.SaveTokens(tokens)
		}
	}
	tokens[issuer] = append(tokens[issuer], config.TokenEntry{
		Token:  token,
		Name:   name,
		GPGKey: gpgKey,
	})
	return config.SaveTokens(tokens)
}
