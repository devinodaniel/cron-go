package monitor

import (
	"fmt"
	"os"

	"github.com/devinodaniel/cron-go/cmd/config"
)

type Prometheus struct {
	Namespace string   `json:"namespace"`
	Prefix    string   `json:"prefix"`
	Metric    Metric   `json:"metric"`
	Metrics   []Metric `json:"metrics"`
}

type Metric struct {
	Name  string `json:"name"`
	Help  string `json:"help"`
	Type  string `json:"type"`
	Value int    `json:"value"`
}

func (p *Prometheus) WriteMetrics() {
	var metricLine string

	// generate the metrics string for each metric
	for _, metric := range p.Metrics {
		metricLine += fmt.Sprintf("# HELP %s%s %s\n", p.Prefix, metric.Name, metric.Help)
		metricLine += fmt.Sprintf("# TYPE %s%s %s\n", p.Prefix, metric.Name, metric.Type)
		metricLine += fmt.Sprintf("%s%s{namespace=\"%s\"} %d\n", p.Prefix, metric.Name, p.Namespace, metric.Value)
	}

	// metric write filepath
	metricsFile := fmt.Sprintf(config.CRON_METRICS_DIR+"/cron_%s_metrics.prom", p.Namespace)

	// write metrics to file
	filePath := metricsFile
	if err := os.WriteFile(filePath, []byte(metricLine), 0644); err != nil {
		fmt.Printf("Error writing metrics to file: %v\n", err)
	}
}
