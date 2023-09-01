package collector

import (
	"fmt"
	"github.com/hpcloud/tail"
	"log"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/satyrius/gonx"

	"github.com/songjiayang/nginx-log-exporter/config"
)

// Collector is a struct containing pointers to all metrics that should be
// exposed to Prometheus
type Collector struct {
	countTotal      *prometheus.CounterVec
	bytesTotal      *prometheus.CounterVec
	upstreamSeconds *prometheus.HistogramVec
	responseSeconds *prometheus.HistogramVec

	externalValues  []string
	dynamicLabels   []string
	dynamicValueLen int
	trackedFiles    []string // List of tracked files to compare which are newly matched files

	cfg    *config.AppConfig
	opts   config.Options
	parser *gonx.Parser

	pollMu sync.Mutex // protects Poll()

}

func NewCollector(cfg *config.AppConfig, opts config.Options) *Collector {
	exlables, exValues := cfg.ExternalLabelSets()
	dynamicLabels := cfg.DynamicLabels()

	labels := append(exlables, dynamicLabels...)

	return &Collector{
		countTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: cfg.Name,
			Name:      "http_response_count_total",
			Help:      "Amount of processed HTTP requests",
		}, labels),

		bytesTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: cfg.Name,
			Name:      "http_response_size_bytes",
			Help:      "Total amount of transferred bytes",
		}, labels),

		upstreamSeconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: cfg.Name,
			Name:      "http_upstream_time_seconds",
			Help:      "Time needed by upstream servers to handle requests",
			Buckets:   cfg.HistogramBuckets,
		}, labels),

		responseSeconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: cfg.Name,
			Name:      "http_response_time_seconds",
			Help:      "Time needed by NGINX to handle requests",
			Buckets:   cfg.HistogramBuckets,
		}, labels),

		externalValues:  exValues,
		dynamicLabels:   dynamicLabels,
		dynamicValueLen: len(dynamicLabels),

		cfg:    cfg,
		opts:   opts,
		parser: gonx.NewParser(cfg.Format),
	}
}

func (c *Collector) Run() {
	c.cfg.Prepare()

	// register to prometheus
	prometheus.MustRegister(c.countTotal)
	prometheus.MustRegister(c.bytesTotal)
	prometheus.MustRegister(c.upstreamSeconds)
	prometheus.MustRegister(c.responseSeconds)
	c.pollLogFiles() // find all files match glob pattern, and tail -f
	if c.opts.PollLogInterval() > 0 {
		c.startLogFilesPollLoop() // start a ticker, and periodic scanning of matching new files
	}
}

func (c *Collector) pollLogFiles() {
	c.pollMu.Lock()
	defer c.pollMu.Unlock()
	for _, f := range c.cfg.SourceFiles {
		// find all files match glob pattern
		matches, err := filepath.Glob(f)
		if err != nil {
			log.Panic(err)
		}
		for _, pathname := range matches {
			absPath, err := filepath.Abs(pathname)
			if err != nil {
				log.Panic(err)
				continue
			}
			c.tailPath(absPath)
		}
	}
}

func (c *Collector) startLogFilesPollLoop() {
	go func() {
		// periodic scanning of matching new files
		ticker := time.NewTicker(c.opts.PollLogInterval())
		defer ticker.Stop()
		for range ticker.C {
			c.pollLogFiles()
		}
	}()
}

// TailPath registers a filesystem pathname to be tailed.
func (c *Collector) tailPath(pathname string) {
	if contains(c.trackedFiles, pathname) {
		//log.Printf("file:%s has tracked, ignore.", pathname)
		return
	}
	log.Printf("begin tail file: %s", pathname)
	t, err := tail.TailFile(pathname, tail.Config{
		Follow: true,
		ReOpen: true,
		Poll:   true,
	})

	if err != nil {
		log.Panic(err)
	}
	// add to tracked file list
	c.trackedFiles = append(c.trackedFiles, pathname)

	go func() {
		for line := range t.Lines {
			entry, err := c.parser.ParseString(line.Text)
			if err != nil {
				fmt.Printf("error while parsing line '%s': %s", line.Text, err)
				continue
			}

			dynamicValues := make([]string, c.dynamicValueLen)

			for i, label := range c.dynamicLabels {
				if s, err := entry.Field(label); err == nil {
					dynamicValues[i] = c.formatValue(label, s)
				}
			}

			labelValues := append(c.externalValues, dynamicValues...)

			c.countTotal.WithLabelValues(labelValues...).Inc()

			if bytes, err := entry.FloatField("body_bytes_sent"); err == nil {
				c.bytesTotal.WithLabelValues(labelValues...).Add(bytes)
			}

			c.updateHistogramMetric(c.upstreamSeconds, labelValues, entry, "upstream_response_time")
			c.updateHistogramMetric(c.responseSeconds, labelValues, entry, "request_time")
		}
	}()
}

func (c *Collector) formatValue(label, value string) string {
	replacement, ok := c.cfg.RelabelConfig.Replacement[label]
	if !ok {
		return value
	}

	if replacement.Trim != "" {
		value = strings.Split(value, replacement.Trim)[0]
	}

	for _, target := range replacement.Replaces {
		if c.opts.EnablePlaceholderReplace() && target.Regexp().MatchString(value) {
			// value contains placeholder
			hasPlaceHolder := target.PlaceHolderRex().MatchString(target.Value)
			if hasPlaceHolder {
				matches := target.Regexp().FindStringSubmatch(value)
				// reslove placeHolders
				return target.PlaceHolderRex().ReplaceAllStringFunc(target.Value, func(src string) string {
					index, _ := strconv.Atoi(src[2:3])
					return matches[index]
				})
			} else {
				return target.Value
			}
		}
	}

	return value
}

func (c *Collector) updateHistogramMetric(metric *prometheus.HistogramVec, labelValues []string, entry *gonx.Entry, field string) {
	value, err := entry.FloatField(field)
	if err != nil {
		//sometime the value duration
		field, err := entry.Field(field)
		if err != nil {
			return
		}
		duration, err := time.ParseDuration(field)
		if err != nil {
			return
		}
		value = duration.Seconds()
	}

	exemplarLabels := c.cfg.ExemplarMatch(entry, field)
	if exemplarLabels == nil {
		metric.WithLabelValues(labelValues...).Observe(value)
		return
	}

	metric.WithLabelValues(labelValues...).(prometheus.ExemplarObserver).ObserveWithExemplar(
		value, *exemplarLabels,
	)
}

// Contains method for a slice
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
