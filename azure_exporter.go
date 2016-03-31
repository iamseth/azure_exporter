package main

import (
	"flag"
	"net/http"
	"strings"
	"sync"

	"github.com/iamseth/azure_exporter/azure"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

const (
	namespace = "azure" // Prefix namespace for Prometheus metrics.
)

var (
	// Version of azure_exporter. Set at build time.
	Version = "0.0.0.dev"

	listenAddress = flag.String("listen-address", ":9080", "The address to listen on for HTTP requests.")
	metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	credentials   = flag.String("credentials-file", getHomeDirectory()+"/.azure/credentials.json", "Specify the JSON file with the Azure credentials.")

	labels = []string{"name", "group"}
)

// Exporter implements prometheus.Collector.
type Exporter struct {
	azureClient               *azure.Client
	mutex                     sync.RWMutex
	up                        prometheus.Gauge
	vpnConnectionStatus       *prometheus.GaugeVec
	vpnConnectionIngressBytes *prometheus.CounterVec
	vpnConnectionEgressBytes  *prometheus.CounterVec
}

// NewAzureExporter returns an initialized Azure Exporter.
func NewAzureExporter(credsPath string) *Exporter {
	c := azure.NewCredentialsFromFile(credsPath)
	client, err := azure.NewClient(c)
	if err != nil {
		log.Fatalf("Unable to log into Azure Resource Manager: %s", err)
	}

	return &Exporter{
		azureClient: client,
		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "up",
			Help:      "Was the last scrape of azure successful.",
		}),
		vpnConnectionStatus:       newGuageVec("vpn_status", "Status of the VPN connection.", labels),
		vpnConnectionEgressBytes:  newCounterVec("vpn_egress_bytes", "Outbound bytes transferred", labels),
		vpnConnectionIngressBytes: newCounterVec("vpn_ingress_bytes", "Inbound bytes transferred", labels),
	}
}

// Describe describes all the metrics ever exported by the HAProxy exporter. It
// implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	e.vpnConnectionStatus.Describe(ch)
	e.vpnConnectionEgressBytes.Describe(ch)
	e.up.Describe(ch)
}

func (e *Exporter) scrapeVPNConnections() error {
	// capture errors and actually return
	conns, err := e.azureClient.FindVPNConnections()
	if err != nil {
		log.Errorf("Unable to find connections in Azure: %s", err)
	}

	var wg sync.WaitGroup
	for _, conn := range conns {
		wg.Add(1)
		defer wg.Done()
		go func(conn azure.VPNConnection) {
			connection, err := e.azureClient.FindVPNConnection(conn.ResourceGroup, conn.Name)
			if err != nil {
				log.Errorf("Unable to retrieve VPN connection %s from Azure: %s", conn.Name, err)
			}
			name := strings.ToLower(conn.Name)
			rgroup := strings.ToLower(conn.ResourceGroup)
			e.vpnConnectionIngressBytes.WithLabelValues(name, rgroup).Add(connection.Properties.IngressBytesTransferred)
			e.vpnConnectionEgressBytes.WithLabelValues(name, rgroup).Add(connection.Properties.EgressBytesTransferred)
			if connection.Properties.Status == "Connected" {
				e.vpnConnectionStatus.WithLabelValues(name, rgroup).Set(1)
			} else {
				e.vpnConnectionStatus.WithLabelValues(name, rgroup).Set(0)
			}
		}(conn)
	}
	return nil
}

// Collect fetches the stats from Azure and delivers them as Prometheus metrics. It implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.scrapeVPNConnections()
	e.vpnConnectionIngressBytes.Collect(ch)
	e.vpnConnectionEgressBytes.Collect(ch)
	e.vpnConnectionStatus.Collect(ch)
}

func main() {
	flag.Parse()
	exporter := NewAzureExporter(*credentials)
	prometheus.MustRegister(exporter)
	http.Handle(*metricsPath, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
                <head><title>Azure Exporter</title></head>
                <body>
                   <h1>Azure Exporter</h1>
                   <p><a href='` + *metricsPath + `'>Metrics</a></p>
                   </body>
                </html>
              `))
	})
	log.Infof("Starting Server: %s", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
