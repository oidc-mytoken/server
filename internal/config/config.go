package config

// Config holds the server configuration
type Config struct {
	DB struct {
		Host     string
		User     string
		Password string
		DB       string
	}
	Server struct {
		Hostname string
	}
	IssuerURL      string
	SigningKeyFile string
}

var conf *Config

// Get returns the config
func Get() *Config {
	return conf
}

// init creates dummy config TODO remove
func init() {
	conf = &Config{
		IssuerURL:      "https://localhost:8000/mytoken",
		SigningKeyFile: "/tmp/mytoken.key",
	}
	conf.DB.Host = "localhost"
	conf.DB.User = "mytoken"
	conf.DB.Password = "mytoken"
	conf.DB.DB = "mytoken_test"
	conf.Server.Hostname = "localhost"
}
