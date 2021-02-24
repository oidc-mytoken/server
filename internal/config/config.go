package config

import (
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/oidc-mytoken/server/pkg/model"
	"github.com/oidc-mytoken/server/pkg/oauth2x"
	"github.com/oidc-mytoken/server/shared/context"
	"github.com/oidc-mytoken/server/shared/utils/fileutil"
	"github.com/oidc-mytoken/server/shared/utils/issuerUtils"
)

var defaultConfig = Config{
	Server: serverConf{
		Port: 8000,
		TLS: tlsConf{
			Enabled:      true, // The default is that TLS is enabled if cert and key are given, this is checked later; we must set true here, because otherwise we cannot distinct this from a false set by the user
			RedirectHTTP: true,
		},
	},
	DB: DBConf{
		Hosts:             []string{"localhost"},
		User:              "mytoken",
		Password:          "mytoken",
		DB:                "mytoken",
		ReconnectInterval: 60,
	},
	Signing: signingConf{
		Alg:       oidc.ES512,
		RSAKeyLen: 2048,
	},
	Logging: loggingConf{
		Access: LoggerConf{
			Dir:    "/var/log/mytoken",
			StdErr: false,
		},
		Internal: LoggerConf{
			Dir:    "/var/log/mytoken",
			StdErr: false,
			Level:  "error",
		},
	},
	ServiceDocumentation: "https://docs-sdm.scc.kit.edu/mytoken/",
	Features: featuresConf{
		EnabledOIDCFlows: []model.OIDCFlow{
			model.OIDCFlowAuthorizationCode,
		},
		TokenRevocation: onlyEnable{true},
		ShortTokens: shortTokenConfig{
			Enabled: true,
			Len:     64,
		},
		TransferCodes: onlyEnable{true},
		Polling: pollingConf{
			Enabled:                 true,
			Len:                     8,
			PollingCodeExpiresAfter: 300,
			PollingInterval:         5,
		},
		AccessTokenGrant: onlyEnable{true},
		SignedJWTGrant:   onlyEnable{true},
	},
	ProviderByIssuer: make(map[string]*ProviderConf),
	API: apiConf{
		MinVersion: 0,
	},
}

// Config holds the server configuration
type Config struct {
	IssuerURL            string                   `yaml:"issuer"`
	Server               serverConf               `yaml:"server"`
	GeoIPDBFile          string                   `yaml:"geo_ip_db_file"`
	API                  apiConf                  `yaml:"api"`
	DB                   DBConf                   `yaml:"database"`
	Signing              signingConf              `yaml:"signing"`
	Logging              loggingConf              `yaml:"logging"`
	ServiceDocumentation string                   `yaml:"service_documentation"`
	Features             featuresConf             `yaml:"features"`
	Providers            []*ProviderConf          `yaml:"providers"`
	ProviderByIssuer     map[string]*ProviderConf `yaml:"-"`
}

type apiConf struct {
	MinVersion int `yaml:"min_supported_version"`
}

type featuresConf struct {
	EnabledOIDCFlows []model.OIDCFlow `yaml:"enabled_oidc_flows"`
	TokenRevocation  onlyEnable       `yaml:"token_revocation"`
	ShortTokens      shortTokenConfig `yaml:"short_tokens"`
	TransferCodes    onlyEnable       `yaml:"transfer_codes"`
	Polling          pollingConf      `yaml:"polling_codes"`
	AccessTokenGrant onlyEnable       `yaml:"access_token_grant"`
	SignedJWTGrant   onlyEnable       `yaml:"signed_jwt_grant"`
}

type shortTokenConfig struct {
	Enabled bool `yaml:"enabled"`
	Len     int  `yaml:"len"`
}

type onlyEnable struct {
	Enabled bool `yaml:"enabled"`
}

type loggingConf struct {
	Access   LoggerConf `yaml:"access"`
	Internal LoggerConf `yaml:"internal"`
}

// LoggerConf holds configuration related to logging
type LoggerConf struct {
	Dir    string `yaml:"dir"`
	StdErr bool   `yaml:"stderr"`
	Level  string `yaml:"level"`
}

type pollingConf struct {
	Enabled                 bool  `yaml:"enabled"`
	Len                     int   `yaml:"len"`
	PollingCodeExpiresAfter int64 `yaml:"expires_after"`
	PollingInterval         int64 `yaml:"polling_interval"`
}

type DBConf struct {
	Hosts             []string `yaml:"hosts"`
	User              string   `yaml:"user"`
	Password          string   `yaml:"password"`
	DB                string   `yaml:"db"`
	ReconnectInterval int64    `yaml:"try_reconnect_interval"`
}

type serverConf struct {
	Hostname string  `yaml:"hostname"`
	Port     int     `yaml:"port"`
	TLS      tlsConf `yaml:"tls"`
}

type tlsConf struct {
	Enabled      bool   `yaml:"enabled"`
	RedirectHTTP bool   `yaml:"redirect_http"`
	Cert         string `yaml:"cert"`
	Key          string `yaml:"key"`
}

type signingConf struct {
	Alg       string `yaml:"alg"`
	KeyFile   string `yaml:"key_file"`
	RSAKeyLen int    `yaml:"rsa_key_len"`
}

// ProviderConf holds information about a provider
type ProviderConf struct {
	Issuer                   string             `yaml:"issuer"`
	ClientID                 string             `yaml:"client_id"`
	ClientSecret             string             `yaml:"client_secret"`
	Scopes                   []string           `yaml:"scopes"`
	Endpoints                *oauth2x.Endpoints `yaml:"-"`
	Provider                 *oidc.Provider     `yaml:"-"`
	Name                     string             `yaml:"name"`
	AudienceRequestParameter string             `yaml:"audience_request_parameter"`
}

var conf *Config

// Get returns the Config
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
	if conf.Server.TLS.Enabled {
		if len(conf.Server.TLS.Key) > 0 && len(conf.Server.TLS.Cert) > 0 {
			conf.Server.Port = 443
		} else {
			conf.Server.TLS.Enabled = false
		}
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
			return fmt.Errorf("error '%s' for provider.issuer '%s' (Index %d)", err, p.Issuer, i)
		}
		p.Provider, err = oidc.NewProvider(context.Get(), p.Issuer)
		if err != nil {
			return fmt.Errorf("error '%s' for provider.issuer '%s' (Index %d)", err, p.Issuer, i)
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
		if p.AudienceRequestParameter == "" {
			p.AudienceRequestParameter = "resource"
		}
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
	model.OIDCFlowAuthorizationCode.AddToSliceIfNotFound(&conf.Features.EnabledOIDCFlows)
	if model.OIDCFlowIsInSlice(model.OIDCFlowDevice, conf.Features.EnabledOIDCFlows) && !conf.Features.Polling.Enabled {
		return fmt.Errorf("oidc flow device flow requires polling_codes to be enabled")
	}
	return nil
}

var possibleConfigLocations = []string{
	"config",
	"/etc/mytoken",
}

// Load reads the config file and populates the Config struct; then validates the Config
func Load() {
	load()
	if err := validate(); err != nil {
		log.WithError(err).Fatal()
	}
}

func load() {
	data, _ := fileutil.ReadConfigFile("config.yaml", possibleConfigLocations)
	conf = &defaultConfig
	err := yaml.Unmarshal(data, conf)
	if err != nil {
		log.WithError(err).Fatal()
		return
	}
}

// LoadForSetup reads the config file and populates the Config struct; it does not validate the Config, since this is not required for setup
func LoadForSetup() {
	load()
}
