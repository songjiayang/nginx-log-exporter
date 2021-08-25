package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/songjiayang/nginx-log-exporter/collector"
	"github.com/songjiayang/nginx-log-exporter/config"
)

var (
	bind, configFile string
)

func main() {
	flag.StringVar(&bind, "web.listen-address", ":9999", "Address to listen on for the web interface and API.")
	flag.StringVar(&configFile, "config.file", "config.yml", "Nginx log exporter configuration file name.")
	flag.Parse()

	cfg, err := config.LoadFile(configFile)
	if err != nil {
		log.Panic(err)
	}

	for _, app := range cfg.App {
		go collector.NewCollector(app).Run()
	}

	fmt.Printf("running HTTP server on address %s\n", bind)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(bind, nil)
}
