# cron-go

A simple cron execution wrapper that allows you to gather information about your crons.

## How to use

### Without cron

Technically, this doens't have to be run with a cron. You can execute any command to get information about it.

```
./cron sleep 60
```

### With cron

```bash
* * * * * cron sleep 60
```

## Migrating your crons

It's simple to start using `cron`. All you need to do is add the binary to the first argument in your cron syntax.

**Before**

```bash
* * * * * /bin/script -d 1 -h 24 -some-other-flag test
```

**After**

```bash
* * * * * CRON_TIMEOUT=60 cron /bin/script -d 1 -h 24 -some-other-flag test
```

## Config

These options can be passed to `cron` per command or set as global environment variables.

`CRON_TIMEOUT` (seconds, default: 86400) kills the running process

`CRON_METRICS` (default: true) set to false to turn off the creation of the metrics file

`CRON_METRICS_PREFIX` (default: none, empty) sets the prefix for the prometheus metrics name

`CRON_METRICS_DIR` (default: /var/lib/node_exporter/textfile_collector/) directry to save the metrics files

`CRON_NAMESPACE` (default: none, empty) sets the namespace for the metric. if one is not supplied then a name will be generated

## Metrics

Metrics are emitted as prometheus scrapable text files. 

```
# HELP cron_start_seconds Start time of cronjob last run
# TYPE cron_start_seconds counter
cron_start_seconds{cronjob_name="$namespace"} $start_time

# HELP cron_end_seconds End time of cronjob last run
# TYPE cron_end_seconds counter
cron_end_seconds{cronjob_name="$namespace"} $end_time

# HELP cron_exit_code Exit code of cronjob last run
# TYPE cron_exit_code gauge
cron_exit_code{cronjob_name="$namespace"} $exit_code
```

Set `CRON_METRIC_PREFIX` to give a prefix to the metrics name. With `CRON_METRIC_PREFIX=cronhost`:

```
cronhost_cron_start_seconds{cronjob_name="$namespace"} $start_time
```