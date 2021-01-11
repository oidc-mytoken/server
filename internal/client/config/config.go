package config

import (
	"fmt"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/zachmann/mytoken/internal/client/model"
	"github.com/zachmann/mytoken/internal/client/utils/cryptutils"
	"github.com/zachmann/mytoken/internal/server/utils/issuerUtils"
	"github.com/zachmann/mytoken/internal/utils/fileutil"
	"github.com/zachmann/mytoken/pkg/mytokenlib"
)

type config struct {
	URL     string              `yaml:"instance"`
	Mytoken *mytokenlib.Mytoken `yaml:"-"`

	DefaultGPGKey   string `yaml:"default_gpg_key"`
	DefaultProvider string `yaml:"default_provider"`

	Providers model.Providers `yaml:"provider_names"`
	Tokens    []struct {
		Provider     string `yaml:"provider"`
		DefaultToken string `yaml:"default_token"`
		Tokens       []struct {
			Token  string `yaml:"token"`
			Name   string `yaml:"name"`
			GPGKey string `yaml:"gpg_key"`
		} `yaml:"tokens"`
	} `yaml:"tokens"`
}

func (c *config) GetToken(issuer, name string) (string, error) {
	for _, tt := range c.Tokens {
		if issuerUtils.CompareIssuerURLs(tt.Provider, issuer) {
			if len(name) == 0 {
				name = tt.DefaultToken
			}
			for _, t := range tt.Tokens {
				if t.Name == name {
					token, err := cryptutils.DecryptGPG(t.Token, t.GPGKey)
					if err != nil {
						err = fmt.Errorf("Failed to decrypt token named '%s' for '%s'", name, issuer)
					}
					return token, err
				}
			}
			return "", fmt.Errorf("Token name '%s' not found for '%s'", name, issuer)
		}
	}
	return "", fmt.Errorf("Provider '%s' not found", issuer)
}

var defaultConfig = config{}

var conf *config

// Get returns the config
func Get() *config {
	return conf
}

func load(name string, locations []string) {
	data := fileutil.ReadConfigFile(name, locations)
	conf = &defaultConfig
	err := yaml.Unmarshal(data, conf)
	if err != nil {
		log.Fatal(err)
	}
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
