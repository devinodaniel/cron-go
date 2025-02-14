package config

import (
	"log"
	"os"
	"strconv"
)

var (
	CRON_TIMEOUT        = EnvInt("CRON_TIMEOUT", 86400)                                           // 24 hours
	CRON_NAMESPACE      = EnvStr("CRON_NAMESPACE", "")                                            // *optional*
	CRON_DRYRUN         = EnvBool("CRON_DRYRUN", false)                                           // *optional*
	CRON_METRICS        = EnvBool("CRON_METRICS", true)                                           // *optional*
	CRON_METRICS_PREFIX = EnvStr("CRON_METRICS_PREFIX", "")                                       // *optional*
	CRON_METRICS_DIR    = EnvStr("CRON_METRICS_DIR", "/var/lib/node_exporter/textfile_collector") // no trailing slash
)

// EnvString retrieves the string value of the environment variable named by the key.
func EnvStr(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	} else {
		return defaultValue
	}

}

// EnvInt retrieves the integer value of the environment variable named by the key.
func EnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		// make sure that the value is an integer
		value, _ := strconv.Atoi(value)
		return value
	} else {
		return defaultValue
	}
}

// EnvBool retrieves the boolean value of the environment variable named by the key.
func EnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		switch value {
		case "true":
			return true
		case "false":
			return false
		default:
			log.Fatalf("Invalid boolean value for %s: %s", key, value)
		}
	}
	return defaultValue
}
