package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/devinodaniel/cron-go/cmd/config"
	"github.com/devinodaniel/cron-go/cmd/monitor"

	"github.com/google/uuid"
)

type Cron struct {
	StartTime  time.Time     `json:"startTime"`
	EndTime    time.Time     `json:"endTime"`
	StatusCode int           `json:"statusCode"` // 0: success, 1: fail, 2: timeout, 3: terminated
	ExitCode   int           `json:"exitCode"`   // command exit code, -1 if not set or unknown
	Monitor    Monitor       `json:"monitor"`
	Timeout    time.Duration `json:"timeout"`
	Duration   time.Duration `json:"duration"`
	Args       []string      `json:"args"`
}

type Monitor struct {
	Prometheus monitor.Prometheus `json:"prometheus"`
	Namespace  string             `json:"namespace"`
	Prefix     string             `json:"prefix"`
}

type StatusCode struct {
	Code int    `json:"code"`
	Name string `json:"name"`
}

type ExitCode struct {
	Code int    `json:"code"`
	Name string `json:"name"`
}

// STATUS CODES

const (
	CRON_STATUS_UNKNOWN    = -1
	CRON_STATUS_SUCCESS    = 0
	CRON_STATUS_FAIL       = 1
	CRON_STATUS_TIMEOUT    = 2
	CRON_STATUS_TERMINATED = 3
	CRON_STATUS_RUNNING    = 4
)

var STATUS_CODES = []StatusCode{
	{CRON_STATUS_UNKNOWN, "UNKNOWN"},
	{CRON_STATUS_SUCCESS, "SUCCESS"},
	{CRON_STATUS_FAIL, "FAIL"},
	{CRON_STATUS_TIMEOUT, "TIMEOUT"},
	{CRON_STATUS_TERMINATED, "TERMINATED"},
	{CRON_STATUS_RUNNING, "RUNNING"},
}

func (c *Cron) SetStatusCode(code int) {
	if _, exists := statusCodetoName[code]; exists {
		c.StatusCode = code
		return
	}
}

func (c *Cron) GetStatusCodeName(code ...int) string {
	statusCode := c.StatusCode
	if len(code) > 0 {
		statusCode = code[0]
	}
	if name, exists := statusCodetoName[statusCode]; exists {
		return name
	}
	return "ERR_INVALID_STATUS_CODE"
}

// EXIT CODES

const (
	CRON_EXITCODE_UNKNOWN      = -1
	CRON_EXITCODE_SUCCESS      = 0
	CRON_EXITCODE_FAIL_GENERIC = 1

	// special exit codes (https://tldp.org/LDP/abs/html/exitcodes.html

	CRON_EXITCODE_PERM_DENIED    = 126
	CRON_EXITCODE_EXEC_NOT_FOUND = 127
	CRON_EXITCODE_SIG_INT        = 130
	CRON_EXITCODE_SIG_TERM       = 143
)

var EXIT_CODES = []ExitCode{
	{CRON_EXITCODE_UNKNOWN, "UNKNOWN"},
	{CRON_EXITCODE_SUCCESS, "SUCCESS"},
	{CRON_EXITCODE_FAIL_GENERIC, "FAIL_GENERIC"},
	{CRON_EXITCODE_PERM_DENIED, "PERM_DENIED"},
	{CRON_EXITCODE_EXEC_NOT_FOUND, "EXEC_NOT_FOUND"},
	{CRON_EXITCODE_SIG_INT, "SIG_INT"},
	{CRON_EXITCODE_SIG_TERM, "SIG_TERM"},
}

func (c *Cron) SetExitCode(code int) {
	if _, exists := exitCodetoName[code]; exists {
		c.ExitCode = code
		return
	}
}

func (c *Cron) GetExitCodeName(code ...int) string {
	exitCode := c.ExitCode
	if len(code) > 0 {
		exitCode = code[0]
	}
	if name, exists := exitCodetoName[exitCode]; exists {
		return name
	}
	return "ERR_INVALID_EXIT_CODE"
}

var statusCodetoName = make(map[int]string)
var statusNametoCode = make(map[string]int)

var exitCodetoName = make(map[int]string)
var exitNametoCode = make(map[string]int)

func init() {
	// initialize status code maps
	for _, s := range STATUS_CODES {
		statusCodetoName[s.Code] = s.Name
		statusNametoCode[s.Name] = s.Code
	}

	// initialize exit code maps
	for _, e := range EXIT_CODES {
		exitCodetoName[e.Code] = e.Name
		exitNametoCode[e.Name] = e.Code
	}
}

// usage prints how to use this little cron runner
func usage() {
	fmt.Println("Usage: cron-runner <any-command-or-script> [args]")
	fmt.Println("Example: CRON_DRYRUN=true cron-runner echo 'hello world'")
	fmt.Println("Example: cron-runner php /path/to/script.php")

	// print the config options
	// these should be set as global environment variables ieL profile
	// or declared inline per cron command
	// CRON_DRYRUN=true ./cron-runner echo 'hello world'
	fmt.Println("\nConfig Options (set as env vars):")
	fmt.Printf("  CRON_TIMEOUT: %d\n", config.CRON_TIMEOUT)
	fmt.Printf("  CRON_METRICS: %t\n", config.CRON_METRICS)
	fmt.Printf("  CRON_METRICS_PREFIX: %s\n", config.CRON_METRICS_PREFIX)
	fmt.Printf("  CRON_NAMESPACE: %s\n", config.CRON_NAMESPACE)
	fmt.Printf("  CRON_DRYRUN: %t\n", config.CRON_DRYRUN)
}

func New(args []string) (*Cron, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("No arguments provided. Nothing to do. Run 'help' for usage.")
	}

	// if 'help' is passed as an argument, print the usage
	if len(args) == 1 && args[0] == "help" {
		usage()
		os.Exit(0)
	}

	return &Cron{
		Args: args,
	}, nil
}

// Run() runs the cron job
func (c *Cron) Run() error {
	// Set up a channel to receive signals
	sigs := make(chan os.Signal, 1)

	// Listen for SIGINT and SIGTERM signals - should we listen for more signals?
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Set up a channel to signal when the cron job is done
	done := make(chan struct{})

	// Run the cron job in a goroutine
	go func() {
		c.start()
		close(done)
	}()

	// Wait for either the cron job to finish or a signal to be received
	select {
	case sig := <-sigs:
		c.terminated(sig)
	case <-done:
		// Cron job completed successfully
	}

	// Ensure finish is called regardless of how the cron job ends
	if err := c.finish(); nil != err {
		return err
	}

	return nil
}

// start() runs the command and sets the metadata
// we dont return an error here because we just want to know if the command failed
// which is set in the metadata
func (c *Cron) start() {
	// set the start metadata
	c.StartTime = time.Now()
	c.StatusCode = CRON_STATUS_RUNNING
	c.ExitCode = CRON_EXITCODE_UNKNOWN

	// set namespace
	c.setNamespace()

	// set prefix
	c.setMetricPrefix()

	// if dryrun is enabled, print the args, metrics and exit
	if config.CRON_DRYRUN {
		if c.Monitor.Prefix != "" {
			fmt.Printf("DRYRUN: Metric Prefix: %s\n", c.Monitor.Prefix)
		}
		fmt.Printf("DRYRUN: Metric Namespace: %s\n", c.Monitor.Namespace)
		fmt.Printf("DRYRUN: Args: %v\n", c.Args)
		fmt.Printf("DRYRUN: Timeout: %v\n", config.CRON_TIMEOUT)
		return
	}

	// write metrics file so we know its running
	if config.CRON_METRICS {
		c.Monitor.Prometheus = monitor.Prometheus{
			Namespace: c.Monitor.Namespace,
			Prefix:    c.Monitor.Prefix,
			Metrics: []monitor.Metric{
				{
					Name:   "cron_start_time_seconds",
					Help:   "Start time of cronjob last run (epoch)",
					Type:   "gauge",
					Value:  int(c.StartTime.Unix()),
					Labels: map[string]string{"namespace": c.Monitor.Namespace},
				},
				{
					Name:   "cron_status_code",
					Help:   "Status code of cronjob last run",
					Type:   "gauge",
					Value:  c.StatusCode,
					Labels: map[string]string{"namespace": c.Monitor.Namespace},
				},
				{
					Name:   "cron_timeout_seconds",
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

		c.writeMetrics()
	}

	// execute the command and get the exit code
	c.ExitCode, c.StatusCode = run_cmd(c.Args)
}

// terminated() updates the metadata after the command has been terminated
func (c *Cron) terminated(sig os.Signal) {
	c.StatusCode = CRON_STATUS_TERMINATED

	switch sig {
	case syscall.SIGINT:
		c.ExitCode = CRON_EXITCODE_SIG_INT
	case syscall.SIGTERM:
		c.ExitCode = CRON_EXITCODE_SIG_TERM
	default:
		c.ExitCode = CRON_EXITCODE_UNKNOWN
	}
}

// finish() updates the metadata after the command has executed
// we return an error here because we want to know if there was an error writing the metrics
func (c *Cron) finish() error {
	// set the end time
	c.EndTime = time.Now()

	// calculate the duration
	c.Duration = c.EndTime.Sub(c.StartTime)

	if config.CRON_METRICS {
		// rewrite all metrics
		c.Monitor.Prometheus.Metrics = []monitor.Metric{
			{
				Name:   "cron_start_time_seconds",
				Help:   "Start time of cronjob last run (epoch)",
				Type:   "gauge",
				Value:  int(c.StartTime.Unix()),
				Labels: map[string]string{"namespace": c.Monitor.Namespace},
			},
			{
				Name:   "cron_status_code",
				Help:   "Status code of cronjob last run",
				Type:   "gauge",
				Value:  c.StatusCode,
				Labels: map[string]string{"namespace": c.Monitor.Namespace},
			},
			{
				Name:   "cron_timeout_seconds",
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
			{
				Name:   "cron_end_time_seconds",
				Help:   "End time of cronjob last run (epoch)",
				Type:   "gauge",
				Value:  int(c.EndTime.Unix()),
				Labels: map[string]string{"namespace": c.Monitor.Namespace},
			},
			{
				Name:   "cron_status_code",
				Help:   "Status code of cronjob last run",
				Type:   "gauge",
				Value:  c.StatusCode,
				Labels: map[string]string{"namespace": c.Monitor.Namespace},
			},
			{
				Name:   "cron_exit_code",
				Help:   "Exit code of cronjob command last run",
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
		}

		if err := c.writeMetrics(); nil != err {
			return err
		}
	}

	return nil
}

// run_cmd() executes the cron job and returns the exit code and status code
func run_cmd(args []string) (int, int) {
	// create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(config.CRON_TIMEOUT)*time.Second)
	defer cancel()

	// config the command with context
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)

	// redirect stdout to os.Stdout
	cmd.Stdout = os.Stdout

	// redirect stderr to os.Stderr
	cmd.Stderr = os.Stderr

	// run it!
	if err := cmd.Run(); err != nil {
		// check if the context deadline was exceeded
		if ctx.Err() == context.DeadlineExceeded {
			return CRON_EXITCODE_FAIL_GENERIC, CRON_STATUS_TIMEOUT
		}

		// we do specific circumstance failure checking here
		// to be able to return the desired corresponding exit code
		// maybe there is a better way to do this but it works for now

		// 127: command not found
		if strings.Contains(err.Error(), "executable file not found in $PATH") {
			return CRON_EXITCODE_EXEC_NOT_FOUND, CRON_STATUS_FAIL
		}

		// 126: permission denied
		if strings.Contains(err.Error(), "permission denied") {
			return CRON_EXITCODE_PERM_DENIED, CRON_STATUS_FAIL
		}

		// check if the command failed for any other reason
		if exitError, ok := err.(*exec.ExitError); ok {
			status, ok := exitError.Sys().(syscall.WaitStatus)
			if ok {
				exitCode := status.ExitStatus()
				return exitCode, CRON_STATUS_FAIL
			}
		}

		// if we reached this point, the command failed for an unknown reason
		return CRON_EXITCODE_UNKNOWN, CRON_STATUS_FAIL
	}

	// if we reached this point, the command was successful
	return CRON_EXITCODE_SUCCESS, CRON_STATUS_SUCCESS
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
		// but the downside is that the metrics will be harder to track because
		// the namespace will be different for each run
		randomID := uuid.New()
		c.Monitor.Namespace = "randomid_" + randomID.String()
		fmt.Printf("Invalid namespace: generated a randomid: %s", c.Monitor.Namespace)
	}
}

func (c *Cron) setMetricPrefix() {
	// is there an env var for the prefix?
	if prefix := os.Getenv("CRON_METRICS_PREFIX"); prefix != "" {
		c.Monitor.Prefix = prefix
	}

	// set the prefix, if provided
	c.Monitor.Prefix = config.CRON_METRICS_PREFIX

	// convert the prefix to lowercase
	if c.Monitor.Prefix != "" {
		c.Monitor.Prefix = strings.ToLower(c.Monitor.Prefix) + "_"
	}
}

// writeMetrics writes the metrics to a file
func (c *Cron) writeMetrics() error {
	// always write metrics all cron statuses
	for _, status := range STATUS_CODES {
		c.Monitor.Prometheus.Metrics = append(c.Monitor.Prometheus.Metrics, monitor.Metric{
			Name:  fmt.Sprintf("cron_status"),
			Help:  fmt.Sprintf("Cron status for %s", c.GetStatusCodeName(status.Code)),
			Type:  "gauge",
			Value: boolToInt(c.StatusCode == status.Code),
			Labels: map[string]string{
				"namespace": c.Monitor.Namespace,
				"status":    c.GetStatusCodeName(status.Code),
				"code":      fmt.Sprintf("%d", status.Code),
			},
		})
	}

	// always write metrics all cron statuses
	for _, exit := range EXIT_CODES {
		c.Monitor.Prometheus.Metrics = append(c.Monitor.Prometheus.Metrics, monitor.Metric{
			Name:  fmt.Sprintf("cron_exit"),
			Help:  fmt.Sprintf("Cron exit for %s", c.GetExitCodeName(exit.Code)),
			Type:  "gauge",
			Value: boolToInt(c.ExitCode == exit.Code),
			Labels: map[string]string{
				"namespace": c.Monitor.Namespace,
				"exit":      c.GetExitCodeName(exit.Code),
				"code":      fmt.Sprintf("%d", exit.Code),
			},
		})
	}

	// write Prometheus metrics to a file
	if err := c.Monitor.Prometheus.WriteMetrics(); nil != err {
		return err
	}

	return nil
}

// boolToInt converts a boolean to an integer aka true -> 1, false -> 0
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func main() {
	// set umask to 022
	// this is to ensure that the files created by the script are not world writable
	// but are readable by others
	oldUmask := syscall.Umask(022)
	defer syscall.Umask(oldUmask)

	// get the arguments passed to the script
	args := os.Args[1:]

	// create a new cron object for keeping track of metadata
	cron, err := New(args)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}

	if err := cron.Run(); nil != err {
		fmt.Printf("ERROR: %v\n", err)
	}
}
