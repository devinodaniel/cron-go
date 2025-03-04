package monitor

import (
	"fmt"
	"os"

	"github.com/devinodaniel/cron-go/cmd/config"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/common/expfmt"

	io_prometheus_client "github.com/prometheus/client_model/go"
)

type Prometheus struct {
	Namespace string `json:"namespace"`
	Prefix    string `json:"prefix"`
	Metrics   []string
}

type Metric struct {
	Name   string            `json:"name"`
	Help   string            `json:"help"`
	Type   string            `json:"type"`
	Value  int               `json:"value"`
	Labels map[string]string `json:"labels"`
}

var (
	PrometheusMetricsRegistry = prometheus.NewRegistry()

	CronStartTimeSeconds = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cron_start_time_seconds",
			Help: "Start time of cronjob last run (epoch)",
		}, []string{"namespace"})

	CronEndTimeSeconds = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cron_end_time_seconds",
			Help: "End time of cronjob last run (epoch)",
		},
		[]string{"namespace"})

	CronStatusCode = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cron_status_code",
			Help: "Status code of cronjob last run",
		},
		[]string{"namespace"})

	CronExitCode = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cron_exit_code",
			Help: "Exit code of cronjob command last run",
		},
		[]string{"namespace"})

	CronDurationMilliseconds = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cron_duration_milliseconds",
			Help: "Duration of cronjob last run (milliseconds)",
		},
		[]string{"namespace"})

	CronTimeoutSeconds = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cron_timeout_seconds",
			Help: "Timeout of cronjob",
		},
		[]string{"namespace"})

	CronDryrun = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cron_dryrun",
			Help: "Dryrun mode",
		},
		[]string{"namespace"})

	CronStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cron_status",
			Help: "Status of cronjob last run",
		},
		[]string{"namespace", "code", "status"})

	CronExit = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cron_exit",
			Help: "Exit of cronjob last run",
		},
		[]string{"namespace", "code", "exit"})
)

func (p *Prometheus) WriteMetrics(namespace string, metrics []*io_prometheus_client.MetricFamily) error {
	// set write filepath
	metricsFile := fmt.Sprintf(config.CRON_METRICS_DIR+"/cron_%s_metrics.prom", namespace)

	// Create or open the file
	file, err := os.Create(metricsFile)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Encode metrics in Prometheus text format
	encoder := expfmt.NewEncoder(file, expfmt.NewFormat(expfmt.TypeTextPlain))
	for _, metricFamily := range metrics {
		if err := encoder.Encode(metricFamily); err != nil {
			panic(err)
		}
	}
	return nil
}
