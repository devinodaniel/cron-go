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
	Name   string            `json:"name"`
	Help   string            `json:"help"`
	Type   string            `json:"type"`
	Value  int               `json:"value"`
	Labels map[string]string `json:"labels"`
}

func (p *Prometheus) WriteMetrics() error {
	var metricLine string

	// generate the metrics string for each metric
	for _, metric := range p.Metrics {
		metricLine += fmt.Sprintf("# HELP %s%s %s\n", p.Prefix, metric.Name, metric.Help)
		metricLine += fmt.Sprintf("# TYPE %s%s %s\n", p.Prefix, metric.Name, metric.Type)

		// generate the labels string for each metric
		var labels string
		for k, v := range metric.Labels {
			labels += fmt.Sprintf("%s=\"%s\",", k, v)
		}
		labels = labels[:len(labels)-1] // remove trailing commagearman

		// full metric line
		// <metric_prefix><metric_name>{<label_name>=<label_value>,...} <value>

		// example:
		// cron_job_duration_seconds{job="job1",status="success"} 0.5
		metricLine += fmt.Sprintf("%s%s{%s} %d\n", p.Prefix, metric.Name, labels, metric.Value)
	}

	// set write filepath
	metricsFile := fmt.Sprintf(config.CRON_METRICS_DIR+"/cron_%s_metrics.prom", p.Namespace)

	// write metrics to file
	filePath := metricsFile
	if err := os.WriteFile(filePath, []byte(metricLine), 0644); err != nil {
		return fmt.Errorf("Error writing metrics to file: %v\n", err)
	}

	return nil
}
