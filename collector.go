package main

import (
	"sync"
	"time"
	"unicode"

	"github.com/heriet/funicula/nifcloud"
	"github.com/heriet/funicula/nifcloud/credential"
	"github.com/heriet/funicula/service/rdb"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	rdbMetricNames = []string{
		"BinLogDiskUsage",
		"CPUUtilization",
		"DatabaseConnections",
		"DiskQueueDepth",
		"FreeableMemory",
		"FreeStorageSpace",
		"ReplicaLag",
		"SwapUsage",
		"ReadIOPS",
		"WriteIOPS",
		"ReadThroughput",
		"WriteThroughput",
	}
)

type metric struct {
	Name string
	Desc *prometheus.Desc
}

type rdbCollector struct {
	env    *RDBEnv
	client *rdb.RDB
}

type nifcloudCollector struct {
	rdbCollectors []rdbCollector
	metrics       []metric

	scrapeTime     prometheus.Gauge
	scrapeFailures prometheus.Counter
	totalRequests  prometheus.Counter
}

func newCollector(cfg *Config) (*nifcloudCollector, error) {
	collector := &nifcloudCollector{
		rdbCollectors: generateRdbCollectors(cfg),
		metrics:       generateMetrics(),
		scrapeTime: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "nifcloud_scrape_duration_seconds",
			Help: "Time this NIFCLOUD scrape took, in seconds.",
		}),
		scrapeFailures: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "nifcloud_failure_requests",
			Help: "The number of failure request made by this scrape.",
		}),
		totalRequests: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "nifcloud_requests_total",
			Help: "API requests made to NIFCLOUD",
		}),
	}
	return collector, nil
}

func generateRdbCollectors(cfg *Config) []rdbCollector {
	collectors := make([]rdbCollector, len(cfg.RDBEnv))
	for i, rdbEnv := range cfg.RDBEnv {
		collectors[i].env = &rdbEnv
		collectors[i].client = newRdbClient(collectors[i].env)
	}
	return collectors
}

func newRdbClient(env *RDBEnv) *rdb.RDB {
	nifCfg := &nifcloud.Config{
		Region: env.Region,
		Credential: &credential.Credential{
			AccessKeyId:     env.AccessKeyId,
			SecretAccessKey: env.SecretAccessKey,
		},
	}
	svc := rdb.New(nifCfg)
	return svc
}

func generateMetrics() []metric {
	ms := make([]metric, len(rdbMetricNames))
	for i, name := range rdbMetricNames {
		ms[i].Name = name
		ms[i].Desc = prometheus.NewDesc(
			"nifcloud_rdb_" + toSnakeCase(name),
			name,
			[]string{
				"env",
				"region",
				"db_instance",
			},
			nil,
		)
	}

	return ms
}

func scrapeRdb(c *nifcloudCollector, rc rdbCollector, ins Instance, ch chan<- prometheus.Metric) {
	labels := []string{}
	labels = append(labels, rc.env.Name, rc.env.Region, ins.Name)

	before1minute := time.Now().Add(-time.Minute).UTC()
	startTime := before1minute.Format("2006-1-2 15:04")

	for _, met := range c.metrics {
		scrapeRdbMetric(c, rc, ins, ch, startTime, labels, met)
	}	
}

func scrapeRdbMetric(c *nifcloudCollector, rc rdbCollector, ins Instance, ch chan<- prometheus.Metric, startTime string, labels []string, met metric) {

	params := &rdb.NiftyGetMetricStatisticsInput{
        Dimensions: []rdb.Dimension{
            {
                Name: "DBInstanceIdentifier",
                Value: ins.Name,
            },
		},
		StartTime: startTime,
        MetricName: met.Name,
	}

	c.totalRequests.Inc()
	output, err := rc.client.NiftyGetMetricStatistics(params)
	if err != nil {
		c.scrapeFailures.Inc()
		return
	}

	datapoints := output.Datapoints
	if len(datapoints) == 0 {
		return
	}

	dataValue := datapoints[0].Sum
	ch <- prometheus.MustNewConstMetric(met.Desc, prometheus.GaugeValue, dataValue, labels...)
}

func (c *nifcloudCollector) Collect(ch chan<- prometheus.Metric) {
	now := time.Now()
	wg := &sync.WaitGroup{}

	for _, rc := range c.rdbCollectors {
		for _, ins := range rc.env.Instances {
			wg.Add(1)

			go func(rc rdbCollector, ins Instance) {
				scrapeRdb(c, rc, ins, ch)
				wg.Done()
			}(rc, ins)

		}
	}

	wg.Wait()
	c.scrapeTime.Set(time.Since(now).Seconds())

	ch <- c.scrapeTime
	ch <- c.scrapeFailures
	ch <- c.totalRequests
}

func (c *nifcloudCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.scrapeTime.Desc()
	ch <- c.scrapeFailures.Desc()
	ch <- c.totalRequests.Desc()

	for _, m := range c.metrics {
		ch <- m.Desc
	}
}

func toSnakeCase(in string) string {
	runes := []rune(in)
	length := len(runes)

	var out []rune
	for i := 0; i < length; i++ {
		if i > 0 && unicode.IsUpper(runes[i]) && ((i+1 < length && unicode.IsLower(runes[i+1])) || unicode.IsLower(runes[i-1])) {
			out = append(out, '_')
		}
		out = append(out, unicode.ToLower(runes[i]))
	}

	return string(out)
}