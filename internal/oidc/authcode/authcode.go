package authcode

import (
	"fmt"
	"log"
	"time"

	"github.com/zachmann/mytoken/internal/utils/issuerUtils"

	"github.com/zachmann/mytoken/internal/config"
	"github.com/zachmann/mytoken/internal/db/dbModels"
	response "github.com/zachmann/mytoken/internal/endpoints/token/super/pkg"
	"github.com/zachmann/mytoken/internal/oidc/issuer"
	"github.com/zachmann/mytoken/internal/utils"
	"golang.org/x/oauth2"
)

var redirectURL string

func Init() {
	redirectURL = utils.CombineURLPath(config.Get().IssuerURL, "/redirect")
}

const stateLen = 16
const pollingCodeLen = 16

type stateInfo struct {
	Native bool
}

const stateFmt = "%d%s"

func createState(info stateInfo) string {
	r := utils.RandASCIIString(stateLen)
	native := 0
	if info.Native {
		native = 1
	}
	return fmt.Sprintf(stateFmt, native, r)
}

func parseState(state string) stateInfo {
	info := stateInfo{}
	native := 0
	var r string
	fmt.Scanf(stateFmt, native, r)
	if native != 0 {
		info.Native = true
	}
	return info
}

func authorizationURL(provider *config.ProviderConf, native bool) (string, string) {
	log.Printf("Generating authorization url")
	oauth2Config := oauth2.Config{
		ClientID:     provider.ClientID,
		ClientSecret: provider.ClientSecret,
		Endpoint:     provider.Provider.Endpoint(),
		RedirectURL:  redirectURL,
		Scopes:       provider.Scopes, //TODO use restrictions
	}
	state := createState(stateInfo{Native: native})
	additionalParams := []oauth2.AuthCodeOption{oauth2.ApprovalForce}
	if issuerUtils.CompareIssuerURLs(provider.Issuer, issuer.GOOGLE) {
		additionalParams = append(additionalParams, oauth2.AccessTypeOffline)
	}
	//TODO add audience from restriction

	return oauth2Config.AuthCodeURL(state, additionalParams...), state
}

func InitAuthCodeFlow(provider *config.ProviderConf, req *response.AuthCodeFlowRequest) (res response.AuthCodeFlowResponse, err error) {
	log.Print("Handle authcode")
	authURL, state := authorizationURL(provider, req.Native())
	authFlowInfo := dbModels.AuthFlowInfo{
		State:        state,
		Issuer:       provider.Issuer,
		Restrictions: req.Restrictions,
		Capabilities: req.Capabilities,
		Name:         req.Name,
	}
	res.AuthorizationURL = authURL
	if req.Native() {
		authFlowInfo.PollingCode = utils.RandASCIIString(pollingCodeLen)
		res.PollingCode = authFlowInfo.PollingCode
		res.PollingCodeExpires = time.Now().Add(time.Duration(config.Get().PollingCodeExpiresAfter) * time.Second)
	}
	if e := authFlowInfo.Store(); e != nil {
		err = e
		return
	}
	return
}
