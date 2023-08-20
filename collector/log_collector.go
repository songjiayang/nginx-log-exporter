package collector

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/hpcloud/tail"
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

	cfg    *config.AppConfig
	parser *gonx.Parser
}

func NewCollector(cfg *config.AppConfig) *Collector {
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

	for _, f := range c.cfg.SourceFiles {
		t, err := tail.TailFile(f, tail.Config{
			Follow: true,
			ReOpen: true,
			Poll:   true,
		})

		if err != nil {
			log.Panic(err)
		}

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
		if target.Regexp().MatchString(value) {
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
