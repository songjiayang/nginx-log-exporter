package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/songjiayang/nginx-log-exporter/collector"
	"github.com/songjiayang/nginx-log-exporter/config"
)

var (
	bind, configFile   string
	placeholderReplace bool
	pollLogInterval    time.Duration
)

func main() {
	flag.StringVar(&bind, "web.listen-address", ":9999", "Address to listen on for the web interface and API.")
	flag.StringVar(&configFile, "config.file", "config.yml", "Nginx log exporter configuration file name.")
	flag.BoolVar(&placeholderReplace, "placeholder.replace", false, "Enable placeholder replacement when rewriting the request path.")
	flag.DurationVar(&pollLogInterval, "poll_log_interval", 0, "Set the interval to find all matched log files for polling; must be positive, or zero to disable polling.  With polling mode, only the files found at mtail startup will be polled.")
	flag.Parse()

	cfg, err := config.LoadFile(configFile)
	if err != nil {
		log.Panic(err)
	}

	var options config.Options
	options.SetPlaceholderReplace(placeholderReplace)
	options.SetPollLogInterval(pollLogInterval)

	for _, app := range cfg.App {
		go collector.NewCollector(app, options).Run()
	}

	fmt.Printf("running HTTP server on address %s\n", bind)

	http.Handle("/metrics", promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		},
	))
	if err := http.ListenAndServe(bind, nil); err != nil {
		log.Fatalf("start server with error: %v\n", err)
	}
}
