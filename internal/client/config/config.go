package config

type config struct {
	URL string `yaml:"url"`
}

var conf *config

// Get returns the config
func Get() *config {
	return conf
}

func Init() {
	conf = &config{
		URL: "http://localhost:8000",
	}
}
