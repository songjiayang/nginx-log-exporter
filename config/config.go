package config

import (
	"io/ioutil"
	"log"
	"regexp"

	"gopkg.in/yaml.v2"
)

type Config struct {
	App []*AppConfig

	original string
}

type AppConfig struct {
	Name   string `yaml:"name"`
	Format string `yaml:"format"`

	SourceFiles   []string          `yaml:"source_files"`
	StaticConfig  map[string]string `yaml:"static_config"`
	RelabelConfig *RelabelConfig    `yaml:"relabel_config"`
	Buckets       []float64         `yaml:"histogram_buckets"`
}

func (cfg *AppConfig) StaticLabelValues() (labels, values []string) {
	labels = make([]string, len(cfg.StaticConfig))
	values = make([]string, len(cfg.StaticConfig))

	i := 0
	for k, v := range cfg.StaticConfig {
		labels[i] = k
		values[i] = v
		i++
	}

	return
}

func (cfg *AppConfig) DynamicLabels() (labels []string) {
	return cfg.RelabelConfig.SourceLabels
}

func (cfg *AppConfig) Prepare() {
	for _, r := range cfg.RelabelConfig.Replacement {
		for _, replaceItem := range r.Replaces {
			replaceItem.prepare()
		}
	}
}

type RelabelConfig struct {
	SourceLabels []string                `yaml:"source_labels"`
	Replacement  map[string]*Replacement `yaml:"replacement"`
}

type Replacement struct {
	Trim     string           `yaml:"trim"`
	Replaces []*ReplaceTarget `yaml:"replace"`
}

type ReplaceTarget struct {
	Target string `yaml:"target"`
	Value  string `yaml:"value"`

	tRex *regexp.Regexp
}

func (rt *ReplaceTarget) Regexp() *regexp.Regexp {
	return rt.tRex
}

func (rt *ReplaceTarget) prepare() {
	replace, err := regexp.Compile(rt.Target)
	if err != nil {
		log.Panic(err)
	}

	rt.tRex = replace
}

func (cfg *Config) Reload() error {
	original, err := load(cfg.original)
	if err != nil {
		return err
	}

	cfg = original
	return nil
}

func LoadFile(filename string) (conf *Config, err error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	conf, err = load(string(content))
	return
}

func load(s string) (*Config, error) {
	var (
		cfg  = &Config{}
		apps []*AppConfig
	)

	err := yaml.Unmarshal([]byte(s), &apps)
	if err != nil {
		return nil, err
	}

	cfg.original = s
	cfg.App = apps

	return cfg, nil
}
