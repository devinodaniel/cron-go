package main

import (
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/devinodaniel/cron-go/cmd/config"
)

func TestNew(t *testing.T) {
	config.CRON_METRICS = false

	args := []string{"echo", "hello"}
	cron, _ := New(args)

	if len(cron.Args) != len(args) {
		t.Errorf("Expected args length %d, got %d", len(args), len(cron.Args))
	}

	for i, arg := range args {
		if cron.Args[i] != arg {
			t.Errorf("Expected arg %s, got %s", arg, cron.Args[i])
		}
	}
}

func TestNewNoArgs(t *testing.T) {
	config.CRON_METRICS = false

	args := []string{}

	cron, err := New(args)

	if cron != nil {
		t.Errorf("Expected cron to be nil, got %v", cron)
	}

	noArgsError := "No arguments provided. Nothing to do. Run 'help' for usage."

	if err.Error() != noArgsError {
		t.Errorf("Expected error '%s', got %s", noArgsError, err.Error())
	}
}

func TestRunSimpleSuccess(t *testing.T) {
	config.CRON_METRICS = false

	args := []string{"echo", "hello"}
	cron, _ := New(args)

	err := cron.Run()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if cron.StatusCode != CRON_STATUS_SUCCESS {
		t.Errorf("Expected status code %d, got %d", CRON_STATUS_SUCCESS, cron.StatusCode)
	}

	if cron.ExitCode != CRON_EXITCODE_SUCCESS {
		t.Errorf("Expected exit code %d, got %d", 0, cron.ExitCode)
	}
}

func TestRunRubySuccess(t *testing.T) {
	config.CRON_METRICS = false

	args := []string{"ruby", "-e", "puts 'hello'"}
	cron, _ := New(args)

	err := cron.Run()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if cron.StatusCode != CRON_STATUS_SUCCESS {
		t.Errorf("Expected status code %d, got %d", CRON_STATUS_SUCCESS, cron.StatusCode)
	}

	if cron.ExitCode != CRON_EXITCODE_SUCCESS {
		t.Errorf("Expected exit code %d, got %d", 0, cron.ExitCode)
	}
}

func TestRunSimpleFailStatus(t *testing.T) {
	config.CRON_METRICS = false

	args := []string{"false"}
	cron, _ := New(args)

	cron.Run()

	if cron.StatusCode != CRON_STATUS_FAIL {
		t.Errorf("Expected status code %d, got %d", CRON_STATUS_FAIL, cron.StatusCode)
	}

	if cron.ExitCode == CRON_EXITCODE_SUCCESS {
		t.Errorf("Expected exit code to not be 0 (SUCCESS), got %d", cron.ExitCode)
	}

	if cron.EndTime.IsZero() {
		t.Errorf("Expected end time to be set, got %v", cron.EndTime)
	}
}

func TestRunSimpleExitCode1(t *testing.T) {
	config.CRON_METRICS = false

	args := []string{"test", "-f", "/tmp/does_not_exist"}
	cron, _ := New(args)

	cron.Run()

	if cron.StatusCode != CRON_STATUS_FAIL {
		t.Errorf("Expected status code %d, got %d", CRON_STATUS_FAIL, cron.StatusCode)
	}

	if cron.ExitCode != CRON_EXITCODE_FAIL_GENERIC {
		t.Errorf("Expected exit code %d, got %d", 1, cron.ExitCode)
	}
}

// TestRunExitCode126 tests the exit code when permission is denied
func TestRunExitCode126(t *testing.T) {
	config.CRON_METRICS = false

	args := []string{"/dev/null"}
	cron, _ := New(args)

	cron.Run()

	if cron.StatusCode != CRON_STATUS_FAIL {
		t.Errorf("Expected status code %d, got %d", CRON_STATUS_FAIL, cron.StatusCode)
	}

	if cron.ExitCode != CRON_EXITCODE_PERM_DENIED {
		t.Errorf("Expected exit code %d, got %d", 127, cron.ExitCode)
	}
}

// TestRunExitCode127 tests the exit code when the command is not found
func TestRunExitCode127(t *testing.T) {
	config.CRON_METRICS = false

	args := []string{"invalidornonexistentcommand"}
	cron, _ := New(args)

	cron.Run()

	if cron.StatusCode != CRON_STATUS_FAIL {
		t.Errorf("Expected status code %d, got %d", CRON_STATUS_FAIL, cron.StatusCode)
	}

	if cron.ExitCode != CRON_EXITCODE_EXEC_NOT_FOUND {
		t.Errorf("Expected exit code %d, got %d", 127, cron.ExitCode)
	}
}

func TestRunSigTerminated(t *testing.T) {
	config.CRON_METRICS = false

	args := []string{"sleep", "5"}
	cron, _ := New(args)

	// Set up a channel to receive the interrupt signal
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT)

	// Send an interrupt signal to the current process after 2 seconds
	go func() {
		time.Sleep(1 * time.Second)
		sigs <- syscall.SIGINT
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}()

	err := cron.Run()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if cron.StatusCode != CRON_STATUS_TERMINATED {
		t.Errorf("Expected status code %d, got %d", CRON_STATUS_TERMINATED, cron.StatusCode)
	}

	if cron.ExitCode != CRON_EXITCODE_SIG_INT {
		t.Errorf("Expected exit code %d, got %d", -1, cron.ExitCode)
	}
}

func TestRunTimeout(t *testing.T) {
	config.CRON_METRICS = false

	// Set a very short timeout for the test
	config.CRON_TIMEOUT = 1

	args := []string{"sleep", "2"}
	cron, _ := New(args)

	cron.Run()

	if cron.StatusCode != CRON_STATUS_TIMEOUT {
		t.Errorf("Expected status code %d, got %d", CRON_STATUS_TIMEOUT, cron.StatusCode)
	}

	if cron.ExitCode != CRON_EXITCODE_FAIL_GENERIC {
		t.Errorf("Expected exit code %d, got %d", 1, cron.ExitCode)
	}
}

func TestCronDuration(t *testing.T) {
	config.CRON_METRICS = false

	args := []string{"sleep", "1"}
	cron, _ := New(args)

	cron.start()
	cron.finish()

	if cron.Duration < 1*time.Second || cron.Duration > 2*time.Second {
		t.Errorf("Expected duration to be at least 1 second, got %s", cron.Duration)
	}

	if cron.Duration.Milliseconds() < 1000 || cron.Duration.Milliseconds() > 2000 {
		t.Errorf("Expected duration to be at least 1000 milliseconds and less than 2000 ms, got %d", cron.Duration.Milliseconds())
	}
}

func TestMetricsNamespace(t *testing.T) {
	config.CRON_NAMESPACE = "test namespace"
	config.CRON_METRICS = false

	args := []string{"echo", "hello"}
	cron, _ := New(args)

	cron.start()
	cron.finish()

	if cron.Monitor.Namespace != "test_namespace" {
		t.Errorf("Expected namespace to be %s, got %s", "test_namespace", cron.Monitor.Namespace)
	}
}

func TestMetricsNamespaceCapsAndDash(t *testing.T) {
	config.CRON_NAMESPACE = "TEST-nameSPACE"
	config.CRON_METRICS = false

	args := []string{"echo", "hello"}
	cron, _ := New(args)

	cron.start()
	cron.finish()

	if cron.Monitor.Namespace != "test_namespace" {
		t.Errorf("Expected namespace to be %s, got %s", "test_namespace", cron.Monitor.Namespace)
	}
}

func TestMetricsWithNamespaceSpecialChars(t *testing.T) {
	config.CRON_NAMESPACE = "TEST-nameSPACE!@$%^&*()-=+"
	config.CRON_METRICS = false

	args := []string{"echo", "hello"}
	cron, _ := New(args)

	cron.start()
	cron.finish()

	if cron.Monitor.Namespace != "test_namespace" {
		t.Errorf("Expected namespace to be %s, got %s", "test_namespace", cron.Monitor.Namespace)
	}
}

func TestWriteMetricsWithNamespaceSpecialCharsWithSpaces(t *testing.T) {
	config.CRON_NAMESPACE = "TEST-nameSPACE!@$%^&*()-=+ TEST AGAIN"
	config.CRON_METRICS = false

	args := []string{"echo", "hello"}
	cron, _ := New(args)

	cron.start()
	cron.finish()

	if cron.Monitor.Namespace != "test_namespace_____________test_again" {
		t.Errorf("Expected namespace to be %s, got %s", "test_namespace_____________test_again", cron.Monitor.Namespace)
	}
}

func TestWriteMetricsWithNamespaceWithFilepath(t *testing.T) {
	config.CRON_NAMESPACE = "TEST-nameSPACE!@$%^&*()-=+ TEST AGAIN"
	config.CRON_METRICS = false

	args := []string{"cat", "/tmp/does_not_exist/this/should/not/exist.txt"}
	cron, _ := New(args)

	cron.start()
	cron.finish()

	if cron.Monitor.Namespace != "test_namespace_____________test_again" {
		t.Errorf("Expected namespace to be %s, got %s", "test_namespace_____________test_again", cron.Monitor.Namespace)
	}

	if cron.StatusCode != CRON_STATUS_FAIL {
		t.Errorf("Expected status code %d, got %d", CRON_STATUS_FAIL, cron.StatusCode)
	}

	if cron.ExitCode != CRON_EXITCODE_FAIL_GENERIC {
		t.Errorf("Expected exit code %d, got %d", 1, cron.ExitCode)
	}
}
