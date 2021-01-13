package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/zachmann/mytoken/internal/client/model"
	"github.com/zachmann/mytoken/internal/client/utils/cryptutils"
	"github.com/zachmann/mytoken/internal/utils/fileutil"
	"github.com/zachmann/mytoken/pkg/mytokenlib"
)

type config struct {
	URL     string              `yaml:"instance"`
	Mytoken *mytokenlib.Mytoken `yaml:"-"`

	DefaultGPGKey   string `yaml:"default_gpg_key"`
	DefaultProvider string `yaml:"default_provider"`
	DefaultOIDCFlow string `yaml:"default_oidc_flow"`

	TokenNamePrefix string `yaml:"token_name_prefix"`

	Providers  model.Providers         `yaml:"providers"`
	TokensFile string                  `yaml:"tokens_file"`
	Tokens     map[string][]TokenEntry `yaml:"-"`

	usedConfigDir string
}

type TokenEntry struct {
	Name   string `json:"name"`
	GPGKey string `json:"gpg_key,omitempty"`
	Token  string `json:"token"`
}

func (c *config) GetToken(issuer, name string) (string, error) {
	tt, found := c.Tokens[issuer]
	if !found {
		return "", fmt.Errorf("No tokens found for provider '%s'", issuer)
	}
	if len(name) == 0 {
		p, _ := c.Providers.FindBy(issuer, true)
		name = p.DefaultToken
	}
	for _, t := range tt {
		if t.Name == name {
			var token string
			var err error
			if len(t.GPGKey) > 0 {
				token, err = cryptutils.DecryptGPG(t.Token, t.GPGKey)
			} else {
				token, err = cryptutils.DecryptPassword(t.Token)
			}
			if err != nil {
				err = fmt.Errorf("Failed to decrypt token named '%s' for '%s'", name, issuer)
			}
			return token, err
		}
	}
	return "", fmt.Errorf("Token name '%s' not found for '%s'", name, issuer)
}

var defaultConfig = config{
	TokensFile:      "tokens.json",
	TokenNamePrefix: "<hostname>",
	DefaultOIDCFlow: "auth",
}

var conf *config

// Get returns the config
func Get() *config {
	return conf
}

func getTokensFilePath() string {
	filename := conf.TokensFile
	if filepath.IsAbs(filename) {
		return filename
	}
	return filepath.Join(conf.usedConfigDir, filename)
}

func SaveTokens(tokens map[string][]TokenEntry) error {
	data, err := json.MarshalIndent(tokens, "", "  ")
	if err != nil {
		return err
	}
	if err = ioutil.WriteFile(getTokensFilePath(), data, 0600); err != nil {
		return err
	}
	conf.Tokens = tokens
	return nil
}

func LoadTokens() (map[string][]TokenEntry, error) {
	tokens := make(map[string][]TokenEntry)
	data, err := ioutil.ReadFile(getTokensFilePath())
	if err != nil {
		return tokens, err
	}
	err = json.Unmarshal(data, &tokens)
	return tokens, err
}

func load(name string, locations []string) {
	data, usedLocation := fileutil.ReadConfigFile(name, locations)
	conf = &defaultConfig
	if err := yaml.Unmarshal(data, conf); err != nil {
		log.Fatal(err)
	}
	conf.usedConfigDir = usedLocation
	if len(conf.URL) == 0 {
		log.Fatal("Must provide url of the mytoken instance in the config file.")
	}
	mytoken, err := mytokenlib.NewMytokenInstance(conf.URL)
	if err != nil {
		log.Fatal(err)
	}
	conf.Mytoken = mytoken

	if len(conf.DefaultGPGKey) > 0 {
		for _, p := range conf.Providers {
			if len(p.GPGKey) == 0 {
				p.GPGKey = conf.DefaultGPGKey
			}
		}
	}
	conf.Tokens, err = LoadTokens()
	if err != nil {
		log.Fatal(err)
	}
	hostname, _ := os.Hostname()
	conf.TokenNamePrefix = strings.ReplaceAll(conf.TokenNamePrefix, "<hostname>", hostname)
}

// LoadDefault loads the config from one of the default config locations
func LoadDefault() {
	load("config.yaml", possibleConfigLocations)
}

// Load loads the config form the provided filepath
func Load(file string) {
	filename := filepath.Base(file)
	path := filepath.Dir(file)
	load(filename, []string{path})
}

var possibleConfigLocations = []string{
	"~/.config/mytoken",
	"~/.mytoken",
}
