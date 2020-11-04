package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/zachmann/mytoken/internal/utils/issuerUtils"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/zachmann/mytoken/internal/model"
	"github.com/zachmann/mytoken/internal/utils/fileutil"
	"gopkg.in/yaml.v3"
)

// Config holds the server configuration
type Config struct {
	DB                                  dbConf                   `yaml:"database"`
	Server                              serverConf               `yaml:"server"`
	Providers                           []*ProviderConf          `yaml:"providers"`
	ProviderByIssuer                    map[string]*ProviderConf `yaml:"-"`
	IssuerURL                           string                   `yaml:"issuer"`
	EnabledOIDCFlows                    []model.OIDCFlow         `yaml:"enabled_oidc_flows"`
	EnabledSuperTokenEndpointGrantTypes []model.GrantType        `yaml:"enabled_super_token_endpoint_grant_types"`
	Signing                             signingConf              `yaml:"signing"`
	ServiceDocumentation                string                   `yaml:"service_documentation"`
	PollingCodeExpiresAfter             int64                    `yaml:"polling_code_expires_after"`
}

type dbConf struct {
	Host     string `yaml:"host"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DB       string `yaml:"db"`
}

type serverConf struct {
	Hostname string `yaml:"hostname"`
}

type signingConf struct {
	Alg     string `yaml:"alg"`
	KeyFile string `yaml:"key_file"`
}

// ProviderConf holds information about a provider
type ProviderConf struct {
	Issuer       string         `yaml:"issuer"`
	ClientID     string         `yaml:"client_id"`
	ClientSecret string         `yaml:"client_secret"`
	Scopes       []string       `yaml:"scopes"`
	Provider     *oidc.Provider `yaml:"-"`
}

var conf *Config

// Get returns the config
func Get() *Config {
	return conf
}

func validate() error {
	if conf == nil {
		return fmt.Errorf("config not set")
	}
	if conf.Server.Hostname == "" {
		return fmt.Errorf("invalid config: server.hostname not set")
	}
	if len(conf.Providers) <= 0 {
		return fmt.Errorf("invalid config: providers must have at least one entry")
	}
	for i, p := range conf.Providers {
		if p.Issuer == "" {
			return fmt.Errorf("invalid config: provider.issuer not set (Index %d)", i)
		}
		ctx := context.Background()
		var err error
		p.Provider, err = oidc.NewProvider(ctx, p.Issuer)
		if err != nil {
			return fmt.Errorf("Error '%s' for provider.issuer '%s' (Index %d)", err, p.Issuer, i)
		}
		if p.ClientID == "" {
			return fmt.Errorf("invalid config: provider.clientid not set (Index %d)", i)
		}
		if p.ClientSecret == "" {
			return fmt.Errorf("invalid config: provider.clientsecret not set (Index %d)", i)
		}
		if len(p.Scopes) <= 0 {
			return fmt.Errorf("invalid config: provider.scopes not set (Index %d)", i)
		}
		iss0, iss1 := issuerUtils.GetIssuerWithAndWithoutSlash(p.Issuer)
		conf.ProviderByIssuer[iss0] = p
		conf.ProviderByIssuer[iss1] = p
	}
	if conf.IssuerURL == "" {
		return fmt.Errorf("invalid config: issuerurl not set")
	}
	if conf.Signing.KeyFile == "" {
		return fmt.Errorf("invalid config: signingkeyfile not set")
	}
	if conf.Signing.Alg == "" {
		return fmt.Errorf("invalid config: tokensigningalg not set")
	}
	model.OIDCFlowAuthorizationCode.AddToSliceIfNotFound(conf.EnabledOIDCFlows)
	model.GrantTypeOIDCFlow.AddToSliceIfNotFound(conf.EnabledSuperTokenEndpointGrantTypes)
	model.GrantTypeSuperToken.AddToSliceIfNotFound(conf.EnabledSuperTokenEndpointGrantTypes)
	return nil
}

var possibleConfigLocations = []string{
	"config",
	"/etc/mytoken",
}

// readConfigFile checks if a file exists in one of the configuration
// directories and returns the content. If no file is found, mytoken exists.
func readConfigFile(filename string) []byte {
	for _, dir := range possibleConfigLocations {
		filep := filepath.Join(dir, filename)
		if strings.HasPrefix(filep, "~") {
			homeDir := os.Getenv("HOME")
			filep = filepath.Join(homeDir, filep[1:])
		}
		log.Printf("Looking for config file at: %s\n", filep)
		if fileutil.FileExists(filep) {
			return fileutil.MustReadFile(filep)
		}
	}
	fmt.Printf("Could not find config file %s in any of the possible directories\n", filename)
	os.Exit(1)
	return nil
}

// Load reads the config file and populates the config struct; then validates the config
func Load() {
	data := readConfigFile("config.yaml")
	conf = newConfig()
	err := yaml.Unmarshal(data, conf)
	if err != nil {
		log.Fatal(err)
		return
	}
	if err := validate(); err != nil {
		log.Fatal(err)
	}
}

func newConfig() *Config {
	return &Config{
		ProviderByIssuer: make(map[string]*ProviderConf),
	}
}
