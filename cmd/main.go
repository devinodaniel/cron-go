package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/devinodaniel/cron-go/cmd/config"
	"github.com/devinodaniel/cron-go/cmd/monitor"

	"github.com/google/uuid"
)

type Cron struct {
	StartTime time.Time     `json:"startTime"`
	EndTime   time.Time     `json:"endTime"`
	ExitCode  int           `json:"exitCode"`
	Monitor   Monitor       `json:"monitor"`
	Timeout   time.Duration `json:"timeout"`
	Duration  time.Duration `json:"duration"`
	Args      []string      `json:"args"`
}

type Monitor struct {
	Prometheus monitor.Prometheus `json:"prometheus"`
	Namespace  string             `json:"namespace"`
	Prefix     string             `json:"prefix"`
}

const (
	CRON_SUCCESS = 0
	CRON_FAIL    = 1
	CRON_TIMEOUT = 2
)

func New(args []string) *Cron {
	return &Cron{
		Args: args,
	}
}

// Run() runs the cron job
func (c *Cron) Run() error {
	c.start()

	c.finish()

	return nil
}

// start() runs the command and sets the metadata
func (c *Cron) start() {
	// set the start time
	c.StartTime = time.Now()

	// set namespace
	c.setNamespace()

	// set prefix
	c.setMetricPrefix()

	if config.CRON_DRYRUN {
		if c.Monitor.Prefix != "" {
			fmt.Printf("DRYRUN: Metric Prefix: %s\n", c.Monitor.Prefix)
		}
		fmt.Printf("DRYRUN: Metric Namespace: %s\n", c.Monitor.Namespace)
		fmt.Printf("DRYRUN: Args: %v\n", c.Args)
		fmt.Printf("DRYRUN: Timeout: %v\n", config.CRON_TIMEOUT)
		return
	} else {
		// execute the command
		if err := raw_cmd(c.Args); err != nil {
			// if timeout, set the exit code to 2 (TIMEOUT)
			if err.Error() == "TIMEOUT" {
				c.ExitCode = CRON_TIMEOUT
				return
			}

			// set the exit code to 1 (FAIL) if the command failed
			c.ExitCode = CRON_FAIL
			return
		}
	}
	// if we reached this the command didn't fail so exit code is 0 (SUCCESS)
	c.ExitCode = CRON_SUCCESS
}

// finish() updates the metadata after the command has executed
func (c *Cron) finish() {
	// set the end time
	c.EndTime = time.Now()

	// calculate the duration
	c.Duration = c.EndTime.Sub(c.StartTime)

	// write the metrics
	if config.CRON_METRICS {
		c.writeMetrics()
	}
}

// raw_cmd() executes the cron job and returns an error if it fails. most failures are due to timeouts
func raw_cmd(args []string) error {
	// create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(config.CRON_TIMEOUT)*time.Second)
	defer cancel()

	// run the command with context
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		// check if the context deadline was exceeded
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("TIMEOUT")
		}
		return err
	}

	return nil
}

func (c *Cron) setNamespace() {
	c.Monitor.Namespace = config.CRON_NAMESPACE

	// if no namespace is provided, generate one from the arguments
	// WARNING: this may cause issues if the arguments are sensitive
	// TODO: add a flag to disable this feature or require a namespace
	if c.Monitor.Namespace == "" {
		c.Monitor.Namespace = strings.Join(c.Args, "_")
	}

	// Replace invalid characters in the namespace
	replacements := []string{".", " ", "-", "/", "\\", ":", ";", ",", "=", "(", ")", "[", "]",
		"{", "}", "<", ">", "|", "?", "*", "\"", "'", "`", "~", "!", "@", "#", "$", "%", "^", "&", "+", "-"}
	for _, r := range replacements {
		c.Monitor.Namespace = strings.ReplaceAll(c.Monitor.Namespace, r, "_")
	}

	// remove leading and trailing underscores
	c.Monitor.Namespace = strings.Trim(c.Monitor.Namespace, "_")

	// convert namespace to lowercase
	c.Monitor.Namespace = strings.ToLower(c.Monitor.Namespace)

	// validate that the namespace matches the regex [a-zA-Z_:][a-zA-Z0-9_:]*
	// https://prometheus.io/docs/concepts/data_model/#metric-names-and-labels
	if ok, _ := regexp.MatchString("^[a-zA-Z_:][a-zA-Z0-9_:]*$", c.Monitor.Namespace); !ok {
		// randomly generate a random id for namespace if the provided one is invalid
		// this is a stopgap to prevent a cronjob from failing because of an invalid namespace
		randomID := uuid.New()
		c.Monitor.Namespace = "randomid_" + randomID.String()
		fmt.Printf("Invalid namespace: generated a randomid: %s", c.Monitor.Namespace)
	}
}

func (c *Cron) setMetricPrefix() {
	// set the prefix, if provided
	c.Monitor.Prefix = config.CRON_METRICS_PREFIX
	if c.Monitor.Prefix != "" {
		c.Monitor.Prefix = strings.ToLower(c.Monitor.Prefix) + "_"
	}
}

// writeMetrics writes the metrics to a file
func (c *Cron) writeMetrics() {
	// PROMETHEUS METRICS
	c.Monitor.Prometheus = monitor.Prometheus{
		Namespace: c.Monitor.Namespace,
		Prefix:    c.Monitor.Prefix,
		Metrics: []monitor.Metric{
			{
				Name:   "cron_start_time",
				Help:   "Start time of cronjob last run (epoch)",
				Type:   "gauge",
				Value:  int(c.StartTime.Unix()),
				Labels: map[string]string{"namespace": c.Monitor.Namespace},
			},
			{
				Name:   "cron_end_time",
				Help:   "End time of cronjob last run (epoch)",
				Type:   "gauge",
				Value:  int(c.EndTime.Unix()),
				Labels: map[string]string{"namespace": c.Monitor.Namespace},
			},
			{
				Name:   "cron_exit_code",
				Help:   "Exit code of cronjob last run",
				Type:   "gauge",
				Value:  c.ExitCode,
				Labels: map[string]string{"namespace": c.Monitor.Namespace},
			},
			{
				Name:   "cron_duration_milliseconds",
				Help:   "Duration of cronjob last run (milliseconds)",
				Type:   "gauge",
				Value:  int(c.Duration.Milliseconds()),
				Labels: map[string]string{"namespace": c.Monitor.Namespace},
			},
			{
				Name:   "cron_timeout",
				Help:   "Timeout of cronjob",
				Type:   "gauge",
				Value:  config.CRON_TIMEOUT,
				Labels: map[string]string{"namespace": c.Monitor.Namespace},
			},
			{
				Name:   "cron_dryrun",
				Help:   "Dryrun mode",
				Type:   "gauge",
				Value:  boolToInt(config.CRON_DRYRUN),
				Labels: map[string]string{"namespace": c.Monitor.Namespace},
			},
		},
	}

	// write the metrics to a file
	c.Monitor.Prometheus.WriteMetrics()
}

// boolToInt converts a boolean to an integer
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func main() {
	// get the arguments passed to the script
	args := os.Args[1:]

	// Check if any arguments were provided
	if len(args) == 0 {
		fmt.Println("ERROR: No arguments provided. Nothing to do.")
		return
	}

	// create a new cron object for keeping track of metadata
	cron := New(args)

	// run the cron job with command
	if err := cron.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
