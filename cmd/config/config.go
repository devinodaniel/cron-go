package config

import (
	"fmt"
	"os"
	"strconv"
)

var (
	CRON_TIMEOUT        = EnvInt("CRON_TIMEOUT", 86400) // 24 hours in seconds
	CRON_NAMESPACE      = EnvStr("CRON_NAMESPACE", "")  // *optional* underlines and lowercase only
	CRON_DRYRUN         bool
	CRON_METRICS        bool
	CRON_METRICS_PREFIX = EnvStr("CRON_METRICS_PREFIX", "")                                       // *optional*
	CRON_METRICS_DIR    = EnvStr("CRON_METRICS_DIR", "/var/lib/node_exporter/textfile_collector") // NO TRAILING SLASH :)
)

func init() {
	var err error
	CRON_DRYRUN, err = EnvBool("CRON_DRYRUN", false) // *optional* true or false
	if err != nil {
		fmt.Printf("Error retrieving CRON_DRYRUN: %v\n", err)
	}
	CRON_METRICS, err = EnvBool("CRON_METRICS", true) // *optional*
	if err != nil {
		fmt.Printf("Error retrieving CRON_METRICS: %v\n", err)
	}
}

// EnvStr retrieves the string value of the environment variable named by the key.
func EnvStr(key, defaultValue string) string {
	if value, found := os.LookupEnv(key); found {
		return value
	}
	return defaultValue
}

// EnvInt retrieves the integer value of the environment variable named by the key.
func EnvInt(key string, defaultValue int) int {
	if value, found := os.LookupEnv(key); found {
		// make sure that the value is an integer
		intValue, _ := strconv.Atoi(value)
		return intValue
	}
	return defaultValue
}

// EnvBool retrieves the boolean value of the environment variable named by the key.
func EnvBool(key string, defaultValue bool) (bool, error) {
	if value, found := os.LookupEnv(key); found {
		switch value {
		case "true":
			return true, nil
		case "false":
			return false, nil
		default:
			return defaultValue, fmt.Errorf("invalid boolean value for %s: %s", key, value)
		}
	}
	return defaultValue, nil
}
