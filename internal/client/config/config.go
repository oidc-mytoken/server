package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/zachmann/mytoken/internal/utils/fileutil"
	"github.com/zachmann/mytoken/pkg/mytokenlib"
)

type config struct {
	URL     string              `yaml:"url"`
	Mytoken *mytokenlib.Mytoken `yaml:"-"`
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
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	mytoken, err := mytokenlib.NewMytokenInstance(conf.URL)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	conf.Mytoken = mytoken
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
