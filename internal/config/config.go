package config

import (
	"io/ioutil"
	"net/url"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v3"

	model2 "github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"

	"github.com/oidc-mytoken/server/pkg/oauth2x"
	"github.com/oidc-mytoken/server/shared/context"
	"github.com/oidc-mytoken/server/shared/model"
	"github.com/oidc-mytoken/server/shared/utils"
	"github.com/oidc-mytoken/server/shared/utils/fileutil"
	"github.com/oidc-mytoken/server/shared/utils/issuerUtils"
)

var defaultConfig = Config{
	Server: serverConf{
		Port: 8000,
		TLS: tlsConf{
			Enabled: true, // The default is that TLS is enabled if cert and key are given, this is checked later;
			// we must set true here, because otherwise we cannot distinct this from a false set by the user
			RedirectHTTP: true,
		},
		Secure: true,
		Limiter: limiterConf{
			Enabled:     true,
			Max:         100,
			Window:      300,
			AlwaysAllow: []string{"127.0.0.1"},
		},
	},
	DB: DBConf{
		Hosts:             []string{"localhost"},
		User:              "mytoken",
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
		Internal: internalLoggerConf{
			LoggerConf: LoggerConf{
				Dir:    "/var/log/mytoken",
				StdErr: false,
				Level:  "error",
			},
			Smart: smartLoggerConf{
				Enabled: true,
				Dir:     "", // if empty equal to normal logging dir
			},
		},
	},
	ServiceDocumentation: "https://mytoken-docs.data.kit.edu/",
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
		TokenRotation: onlyEnable{true},
		TokenInfo: tokeninfoConfig{
			Introspect: onlyEnable{true},
			History:    onlyEnable{true},
			Tree:       onlyEnable{true},
			List:       onlyEnable{true},
		},
		WebInterface: onlyEnable{true},
		SSH: sshConf{
			Enabled: false,
		},
	},
	ProviderByIssuer: make(map[string]*ProviderConf),
	API: apiConf{
		MinVersion: 0,
	},
}

// Config holds the server configuration
type Config struct {
	IssuerURL            string                   `yaml:"issuer"`
	Host                 string                   // Extracted from the IssuerURL
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
	ServiceOperator      ServiceOperatorConf      `yaml:"service_operator"`
}

type apiConf struct {
	MinVersion int `yaml:"min_supported_version"`
}

type featuresConf struct {
	EnabledOIDCFlows        []model.OIDCFlow         `yaml:"enabled_oidc_flows"`
	TokenRevocation         onlyEnable               `yaml:"token_revocation"`
	ShortTokens             shortTokenConfig         `yaml:"short_tokens"`
	TransferCodes           onlyEnable               `yaml:"transfer_codes"`
	Polling                 pollingConf              `yaml:"polling_codes"`
	TokenRotation           onlyEnable               `yaml:"token_rotation"`
	TokenInfo               tokeninfoConfig          `yaml:"tokeninfo"`
	WebInterface            onlyEnable               `yaml:"web_interface"`
	DisabledRestrictionKeys model2.RestrictionClaims `yaml:"unsupported_restrictions"`
	SSH                     sshConf                  `yaml:"ssh"`
}

type sshConf struct {
	Enabled          bool         `yaml:"enabled"`
	UseProxyProtocol bool         `yaml:"use_proxy_protocol"`
	KeyFiles         []string     `yaml:"keys"`
	PrivateKeys      []ssh.Signer `yaml:"-"`
}

type tokeninfoConfig struct {
	Enabled    bool       `yaml:"-"`
	Introspect onlyEnable `yaml:"introspect"`
	History    onlyEnable `yaml:"event_history"`
	Tree       onlyEnable `yaml:"subtoken_tree"`
	List       onlyEnable `yaml:"list_mytokens"`
}

type shortTokenConfig struct {
	Enabled bool `yaml:"enabled"`
	Len     int  `yaml:"len"`
}

type onlyEnable struct {
	Enabled bool `yaml:"enabled"`
}

type loggingConf struct {
	Access   LoggerConf         `yaml:"access"`
	Internal internalLoggerConf `yaml:"internal"`
}

type internalLoggerConf struct {
	LoggerConf `yaml:",inline"`
	Smart      smartLoggerConf `yaml:"smart"`
}

// LoggerConf holds configuration related to logging
type LoggerConf struct {
	Dir    string `yaml:"dir"`
	StdErr bool   `yaml:"stderr"`
	Level  string `yaml:"level"`
}

type smartLoggerConf struct {
	Enabled bool   `yaml:"enabled"`
	Dir     string `yaml:"dir"`
}

func checkLoggingDirExists(dir string) error {
	if dir != "" && !fileutil.FileExists(dir) {
		return errors.Errorf("logging directory '%s' does not exist", dir)
	}
	return nil
}

func (log *loggingConf) validate() error {
	if err := checkLoggingDirExists(log.Access.Dir); err != nil {
		return err
	}
	if err := checkLoggingDirExists(log.Internal.Dir); err != nil {
		return err
	}
	if log.Internal.Smart.Enabled {
		if log.Internal.Smart.Dir == "" {
			log.Internal.Smart.Dir = log.Internal.Dir
		}
		if err := checkLoggingDirExists(log.Internal.Smart.Dir); err != nil {
			return err
		}
	}
	return nil
}

type pollingConf struct {
	Enabled                 bool  `yaml:"enabled"`
	Len                     int   `yaml:"len"`
	PollingCodeExpiresAfter int64 `yaml:"expires_after"`
	PollingInterval         int64 `yaml:"polling_interval"`
}

// DBConf is type for holding configuration for a db
type DBConf struct {
	Hosts                  []string `yaml:"hosts"`
	User                   string   `yaml:"user"`
	Password               string   `yaml:"password"`
	PasswordFile           string   `yaml:"password_file"`
	DB                     string   `yaml:"db"`
	ReconnectInterval      int64    `yaml:"try_reconnect_interval"`
	EnableScheduledCleanup bool     `yaml:"schedule_cleanup"`
}

type serverConf struct {
	Port   int     `yaml:"port"`
	TLS    tlsConf `yaml:"tls"`
	Secure bool    `yaml:"-"` // Secure indicates if the connection to the mytoken server is secure. This is
	// independent of TLS, e.g. a Proxy can be used.
	ProxyHeader string      `yaml:"proxy_header"`
	Limiter     limiterConf `yaml:"request_limits"`
}

type limiterConf struct {
	Enabled     bool     `yaml:"enabled"`
	Max         int      `yaml:"max_requests"`
	Window      int      `yaml:"window"`
	AlwaysAllow []string `yaml:"always_allow"`
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
	MytokensMaxLifetime      int64              `yaml:"mytokens_max_lifetime"`
	Endpoints                *oauth2x.Endpoints `yaml:"-"`
	Provider                 *oidc.Provider     `yaml:"-"`
	Name                     string             `yaml:"name"`
	AudienceRequestParameter string             `yaml:"audience_request_parameter"`
}

// ServiceOperatorConf is type holding the configuration for the service operator of this mytoken instance
type ServiceOperatorConf struct {
	Name     string `yaml:"name"`
	Homepage string `yaml:"homepage"`
	Contact  string `yaml:"mail_contact"`
	Privacy  string `yaml:"mail_privacy"`
}

// GetPassword returns the password for this database config. If necessary it reads it from the password file.
func (conf *DBConf) GetPassword() string {
	if conf.Password != "" {
		return conf.Password
	}
	if conf.PasswordFile == "" {
		return ""
	}
	content, err := ioutil.ReadFile(conf.PasswordFile)
	if err != nil {
		log.WithError(err).Error()
		return ""
	}
	conf.Password = strings.Split(string(content), "\n")[0]
	return conf.Password
}

func (so *ServiceOperatorConf) validate() error {
	if so.Name == "" {
		return errors.New("invalid config: service_operator.name not set")
	}
	if so.Contact == "" {
		return errors.New("invalid config: service_operator.mail_contact not set")
	}
	if so.Privacy == "" {
		so.Privacy = so.Contact
	}
	return nil
}

var conf *Config

// Get returns the Config
func Get() *Config {
	return conf
}

func init() {
	conf = &defaultConfig
}

func validate() error {
	if conf == nil {
		return errors.New("config not set")
	}
	if conf.IssuerURL == "" {
		return errors.New("invalid config: issuer_url not set")
	}
	//goland:noinspection HttpUrlsUsage
	if strings.HasPrefix(conf.IssuerURL, "http://") {
		conf.Server.Secure = false
	}
	u, err := url.Parse(conf.IssuerURL)
	if err != nil {
		return errors.Wrap(err, "invalid config: issuer_url not valid")
	}
	conf.Host = u.Hostname()
	if conf.Server.TLS.Enabled {
		if conf.Server.TLS.Key != "" && conf.Server.TLS.Cert != "" {
			conf.Server.Port = 443
		} else {
			conf.Server.TLS.Enabled = false
		}
	}
	if err = conf.Logging.validate(); err != nil {
		return err
	}
	if err = conf.ServiceOperator.validate(); err != nil {
		return err
	}
	if len(conf.Providers) <= 0 {
		return errors.New("invalid config: providers must have at least one entry")
	}
	for i, p := range conf.Providers {
		if p.Issuer == "" {
			return errors.Errorf("invalid config: provider.issuer not set (Index %d)", i)
		}
		oc, err := oauth2x.NewConfig(context.Get(), p.Issuer)
		if err != nil {
			return errors.Errorf("error '%s' for provider.issuer '%s' (Index %d)", err, p.Issuer, i)
		}
		// Endpoints only returns an error if it does discovery but this was already done in NewConfig, so we can ignore
		// the error value
		p.Endpoints, _ = oc.Endpoints()
		p.Provider, err = oidc.NewProvider(context.Get(), p.Issuer)
		if err != nil {
			return errors.Errorf("error '%s' for provider.issuer '%s' (Index %d)", err, p.Issuer, i)
		}
		if p.ClientID == "" {
			return errors.Errorf("invalid config: provider.clientid not set (Index %d)", i)
		}
		if p.ClientSecret == "" {
			return errors.Errorf("invalid config: provider.clientsecret not set (Index %d)", i)
		}
		if len(p.Scopes) <= 0 {
			return errors.Errorf("invalid config: provider.scopes not set (Index %d)", i)
		}
		iss0, iss1 := issuerUtils.GetIssuerWithAndWithoutSlash(p.Issuer)
		conf.ProviderByIssuer[iss0] = p
		conf.ProviderByIssuer[iss1] = p
		if p.AudienceRequestParameter == "" {
			p.AudienceRequestParameter = "resource"
		}
	}
	if conf.IssuerURL == "" {
		return errors.New("invalid config: issuer_url not set")
	}
	if conf.Signing.KeyFile == "" {
		return errors.New("invalid config: signing keyfile not set")
	}
	if conf.Signing.Alg == "" {
		return errors.New("invalid config: token signing alg not set")
	}
	model.OIDCFlowAuthorizationCode.AddToSliceIfNotFound(&conf.Features.EnabledOIDCFlows)
	if conf.Features.SSH.Enabled {
		if len(conf.Features.SSH.KeyFiles) == 0 {
			return errors.New("invalid config: ssh feature enabled, but no ssh private key set")
		}
		for _, pkf := range conf.Features.SSH.KeyFiles {
			pemBytes, err := ioutil.ReadFile(pkf)
			if err != nil {
				return errors.Wrap(err, "reading ssh private key")
			}
			signer, err := ssh.ParsePrivateKey(pemBytes)
			if err != nil {
				return errors.Wrap(err, "parsing ssh private key")
			}
			conf.Features.SSH.PrivateKeys = append(conf.Features.SSH.PrivateKeys, signer)
		}
	}
	if !conf.Features.TokenInfo.Introspect.Enabled && conf.Features.WebInterface.Enabled {
		return errors.New("web interface requires tokeninfo.introspect to be enabled")
	}
	conf.Features.TokenInfo.Enabled = utils.OR(
		conf.Features.TokenInfo.Introspect.Enabled,
		conf.Features.TokenInfo.History.Enabled,
		conf.Features.TokenInfo.Tree.Enabled,
		conf.Features.TokenInfo.List.Enabled,
	)
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
		log.Fatalf("%s", errorfmt.Full(err))
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

// LoadForSetup reads the config file and populates the Config struct; it does not validate the Config, since this is
// not required for setup
func LoadForSetup() {
	load()
	conf.Logging.Internal.StdErr = true
}
