package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	listenAddress = flag.String("web.listen-address", ":9042", "Address on which to expose metrics and web interface.")
	metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose exporter's metrics.")
	configFile    = flag.String("config.file", "config.yml", "Path to configuration file.")

	cfg *Config
)

func main() {
	flag.Parse()

	cfg = &Config{}
	err := cfg.Load(*configFile)
	if err != nil {
		fmt.Printf("Can't read configuration file: %s\n", err.Error())
		os.Exit(-1)
	}

	collector, err := newCollector(cfg)
	if err != nil {
		log.Fatal("Can't create", err)
		os.Exit(-1)
	}

	prometheus.MustRegister(collector)

	http.Handle(*metricsPath, promhttp.Handler())

	log.Print("Starting nifcloud exporter")
	log.Print("listen address: ", *listenAddress)
	log.Print("telemetry path: ", *metricsPath)
	log.Print("config file: ", *configFile)

	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
