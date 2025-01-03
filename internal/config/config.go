package config

import (
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/lestrrat-go/jwx/jwa"
	"github.com/oidc-mytoken/utils/context"
	utils2 "github.com/oidc-mytoken/utils/utils"
	"github.com/oidc-mytoken/utils/utils/fileutil"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	oidfed "github.com/zachmann/go-oidfed/pkg"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v3"

	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/server/paths"
	"github.com/oidc-mytoken/server/internal/utils"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"

	"github.com/oidc-mytoken/server/pkg/oauth2x"
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
	Signing: signingConfs{
		Mytoken: signingConf{
			Alg:       jwa.ES512,
			RSAKeyLen: 2048,
		},
		OIDC: signingConf{
			Alg:       jwa.ES512,
			RSAKeyLen: 2048,
		},
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
		OIDCFlows: oidcFlowsConf{
			AuthCode: authcodeConf{
				Web: authcodeWebClientsConf{
					CookieLifetime: 3600 * 24 * 7,
				},
			},
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
		WebInterface: webConfig{Enabled: true},
		SSH: sshConf{
			Enabled: false,
		},
		ServerProfiles: serverProfilesConf{
			Enabled: true,
			Groups:  make(map[string]string),
		},
		Notifications: notificationConf{
			Mail: MailNotificationConf{
				Enabled: false,
				MailServer: MailServerConf{
					Port: 587,
				},
			},
			ICS: onlyEnable{true},
		},
		Federation: federationConf{
			Enabled:                     false,
			EntityConfigurationLifetime: 7 * 24 * 60 * 60,
			Signing: signingConf{
				Alg:       jwa.ES512,
				RSAKeyLen: 2048,
			},
		},
	},
	API: apiConf{
		MinVersion: 0,
	},
	Caching: cacheConf{
		Internal: internalCacheConf{
			DefaultExpiration: 300,
			CleanupInterval:   600,
		},
	},
}

// Config holds the server configuration
type Config struct {
	IssuerURL            string              `yaml:"issuer"`
	Host                 string              // Extracted from the IssuerURL
	Server               serverConf          `yaml:"server"`
	GeoIPDBFile          string              `yaml:"geo_ip_db_file"`
	API                  apiConf             `yaml:"api"`
	DB                   DBConf              `yaml:"database"`
	Signing              signingConfs        `yaml:"signing"`
	Logging              loggingConf         `yaml:"logging"`
	ServiceDocumentation string              `yaml:"service_documentation"`
	Features             featuresConf        `yaml:"features"`
	Providers            []*ProviderConf     `yaml:"providers"`
	ServiceOperator      ServiceOperatorConf `yaml:"service_operator"`
	Caching              cacheConf           `yaml:"cache"`
}

type apiConf struct {
	MinVersion int `yaml:"min_supported_version"`
}

type featuresConf struct {
	OIDCFlows               oidcFlowsConf           `yaml:"oidc_flows"`
	TokenRevocation         onlyEnable              `yaml:"token_revocation"`
	ShortTokens             shortTokenConfig        `yaml:"short_tokens"`
	TransferCodes           onlyEnable              `yaml:"transfer_codes"`
	Polling                 pollingConf             `yaml:"polling_codes"`
	TokenRotation           onlyEnable              `yaml:"token_rotation"`
	TokenInfo               tokeninfoConfig         `yaml:"tokeninfo"`
	WebInterface            webConfig               `yaml:"web_interface"`
	DisabledRestrictionKeys model.RestrictionClaims `yaml:"unsupported_restrictions"`
	SSH                     sshConf                 `yaml:"ssh"`
	ServerProfiles          serverProfilesConf      `yaml:"server_profiles"`
	Federation              federationConf          `yaml:"federation"`
	GuestMode               onlyEnable              `yaml:"guest_mode"`
	Notifications           notificationConf        `yaml:"notifications"`
}

func (c *featuresConf) validate() error {
	if err := c.OIDCFlows.validate(); err != nil {
		return err
	}
	if err := c.ServerProfiles.validate(); err != nil {
		return err
	}
	if err := c.Federation.validate(); err != nil {
		return err
	}
	if err := c.Notifications.validate(); err != nil {
		return err
	}
	if err := c.SSH.validate(); err != nil {
		return err
	}
	return nil
}

type oidcFlowsConf struct {
	AuthCode authcodeConf `yaml:"authorization_code"`
}

type authcodeConf struct {
	Web authcodeWebClientsConf `yaml:"web"`
}

type authcodeWebClientsConf struct {
	TrustedRedirectURIs   []string         `yaml:"trusted_redirect_uris"`
	TrustedRedirectsRegex []*regexp.Regexp `yaml:"-"`
	CookieLifetime        int              `yaml:"cookie_lifetime"`
}

func (c *oidcFlowsConf) validate() error {
	return c.AuthCode.validate()
}
func (c *authcodeConf) validate() error {
	return c.Web.validate()
}
func (c *authcodeWebClientsConf) validate() error {
	for _, r := range c.TrustedRedirectURIs {
		reg, err := regexp.Compile(r)
		if err != nil {
			return errors.Errorf("invalid config: invalid regex in truested_redirect_uris: '%s'", r)
		}
		c.TrustedRedirectsRegex = append(c.TrustedRedirectsRegex, reg)
	}
	return nil
}

type sshConf struct {
	Enabled          bool         `yaml:"enabled"`
	UseProxyProtocol bool         `yaml:"use_proxy_protocol"`
	KeyFiles         []string     `yaml:"keys"`
	PrivateKeys      []ssh.Signer `yaml:"-"`
}

func (c *sshConf) validate() error {
	if !c.Enabled {
		return nil
	}
	if len(c.KeyFiles) == 0 {
		return errors.New("invalid config: ssh feature enabled, but no ssh private key set")
	}
	for _, pkf := range c.KeyFiles {
		pemBytes, err := os.ReadFile(pkf)
		if err != nil {
			return errors.Wrap(err, "reading ssh private key")
		}
		signer, err := ssh.ParsePrivateKey(pemBytes)
		if err != nil {
			return errors.Wrap(err, "parsing ssh private key")
		}
		c.PrivateKeys = append(c.PrivateKeys, signer)
	}
	return nil
}

type serverProfilesConf struct {
	Enabled bool                     `yaml:"enabled"`
	Groups  profileGroupsCredentials `yaml:"groups"`
}

func (c serverProfilesConf) validate() error {
	return c.Groups.validate()
}

// profileGroupsCredentials holds the credentials for a profile groups
type profileGroupsCredentials map[string]string

func (g profileGroupsCredentials) validate() error {
	for u, pw := range g {
		if u == "" {
			return errors.New("'name' not set in profile group")
		}
		if pw == "" {
			return errors.Errorf("'password' not set in profile group '%s'", u)
		}
	}
	return nil
}

type notificationConf struct {
	AnyEnabled     bool                 `yaml:"-"`
	Mail           MailNotificationConf `yaml:"email"`
	Websocket      onlyEnable           `yaml:"ws"`
	ICS            onlyEnable           `yaml:"ics"`
	NotifierServer string               `yaml:"notifier_server_url"`
}

func (c *notificationConf) validate() error {
	c.AnyEnabled = c.Mail.Enabled || c.Websocket.Enabled || c.ICS.Enabled
	if !c.AnyEnabled {
		return nil
	}
	if conf.Server.DistributedServers {
		if c.NotifierServer == "" {
			return errors.New("distributed deployment, but no notifier_server_url set")
		}
	}
	if c.NotifierServer != "" {
		if c.Mail.MailServer.Host != "" {
			log.Warning(
				"a standalone notifier server is used; however mail_server configuration is given here",
			)
		}
	}
	return nil
}

// MailNotificationConf holds the configuration for email notifications
type MailNotificationConf struct {
	Enabled      bool           `yaml:"enabled"`
	MailServer   MailServerConf `yaml:"mail_server"`
	OverwriteDir string         `yaml:"overwrite_dir"`
}

// MailServerConf holds the configuration for the email server
type MailServerConf struct {
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	Username    string `yaml:"user"`
	Password    string `yaml:"password"`
	FromAddress string `yaml:"from_address"`
}

type tokeninfoConfig struct {
	Enabled    bool       `yaml:"-"`
	Introspect onlyEnable `yaml:"introspect"`
	History    onlyEnable `yaml:"event_history"`
	Tree       onlyEnable `yaml:"subtoken_tree"`
	List       onlyEnable `yaml:"list_mytokens"`
}

type webConfig struct {
	Enabled      bool   `yaml:"enabled"`
	OverwriteDir string `yaml:"overwrite_dir"`
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
	ProxyHeader        string           `yaml:"proxy_header"`
	Limiter            limiterConf      `yaml:"request_limits"`
	DistributedServers bool             `yaml:"distributed_servers"`
	Healthcheck        healtcheckConfig `yaml:"healthcheck"`
}

type healtcheckConfig struct {
	Enabled bool `yaml:"enabled"`
	Port    int  `yaml:"port"`
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

type signingConfs struct {
	Mytoken signingConf `yaml:"mytoken"`
	OIDC    signingConf `yaml:"oidc"`
}

type signingConf struct {
	Alg       jwa.SignatureAlgorithm `yaml:"alg"`
	KeyFile   string                 `yaml:"key_file"`
	RSAKeyLen int                    `yaml:"rsa_key_len"`
}

// ProviderConf holds information about a provider
type ProviderConf struct {
	Issuer               string                   `yaml:"issuer"`
	ClientID             string                   `yaml:"client_id"`
	ClientSecret         string                   `yaml:"client_secret"`
	Scopes               []string                 `yaml:"scopes"`
	MytokensMaxLifetime  int64                    `yaml:"mytokens_max_lifetime"`
	EnforcedRestrictions EnforcedRestrictionsConf `yaml:"enforced_restrictions"`
	Endpoints            *oauth2x.Endpoints       `yaml:"-"`
	Name                 string                   `yaml:"name"`
	Audience             *model.AudienceConf      `yaml:"audience"`
}

// EnforcedRestrictionsConf is a type for holding configuration for enforced restrictions
type EnforcedRestrictionsConf struct {
	Enabled         bool              `yaml:"-"`
	ClaimSources    map[string]string `yaml:"claim_sources"`
	DefaultTemplate string            `yaml:"default_template"`
	ForbidOnDefault bool              `yaml:"forbid_on_default"`
	HelpHTMLText    string            `yaml:"help_html"`
	HelpHTMLFile    string            `yaml:"help_html_file"`
	Mapping         map[string]string `yaml:"mapping"`
}

func (c *EnforcedRestrictionsConf) validate() error {
	if len(c.ClaimSources) >= 1 {
		c.Enabled = true
	}
	if c.HelpHTMLFile != "" {
		content, err := os.ReadFile(c.HelpHTMLFile)
		if err != nil {
			return errors.Wrapf(
				err,
				"error reading enforced restrictions help html file '%s'", c.HelpHTMLFile,
			)
		}
		c.HelpHTMLText = string(content)
	}
	return nil
}

// ServiceOperatorConf is type holding the configuration for the service operator of this mytoken instance
type ServiceOperatorConf struct {
	Name     string `yaml:"name"`
	Homepage string `yaml:"homepage"`
	Contact  string `yaml:"mail_contact"`
	Privacy  string `yaml:"mail_privacy"`
}

type cacheConf struct {
	Internal internalCacheConf  `yaml:"internal"`
	External *externalCacheConf `yaml:"external"`
}

type internalCacheConf struct {
	DefaultExpiration int64 `yaml:"default_ttl"`
	CleanupInterval   int64 `yaml:"cleanup_interval"`
}

type externalCacheConf struct {
	Redis *redisCacheConf `yaml:"redis"`
}

type redisCacheConf struct {
	Addr     string `yaml:"addr"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// GetPassword returns the password for this database config. If necessary it reads it from the password file.
func (conf *DBConf) GetPassword() string {
	if conf.PasswordFile == "" {
		return conf.Password
	}
	content, err := os.ReadFile(conf.PasswordFile)
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

type federationConf struct {
	Enabled                     bool                   `yaml:"enabled"`
	TrustAnchors                oidfed.TrustAnchors    `yaml:"trust_anchors"`
	AuthorityHints              []string               `yaml:"authority_hints"`
	EntityConfigurationLifetime int64                  `yaml:"entity_configuration_lifetime"`
	Signing                     signingConf            `yaml:"signing"`
	Entity                      *oidfed.FederationLeaf `yaml:"-"`
}

func (f *federationConf) validate() (err error) {
	if !f.Enabled {
		return nil
	}
	if Get().Signing.OIDC.KeyFile == "" {
		return errors.New("if federation is enabled an OIDC signing key must be set under signing.oidc.key_file")
	}
	if Get().Signing.OIDC.Alg == "" {
		return errors.New("if federation is enabled an OIDC signing alg must be set under signing.oidc.alg")
	}
	if len(f.TrustAnchors) == 0 {
		return errors.New("federation enabled, but no trust anchors specified")
	}
	if len(f.AuthorityHints) == 0 {
		return errors.New("federation enabled, but no authority hints specified")
	}
	if f.Signing.KeyFile == "" {
		return errors.New("federation enabled, but no signing keyfile specified")
	}
	if f.Signing.Alg == "" {
		return errors.New("federation enabled, but no signing alg specified")
	}
	if f.EntityConfigurationLifetime == 0 {
		f.EntityConfigurationLifetime = 7 * 24 * 60 * 60
	}

	return
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
	if err := validateIssuerURL(); err != nil {
		return err
	}
	if err := configureServerTLS(); err != nil {
		return err
	}
	if err := validateConfigSections(); err != nil {
		return err
	}
	if err := validateProviders(); err != nil {
		return err
	}
	if conf.Features.GuestMode.Enabled {
		addGuestModeProvider()
	}
	if err := validateSigningConfig(); err != nil {
		return err
	}
	return validateWebInterface()
}

func validateIssuerURL() error {
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
	return nil
}

func configureServerTLS() error {
	if conf.Server.TLS.Enabled {
		if conf.Server.TLS.Key != "" && conf.Server.TLS.Cert != "" {
			conf.Server.Port = 443
		} else {
			conf.Server.TLS.Enabled = false
		}
	}
	return nil
}

func validateConfigSections() error {
	if err := conf.Logging.validate(); err != nil {
		return err
	}

	if err := conf.ServiceOperator.validate(); err != nil {
		return err
	}

	return conf.Features.validate()
}

func validateProviders() error {
	if len(conf.Providers) == 0 {
		return errors.New("invalid config: providers must have at least one entry")
	}
	for i, p := range conf.Providers {
		if err := validateProvider(p, i); err != nil {
			return err
		}
		conf.Providers[i] = p
	}
	return nil
}

func validateProvider(p *ProviderConf, i int) error {
	if p.Issuer == "" {
		return errors.Errorf("invalid config: provider.issuer not set (Index %d)", i)
	}
	if err := p.EnforcedRestrictions.validate(); err != nil {
		return err
	}
	oc, err := oauth2x.NewConfig(context.Get(), p.Issuer)
	if err != nil {
		return errors.Errorf("error '%s' for provider.issuer '%s' (Index %d)", err, p.Issuer, i)
	}
	p.Endpoints, err = oc.Endpoints()
	if err != nil {
		return errors.Errorf("error '%s' for provider.issuer '%s' (Index %d)", err, p.Issuer, i)
	}
	if p.ClientID == "" {
		return errors.Errorf("invalid config: provider.clientid not set (Index %d)", i)
	}
	if p.ClientSecret == "" {
		return errors.Errorf("invalid config: provider.clientsecret not set (Index %d)", i)
	}
	if len(p.Scopes) == 0 {
		return errors.Errorf("invalid config: provider.scopes not set (Index %d)", i)
	}
	if p.Audience == nil {
		p.Audience = &model.AudienceConf{RFC8707: true}
	}
	if p.Audience.RFC8707 {
		p.Audience.RequestParameter = model.AudienceParameterResource
		p.Audience.SpaceSeparateAuds = false
	} else if p.Audience.RequestParameter == "" {
		p.Audience.RequestParameter = model.AudienceParameterResource
	}
	return nil
}

func addGuestModeProvider() {
	iss := utils2.CombineURLPath(conf.IssuerURL, paths.GetCurrentAPIPaths().GuestModeOP)
	p := &ProviderConf{
		Issuer: iss,
		Name:   "Guest Mode",
		Scopes: []string{"openid"},
		Endpoints: &oauth2x.Endpoints{
			Authorization: utils2.CombineURLPath(iss, "auth"),
			Token:         utils2.CombineURLPath(iss, "token"),
		},
	}
	conf.Providers = append(conf.Providers, p)
}

func validateSigningConfig() error {
	if conf.Signing.Mytoken.KeyFile == "" {
		return errors.New("invalid config: signing keyfile not set")
	}
	if conf.Signing.Mytoken.Alg == "" {
		return errors.New("invalid config: token signing alg not set")
	}
	return nil
}

func validateWebInterface() error {
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
		log.Fatalf("%s", errorfmt.Full(err)) // skipcq RVV-A0003
	}
}

func load() {
	data, _ := fileutil.MustReadConfigFile("config.yaml", possibleConfigLocations)
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
