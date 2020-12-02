package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	log "github.com/sirupsen/logrus"
	"github.com/zachmann/mytoken/internal/context"
	"github.com/zachmann/mytoken/internal/model"
	"github.com/zachmann/mytoken/internal/utils/fileutil"
	"github.com/zachmann/mytoken/internal/utils/issuerUtils"
	"github.com/zachmann/mytoken/pkg/oauth2x"
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
	EnabledResponseTypes                []model.ResponseType     `yaml:"enabled_response_types"`
	Signing                             signingConf              `yaml:"signing"`
	ServiceDocumentation                string                   `yaml:"service_documentation"`
	Polling                             pollingConf              `yaml:"polling_codes"`
	Logging                             loggingConf              `yaml:"logging"`
}

type loggingConf struct {
	Access   LoggerConf `yaml:"access"`
	Internal LoggerConf `yaml:"internal"`
}

type LoggerConf struct {
	Dir    string `yaml:"dir"`
	StdErr bool   `yaml:"stderr"`
	Level  string `yaml:"level"`
}

type pollingConf struct {
	PollingCodeExpiresAfter int64 `yaml:"expires_after"`
	PollingInterval         int64 `yaml:"polling_interval"`
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
	Alg       string `yaml:"alg"`
	KeyFile   string `yaml:"key_file"`
	RSAKeyLen int    `yaml:"rsa_key_len"`
}

// ProviderConf holds information about a provider
type ProviderConf struct {
	Issuer       string             `yaml:"issuer"`
	ClientID     string             `yaml:"client_id"`
	ClientSecret string             `yaml:"client_secret"`
	Scopes       []string           `yaml:"scopes"`
	Endpoints    *oauth2x.Endpoints `yaml:"-"`
	Provider     *oidc.Provider     `yaml:"-"`
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
		var err error
		p.Endpoints, err = oauth2x.NewConfig(context.Get(), p.Issuer).Endpoints()
		if err != nil {
			return fmt.Errorf("Error '%s' for provider.issuer '%s' (Index %d)", err, p.Issuer, i)
		}
		p.Provider, err = oidc.NewProvider(context.Get(), p.Issuer)
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
	if conf.Polling.PollingInterval == 0 {
		conf.Polling.PollingInterval = 5
	}
	if conf.Polling.PollingCodeExpiresAfter == 0 {
		conf.Polling.PollingCodeExpiresAfter = 300
	}
	model.OIDCFlowAuthorizationCode.AddToSliceIfNotFound(conf.EnabledOIDCFlows)
	model.GrantTypeOIDCFlow.AddToSliceIfNotFound(conf.EnabledSuperTokenEndpointGrantTypes)
	model.GrantTypeSuperToken.AddToSliceIfNotFound(conf.EnabledSuperTokenEndpointGrantTypes)
	model.ResponseTypeToken.AddToSliceIfNotFound(conf.EnabledResponseTypes)
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
		log.WithField("filepath", filep).Debug("Looking for config file")
		if fileutil.FileExists(filep) {
			return fileutil.MustReadFile(filep)
		}
	}
	log.WithField("filepath", filename).Fatal("Could not find config file in any of the possible directories")
	return nil
}

// Load reads the config file and populates the config struct; then validates the config
func Load() {
	load()
	if err := validate(); err != nil {
		log.WithError(err).Fatal()
	}
}

func load() {
	data := readConfigFile("config.yaml")
	conf = newConfig()
	err := yaml.Unmarshal(data, conf)
	if err != nil {
		log.WithError(err).Fatal()
		return
	}
}

// LoadForSetup reads the config file and populates the config struct; it does not validate the config, since this is not required for setup
func LoadForSetup() {
	load()
}

func newConfig() *Config {
	return &Config{
		ProviderByIssuer: make(map[string]*ProviderConf),
	}
}
