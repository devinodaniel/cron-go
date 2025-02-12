package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/devinodaniel/cron-go/cmd/config"
)

type Cron struct {
	EpochStart   int64         `json:"epochStart"`
	EpochEnd     int64         `json:"epochEnd"`
	Namespace    string        `json:"namespace"`
	MetricPrefix string        `json:"metricPrefix"`
	ExitCode     int           `json:"exitCode"`
	Timeout      time.Duration `json:"timeout"`
	Duration     int64         `json:"duration"`
	Args         []string      `json:"args"`
}

func New(args []string) *Cron {
	return &Cron{
		Args: args,
	}
}

// Run() runs the cron job
func (c *Cron) Run() error {
	// start the cron job
	if err := c.start(); err != nil {
		return err
	}

	// finish the cron job
	if err := c.finish(); err != nil {
		return err
	}

	return nil
}

// start() runs the command and sets the metadata if the command is not already running
func (c *Cron) start() error {
	// set the start time
	c.EpochStart = time.Now().Unix()

	// execute the command
	if err := raw_cmd(c.Args); err != nil {
		c.ExitCode = 1
		return err
	}

	// set the exit code
	c.ExitCode = 0

	return nil
}

// finish() updates the metadata after the command has executed
func (c *Cron) finish() error {
	// set the end time
	c.EpochEnd = time.Now().Unix()

	// calculate the duration
	c.Duration = c.EpochEnd - c.EpochStart

	// write the metrics
	writeMetrics(c)

	return nil
}

// raw_cmd() executes the cron commands or scripts and returns an error if it fails
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

func writeMetrics(c *Cron) {
	// write a file containing the metrics
	// get the namespace
	namespace := config.CRON_NAMESPACE
	if namespace == "" {
		// generate a namespace from the arguments
		namespace = strings.Join(c.Args, "_")
	}

	// replace any invalid characters
	namespace = strings.ReplaceAll(namespace, " ", "_")
	namespace = strings.ReplaceAll(namespace, ".", "")
	namespace = strings.ReplaceAll(namespace, "-", "_")
	// convert to lowercase
	namespace = strings.ToLower(namespace)

	metricPrefix := config.CRON_METRIC_PREFIX
	if metricPrefix != "" {
		metricPrefix = strings.ToLower(metricPrefix) + "_"
	}

	metrics := fmt.Sprintf(`
# HELP %scron_start_seconds Start time of cronjob last run
# TYPE %scron_start_seconds counter
%scron_start_seconds{cronjob_name="%s"} %d
	`,
		metricPrefix, metricPrefix, metricPrefix, namespace, c.EpochStart,
	)

	metrics += fmt.Sprintf(`
# HELP %scron_end_seconds End time of cronjob last run
# TYPE %scron_end_seconds counter
%scron_end_seconds{cronjob_name="%s"} %d
	`,
		metricPrefix, metricPrefix, metricPrefix, namespace, c.EpochEnd,
	)

	metrics += fmt.Sprintf(`
# HELP %s_cron_exit_code Exit code of cronjob last run
# TYPE %scron_exit_code gauge
%scron_exit_code{cronjob_name="%s"} %d
	`,
		metricPrefix, metricPrefix, metricPrefix, namespace, c.ExitCode,
	)

	metrics += fmt.Sprintf(`
# HELP %scron_duration_seconds Duration of cronjob last run
# TYPE %scron_duration_seconds counter
%scron_duration_seconds{cronjob_name="%s"} %d
	`,
		metricPrefix, metricPrefix, metricPrefix, namespace, c.Duration,
	)

	// write to file
	filePath := fmt.Sprintf(config.CRON_METRIC_DIR+"/cron_%s_metrics.prom", namespace)
	if err := os.WriteFile(filePath, []byte(metrics), 0644); err != nil {
		fmt.Printf("Error writing metrics to file: %v\n", err)
	}
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
