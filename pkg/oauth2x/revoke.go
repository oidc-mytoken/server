package oauth2x

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
)

// RevokeToken revokes a token at the revocation endpoint
// This function is untested
func (c *Config) RevokeToken(oauth2Config *oauth2.Config, token string) error {
	req, err := http.NewRequest("POST", c.endpoints.Revocation, strings.NewReader(fmt.Sprintf("token=%s", token)))
	if err != nil {
		return err
	}
	req.SetBasicAuth(oauth2Config.ClientID, oauth2Config.ClientSecret)
	resp, err := doRequest(c.Ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("unable to read response body: %v", err)
		}
		return fmt.Errorf("oauth2x: failed to revoke token: %s: %s", resp.Status, body)
	}
	return nil
}
