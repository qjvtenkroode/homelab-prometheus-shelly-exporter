package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Shelly struct {
	Tmp struct {
		Temperature float64 `json:"tC"`
	} `json:"tmp"`
	Meters []ShellyMeter `json:"meters"`
    Uptime float64 `json:"uptime"`
}

type ShellyMeter struct {
	Power      float64 `json:"power"`
	TotalPower float64 `json:"total"`
}

type ShellyCollector struct {
	target          string
	powerDesc       *prometheus.Desc
	totalPowerDesc  *prometheus.Desc
	temperatureDesc *prometheus.Desc
	uptimeDesc *prometheus.Desc
}

func (c *ShellyCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.powerDesc
	ch <- c.totalPowerDesc
	ch <- c.temperatureDesc
	ch <- c.uptimeDesc
}

func (c *ShellyCollector) Collect(ch chan<- prometheus.Metric) {
	shellyApiClient := http.Client{Timeout: time.Second * 2}
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/status/", c.target), nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("User-Agent", "prometheus-exporter")

	log.Info(fmt.Sprintf("Retrieving data for scrape on: http://%s/status/", c.target))

	res, getErr := shellyApiClient.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
	}
	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	shelly := Shelly{}
	jsonErr := json.Unmarshal(body, &shelly)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	for _, s := range shelly.Meters {
		ch <- prometheus.MustNewConstMetric(
			c.powerDesc,
			prometheus.GaugeValue,
			s.Power,
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalPowerDesc,
			prometheus.CounterValue,
			s.TotalPower,
		)
	}
    ch <- prometheus.MustNewConstMetric(
        c.temperatureDesc,
        prometheus.GaugeValue,
        shelly.Tmp.Temperature,
    )
    ch <- prometheus.MustNewConstMetric(
        c.uptimeDesc,
        prometheus.GaugeValue,
        shelly.Uptime,
    )
}

func NewShellyCollector(t string) *ShellyCollector {
	return &ShellyCollector{
		target:          t,
		powerDesc:       prometheus.NewDesc("shelly_power_watts", "Current power consumption in watts.", nil, nil),
		totalPowerDesc:  prometheus.NewDesc("shelly_total_power_watts", "Total power consumption in watts since reboot.", nil, nil),
		temperatureDesc: prometheus.NewDesc("shelly_temperature", "Current temperature of shelly device in celcius.", nil, nil),
		uptimeDesc: prometheus.NewDesc("shelly_uptime", "Total uptime of shelly device in seconds.", nil, nil),
	}
}

func probeHandler(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	target := params.Get("target")
	registry := prometheus.NewRegistry()
	c := NewShellyCollector(target)
	registry.MustRegister(c)
	// Delegate http serving to Prometheus client library, which will call collector.Collect.
	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func init() {
}

func main() {
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/probe", func(w http.ResponseWriter, r *http.Request) {
		probeHandler(w, r)
	})
	log.Info("Beginning to listen on port :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
