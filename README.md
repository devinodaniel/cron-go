# cron-runner

A simple cron execution wrapper that allows you to gather information about your crons.

## Usage

Currently, `cron-runner` has only been tested for *nix system and with the Prometheus datasource. If you've found a bug please open an Issue or PR! Seriously, that'd be cool.

Premise: You will never need to rewrite your crons or change them to collect data about them. Simply add this to the cron and that's it. Keep going for more info on how to use.

### Install - Build the binary

```bash 
go build -v -o cron-runner ./cmd
```

### Without cron

After you've built the binary (above)... technically, this doesn't have to be run with a cron. You can execute any command to create metrics about it.

```
$ ./cron-runner sleep 1
$ cat cron_sleep_1_metrics.prom  

# HELP cron_start_time_seconds Start time of cronjob last run (epoch)
# TYPE cron_start_time_seconds gauge
cron_start_time_seconds{namespace="sleep_1"} 1740180749
# HELP cron_end_time_seconds End time of cronjob last run (epoch)
# TYPE cron_end_time_seconds gauge
cron_end_time_seconds{namespace="sleep_1"} 1740180750
# HELP cron_status_code Status code of cronjob last run
# TYPE cron_status_code gauge
cron_status_code{namespace="sleep_1"} 0
# HELP cron_exit_code Exit code of cronjob last run
# TYPE cron_exit_code gauge
cron_exit_code{namespace="sleep_1"} 0
# HELP cron_duration_milliseconds Duration of cronjob last run (milliseconds)
# TYPE cron_duration_milliseconds gauge
cron_duration_milliseconds{namespace="sleep_1"} 1017
# HELP cron_timeout_seconds Timeout of cronjob
# TYPE cron_timeout_seconds gauge
cron_timeout_seconds{namespace="sleep_1"} 86400
# HELP cron_dryrun Dryrun mode
# TYPE cron_dryrun gauge
cron_dryrun{namespace="sleep_1"} 0
```

### With cron

```bash
* * * * * ./cron-runner sleep 1
```

## Migrating your crons

It's simple to start using `cron-runner`. All you need to do is add the binary to the first argument in your cron syntax.

**Before**

```bash
* * * * * /bin/script -d 1 -h 24 --some-other-flag test
```

**After**

```bash
* * * * * CRON_METRICS_DIR=./ ./cron-runner /bin/script -d 1 -h 24 --some-other-flag test
```

## Config

These options can be passed to `cron-runner` per command or set as global environment variables.
| Option               | Description                                                                                       | Default                                      |
|----------------------|---------------------------------------------------------------------------------------------------|----------------------------------------------|
| `CRON_TIMEOUT`       | Kills the running process after the specified number of seconds                                   | 86400 (24 hours)                                       |
| `CRON_NAMESPACE`     | Sets the namespace for the metric. If not supplied, a name will be generated                      | None, empty                                  |
| `CRON_DRYRUN`        | If set to true, skips executing the cron commands and prints the arguments                        | False                                        |
| `CRON_METRICS`       | Set to false to turn off the creation of the metrics file                                         | True                                         |
| `CRON_METRICS_PREFIX`| Sets the prefix for the Prometheus metrics name                                                   | None, empty (generated)                      |
| `CRON_METRICS_DIR`   | Directory to save the metrics files. This is the default Node Exporter directory                  | `/var/lib/node_exporter/textfile_collector`  |


## Metrics
This software currently emits Prometheus metrics to a `.prom` file. Currently, duration and exit code are the primary metrics. The file created is a [Node Exporter][node-exporter], and more specifically, [Textfile Collector][text-collector] scrapable text file which is just a bunch of time-series metrics. We're making some assumptions that everyone that uses this knows what Prometheus and Node Exporter is. If it's not clear, submit a PR or Issue so we can give more detail!

[prometheus]: https://prometheus.io
[node-exporter]: https://github.com/prometheus/node_exporter?tab=readme-ov-file#textfile-collector
[text-collector]: https://github.com/prometheus/node_exporter?tab=readme-ov-file#textfile-collector

Set `CRON_METRIC_PREFIX` to give a prefix to the metrics name. For example, with `CRON_METRIC_PREFIX=prodcronhost` it would be this:

```
prodcronhost_cron_start_seconds{cronjob_name="$namespace"} $start_time
```

### Exit Code vs Status Code

Exit codes are the codes returned by the underlying script or command. Status codes are the status of cron itself. If a cron succeeds, its `exit_code` is equal to `0 (SUCCESS)` and its `status_code` is also equal to `0 (SUCCESS)`. If a cron fails, and it's not due to a timeout `2 (TIMEOUT)` or termination `3 (TERMINATED)` (think CTRL+C), then its `exit_code` is equal to `1 (FAIL)` or the exit code of the underlying command `(0-255)`, and its `status_code` is equal to `1 (FAIL)`.

This separation makes it easy to determine if crons are failing and why via `status_code`, while the `exit_code` provides more data to help determine failure without having to look into logs.

| Name                  | Status Code |
|-----------------------|-------------|
| CRON_STATUS_UNKNOWN   | -1          |
| CRON_STATUS_SUCCESS   | 0           |
| CRON_STATUS_FAIL      | 1           |
| CRON_STATUS_TIMEOUT   | 2           |
| CRON_STATUS_TERMINATED| 3           |

| Name                         | Exit Code |
|------------------------------|-----------|
| CRON_EXITCODE_UNKNOWN        | -1        |
| CRON_EXITCODE_SUCCESS        | 0         |
| CRON_EXITCODE_FAIL_GENERIC   | 1         |
| CRON_EXITCODE_PERM_DENIED    | 126       |
| CRON_EXITCODE_EXEC_NOT_FOUND | 127       |
| CRON_EXITCODE_SIG_INT        | 130       |
| CRON_EXITCODE_SIG_TERM       | 143       |

## Recommended Alerts

## Security concerns

If a unique namespace is not provided via `CRON_NAMESPACE=<custom_namespace>` one will be generated using the command and arguments in the cron task. If for some reason there is sensitive data in the command or arguments (passwords, hidden file paths, tokens, etc etc...) it may be present in the `namespace` label of the outputted metrics file. Run your cron with `CRON_DRYRUN=true` to verify the namespace that will be generated. Set a safe namespace to eliminate this concern.

## Tests

Tests outline the true behavior and should be referenced to understand more throughly how things work. To run the tests, run `go test -v ./.../`.
