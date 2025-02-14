package main

import (
	"testing"
	"time"

	"github.com/devinodaniel/cron-go/cmd/config"
)

func TestNew(t *testing.T) {
	config.CRON_METRICS = false

	args := []string{"echo", "hello"}
	cron := New(args)

	if len(cron.Args) != len(args) {
		t.Errorf("Expected args length %d, got %d", len(args), len(cron.Args))
	}

	for i, arg := range args {
		if cron.Args[i] != arg {
			t.Errorf("Expected arg %s, got %s", arg, cron.Args[i])
		}
	}
}

func TestRunSuccess(t *testing.T) {
	config.CRON_METRICS = false

	args := []string{"echo", "hello"}
	cron := New(args)

	err := cron.Run()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if cron.ExitCode != CRON_SUCCESS {
		t.Errorf("Expected exit code %d, got %d", CRON_SUCCESS, cron.ExitCode)
	}
}

func TestRunFail(t *testing.T) {
	config.CRON_METRICS = false

	args := []string{"false"}
	cron := New(args)

	err := cron.Run()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if cron.ExitCode != CRON_FAIL {
		t.Errorf("Expected exit code %d, got %d", CRON_FAIL, cron.ExitCode)
	}
}

func TestRunExitCode1(t *testing.T) {
	config.CRON_METRICS = false

	args := []string{"test", "-f", "/tmp/does_not_exist"}
	cron := New(args)

	err := cron.Run()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if cron.ExitCode != CRON_FAIL {
		t.Errorf("Expected exit code %d, got %d", CRON_FAIL, cron.ExitCode)
	}
}

func TestRunTimeout(t *testing.T) {
	config.CRON_METRICS = false

	// Set a very short timeout for the test
	config.CRON_TIMEOUT = 1

	args := []string{"sleep", "2"}
	cron := New(args)

	err := cron.Run()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if cron.ExitCode != CRON_TIMEOUT {
		t.Errorf("Expected exit code %d, got %d", CRON_TIMEOUT, cron.ExitCode)
	}
}

func TestCronDuration(t *testing.T) {
	config.CRON_METRICS = false

	args := []string{"echo", "hello"}
	cron := New(args)

	cron.start()
	time.Sleep(1 * time.Second)
	cron.finish()

	if cron.Duration < 1*time.Second || cron.Duration > 2*time.Second {
		t.Errorf("Expected duration to be at least 1 second, got %s", cron.Duration)
	}

	if cron.Duration.Milliseconds() < 1000 || cron.Duration.Milliseconds() > 2000 {
		t.Errorf("Expected duration to be at least 1000 milliseconds, got %d", cron.Duration.Milliseconds())
	}
}

func TestWriteMetricsWithNamespace(t *testing.T) {
	config.CRON_NAMESPACE = "test namespace"
	config.CRON_METRICS = false

	args := []string{"echo", "hello"}
	cron := New(args)

	cron.start()
	time.Sleep(1 * time.Second)
	cron.finish()

	if cron.Monitor.Namespace != "test_namespace" {
		t.Errorf("Expected namespace to be %s, got %s", "test_namespace", cron.Monitor.Namespace)
	}
}

func TestWriteMetricsWithNamespaceCapsAndDash(t *testing.T) {
	config.CRON_NAMESPACE = "TEST-nameSPACE"
	config.CRON_METRICS = false

	args := []string{"echo", "hello"}
	cron := New(args)

	cron.start()
	time.Sleep(1 * time.Second)
	cron.finish()

	if cron.Monitor.Namespace != "test_namespace" {
		t.Errorf("Expected namespace to be %s, got %s", "test_namespace", cron.Monitor.Namespace)
	}
}

func TestWriteMetricsWithNamespaceSpecialChars(t *testing.T) {
	config.CRON_NAMESPACE = "TEST-nameSPACE!@$%^&*()-=+"
	config.CRON_METRICS = false

	args := []string{"echo", "hello"}
	cron := New(args)

	cron.start()
	time.Sleep(1 * time.Second)
	cron.finish()

	if cron.Monitor.Namespace != "test_namespace" {
		t.Errorf("Expected namespace to be %s, got %s", "test_namespace", cron.Monitor.Namespace)
	}
}

func TestWriteMetricsWithNamespaceSpecialCharsWithSpaces(t *testing.T) {
	config.CRON_NAMESPACE = "TEST-nameSPACE!@$%^&*()-=+ TEST AGAIN"
	config.CRON_METRICS = false

	args := []string{"echo", "hello"}
	cron := New(args)

	cron.start()
	time.Sleep(1 * time.Second)
	cron.finish()

	if cron.Monitor.Namespace != "test_namespace_____________test_again" {
		t.Errorf("Expected namespace to be %s, got %s", "test_namespace_____________test_again", cron.Monitor.Namespace)
	}
}
