package config

import (
	"io/ioutil"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/satyrius/gonx"

	"gopkg.in/yaml.v2"
)

type Config struct {
	App []*AppConfig

	original string
}

type AppConfig struct {
	Name   string `yaml:"name"`
	Format string `yaml:"format"`

	SourceFiles      []string          `yaml:"source_files"`
	ExternalLabels   map[string]string `yaml:"external_labels"`
	RelabelConfig    *RelabelConfig    `yaml:"relabel_config"`
	HistogramBuckets []float64         `yaml:"histogram_buckets"`
	ExemplarConfig   *ExemplarConfig   `yaml:"exemplar_config"`
}

func (cfg *AppConfig) ExternalLabelSets() (labels, values []string) {
	labels = make([]string, len(cfg.ExternalLabels))
	values = make([]string, len(cfg.ExternalLabels))

	i := 0
	for k, v := range cfg.ExternalLabels {
		labels[i] = k
		values[i] = v
		i++
	}

	return
}

func (cfg *AppConfig) DynamicLabels() (labels []string) {
	return cfg.RelabelConfig.SourceLabels
}

func (cfg *AppConfig) ExemplarMatch(entry *gonx.Entry, field string) *prometheus.Labels {
	if cfg.ExemplarConfig == nil {
		return nil
	}

	return cfg.ExemplarConfig.Match(entry, field)
}

func (cfg *AppConfig) Prepare() {
	for _, r := range cfg.RelabelConfig.Replacement {
		for _, replaceItem := range r.Replaces {
			replaceItem.prepare()
		}
	}

	if cfg.ExemplarConfig != nil {
		cfg.ExemplarConfig.load()
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

type ExemplarConfig struct {
	MatchConfig map[string]string `yaml:"match"`
	Labels      []string          `yaml:"labels"`
	matchers    map[string]*ExemplarMatcher
}

type ExemplarMatcher struct {
	isEqual  bool
	isBigger bool
	value    float64
}

func newExemplarMatcher(matchStr string) *ExemplarMatcher {
	ret := &ExemplarMatcher{
		isEqual:  strings.Contains(matchStr, "="),
		isBigger: strings.Contains(matchStr, ">"),
	}

	valueStr := strings.Replace(matchStr, "=", "", -1)
	valueStr = strings.Replace(valueStr, ">", "", -1)
	valueStr = strings.TrimSpace(valueStr)

	// ignore err here
	ret.value, _ = strconv.ParseFloat(valueStr, 64)

	return ret
}

func (ec *ExemplarConfig) load() {
	ec.matchers = make(map[string]*ExemplarMatcher)
	for k, v := range ec.MatchConfig {
		ec.matchers[k] = newExemplarMatcher(v)
	}
}

func (ec *ExemplarConfig) Match(entry *gonx.Entry, field string) *prometheus.Labels {
	var matched bool

	matcher, ok := ec.matchers[field]
	if !ok {
		return nil
	}

	if value, err := entry.FloatField(field); err == nil {
		if matcher.isEqual && value == matcher.value {
			matched = true
		}

		if matcher.isBigger && value > matcher.value {
			matched = true
		}
	}

	if !matched {
		return nil
	}

	ret := prometheus.Labels{}
	for _, k := range ec.Labels {
		if value, err := entry.Field(k); err == nil {
			ret[k] = value
		}
	}
	return &ret
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
