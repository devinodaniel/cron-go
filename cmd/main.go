package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/devinodaniel/cron-go/cmd/config"
	"github.com/devinodaniel/cron-go/cmd/monitor"
)

type Cron struct {
	EpochStart int64         `json:"epochStart"`
	EpochEnd   int64         `json:"epochEnd"`
	ExitCode   int           `json:"exitCode"`
	Monitor    Monitor       `json:"monitor"`
	Timeout    time.Duration `json:"timeout"`
	Duration   int64         `json:"duration"`
	Args       []string      `json:"args"`
}

type Monitor struct {
	Prometheus monitor.Prometheus `json:"prometheus"`
	Namespace  string             `json:"namespace"`
	Prefix     string             `json:"prefix"`
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

// start() runs the command and sets the metadata
func (c *Cron) start() error {
	// set the start time
	c.EpochStart = time.Now().Unix()

	// execute the command
	if err := raw_cmd(c.Args); err != nil {
		// set the exit code to 1 (FAIL) if the command failed
		c.ExitCode = 1
		return err
	}

	// if we reached this the command didn't fail so exit code is 0 (SUCCESS)
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
	c.writeMetrics()

	return nil
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

// writeMetrics writes the metrics to a file
func (c *Cron) writeMetrics() {
	// set the namespace
	c.Monitor.Namespace = config.CRON_NAMESPACE

	// if no namespace is provided, generate one from the arguments
	// WARNING: this may cause issues if the arguments are sensitive
	// TODO: add a flag to disable this feature or require a namespace
	if c.Monitor.Namespace == "" {
		c.Monitor.Namespace = strings.Join(c.Args, "_")
	}

	// REPLACE invalid characters in the namespace
	c.Monitor.Namespace = strings.ReplaceAll(c.Monitor.Namespace, ".", "_")  // period->underscore
	c.Monitor.Namespace = strings.ReplaceAll(c.Monitor.Namespace, " ", "_")  // space->underscore
	c.Monitor.Namespace = strings.ReplaceAll(c.Monitor.Namespace, "-", "_")  // hyphen->underscore
	c.Monitor.Namespace = strings.ReplaceAll(c.Monitor.Namespace, "/", "_")  // slash->underscore
	c.Monitor.Namespace = strings.ReplaceAll(c.Monitor.Namespace, "\\", "_") // backslash->underscore
	c.Monitor.Namespace = strings.ReplaceAll(c.Monitor.Namespace, ":", "_")  // colon->underscore
	c.Monitor.Namespace = strings.ReplaceAll(c.Monitor.Namespace, ";", "_")  // semicolon->underscore
	c.Monitor.Namespace = strings.ReplaceAll(c.Monitor.Namespace, ",", "_")  // comma->underscore
	c.Monitor.Namespace = strings.ReplaceAll(c.Monitor.Namespace, "=", "_")  // equal->underscore
	c.Monitor.Namespace = strings.ReplaceAll(c.Monitor.Namespace, "(", "_")  // left_parenthesis->underscore
	c.Monitor.Namespace = strings.ReplaceAll(c.Monitor.Namespace, ")", "_")  // right_parenthesis->underscore
	c.Monitor.Namespace = strings.ReplaceAll(c.Monitor.Namespace, "[", "_")  // left_bracket->underscore
	c.Monitor.Namespace = strings.ReplaceAll(c.Monitor.Namespace, "]", "_")  // right_bracket->underscore
	c.Monitor.Namespace = strings.ReplaceAll(c.Monitor.Namespace, "{", "_")  // left_brace->underscore
	c.Monitor.Namespace = strings.ReplaceAll(c.Monitor.Namespace, "}", "_")  // right_brace->underscore
	c.Monitor.Namespace = strings.ReplaceAll(c.Monitor.Namespace, "<", "_")  // less_than->underscore
	c.Monitor.Namespace = strings.ReplaceAll(c.Monitor.Namespace, ">", "_")  // greater_than->underscore
	c.Monitor.Namespace = strings.ReplaceAll(c.Monitor.Namespace, "|", "_")  // pipe->underscore
	c.Monitor.Namespace = strings.ReplaceAll(c.Monitor.Namespace, "?", "_")  // question_mark->underscore
	c.Monitor.Namespace = strings.ReplaceAll(c.Monitor.Namespace, "*", "_")  // asterisk->underscore
	c.Monitor.Namespace = strings.ReplaceAll(c.Monitor.Namespace, "\"", "_") // double_quote->underscore
	c.Monitor.Namespace = strings.ReplaceAll(c.Monitor.Namespace, "'", "_")  // single_quote->underscore
	c.Monitor.Namespace = strings.ReplaceAll(c.Monitor.Namespace, "`", "_")  // backtick->underscore
	c.Monitor.Namespace = strings.ReplaceAll(c.Monitor.Namespace, "~", "_")  // tilde->underscore
	c.Monitor.Namespace = strings.ReplaceAll(c.Monitor.Namespace, "!", "_")  // exclamation_mark->underscore
	c.Monitor.Namespace = strings.ReplaceAll(c.Monitor.Namespace, "@", "_")  // at_sign->underscore
	c.Monitor.Namespace = strings.ReplaceAll(c.Monitor.Namespace, "#", "_")  // hash->underscore
	c.Monitor.Namespace = strings.ReplaceAll(c.Monitor.Namespace, "$", "_")  // dollar->underscore
	c.Monitor.Namespace = strings.ReplaceAll(c.Monitor.Namespace, "%", "_")  // percent->underscore
	c.Monitor.Namespace = strings.ReplaceAll(c.Monitor.Namespace, "^", "_")  // caret->underscore
	c.Monitor.Namespace = strings.ReplaceAll(c.Monitor.Namespace, "&", "_")  // ampersand->underscore
	c.Monitor.Namespace = strings.ReplaceAll(c.Monitor.Namespace, "+", "_")  // plus->underscore
	c.Monitor.Namespace = strings.ReplaceAll(c.Monitor.Namespace, "-", "_")  // minus->underscore

	// remove any leading or trailing underscores
	c.Monitor.Namespace = strings.Trim(c.Monitor.Namespace, "_")

	// convert namespace to lowercase
	c.Monitor.Namespace = strings.ToLower(c.Monitor.Namespace)

	// set the prefix, if provided
	c.Monitor.Prefix = config.CRON_METRICS_PREFIX
	if c.Monitor.Prefix != "" {
		c.Monitor.Prefix = strings.ToLower(c.Monitor.Prefix) + "_"
	}

	// PROMETHEUS METRICS
	c.Monitor.Prometheus = monitor.Prometheus{
		Namespace: c.Monitor.Namespace,
		Prefix:    c.Monitor.Prefix,
		Metrics: []monitor.Metric{
			{
				Name:  "cron_start_seconds",
				Help:  "Start time of cronjob last run",
				Type:  "counter",
				Value: int(c.EpochStart),
			},
			{
				Name:  "cron_end_seconds",
				Help:  "End time of cronjob last run",
				Type:  "counter",
				Value: int(c.EpochEnd),
			},
			{
				Name:  "cron_exit_code",
				Help:  "Exit code of cronjob last run",
				Type:  "gauge",
				Value: c.ExitCode,
			},
			{
				Name:  "cron_duration_seconds",
				Help:  "Duration of cronjob last run",
				Type:  "gauge",
				Value: int(c.Duration),
			},
		},
	}

	// write the metrics to a file
	c.Monitor.Prometheus.WriteMetrics()
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
