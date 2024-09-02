package jobWorker

import (
	"io"
	"strings"
	"testing"
	"time"
)

func Test_Job_Running(t *testing.T) {
	t.Parallel()

	config := JobConfig{
		Command:          "echo",
		Arguments:        []string{"hello", "world"},
		CPU:              0.5,           // half a CPU core
		IOBytesPerSecond: 100_000_000,   // 100 MB/s
		MemBytes:         1_000_000_000, // 1 GB
	}

	testJob := NewJob(&config)

	// start the job
	err := testJob.Start()
	if err != nil {
		t.Fatalf("error starting job: %v", err)
	}

	// wait for the job to finish by waiting for io.ReadAll to complete
	output, err := io.ReadAll(testJob.Stream())
	if err != nil {
		t.Fatalf("error reading output: %v", err)
	}

	status := testJob.Status()

	if status.State != JOB_STATUS_COMPLETED {
		t.Errorf("expected job state to be 'Completed', got '%s'", status.State)
	}

	if status.ExitCode != 0 {
		t.Errorf("expected job exit code to be 0, got %d", status.ExitCode)
	}

	if len(status.ExitReason) != 0 {
		t.Errorf("expected job exit reason to be empty, got %v", status.ExitReason)
	}

	if string(output) != "hello world\n" {
		t.Fatalf("expected output to be 'hello world\n', got '%s'", output)
	}
}

func Test_Job_Prevents_NetworkRequests(t *testing.T) {
	// Prove that the job-executor binary is not able to make network requests by showing that ping
	// to localhost fails since the loopback device is not turned on.
	t.Parallel()
	t.Skip()

	config := JobConfig{
		Command:          "ping",
		Arguments:        []string{"-c", "1", "google.com"}, //or localhost as an option {"-c", "1", "127.0.0.1"},
		CPU:              0.5,                               // half a CPU core
		IOBytesPerSecond: 100_000_000,                       // 100 MB/s
		MemBytes:         1_000_000_000,                     // 1 GB
	}

	testJob := NewJob(&config)

	// start the job
	if err := testJob.Start(); err != nil {
		t.Fatalf("error starting job: %v", err)
	}

	// wait for the job to finish by waiting for io.ReadAll to complete
	output, err := io.ReadAll(testJob.Stream())
	if err != nil {
		t.Fatalf("error reading output: %v", err)
	}

	status := testJob.Status()

	if status.State != JOB_STATUS_COMPLETED {
		t.Errorf("expected job state to be 'completed', got '%s'", status.State)
	}

	if status.ExitCode != 1 {
		t.Errorf("expected job exit code to be 1, got %d", status.ExitCode)
	}

	if len(status.ExitReason) == 0 {
		t.Errorf("expected job exit reason to be set when command errors, but got nil")
	}

	expectedPingOutput := "Network is unreachable"
	if !strings.Contains(string(output), expectedPingOutput) {
		t.Fatalf("expected output to contain %q, got %q", expectedPingOutput, output)
	}
}

func Test_Job_Stopping_Long_Lived_Command(t *testing.T) {
	t.Parallel()

	config := JobConfig{
		Command:          "bin/bash",
		Arguments:        []string{"c-", "'while sleep 2; do echo thinking; done'"},
		CPU:              0.5,           // half a CPU core
		IOBytesPerSecond: 100_000_000,   // 100 MB/s
		MemBytes:         1_000_000_000, // 1 GB
	}

	testJob := NewJob(&config)

	// start the job
	err := testJob.Start()
	if err != nil {
		t.Fatalf("error starting job: %v", err)
	}

	// validate status while job is running
	status := testJob.Status()

	if status.State != JOB_STATUS_RUNNING {
		t.Errorf("expected job state to be 'running', got '%s'", status.State)
	}

	if status.ExitCode != -1 {
		t.Errorf("expected job exit code to be -1, got %d", status.ExitCode)
	}

	if len(status.ExitReason) > 0 {
		t.Errorf("expected job exit reason to not be set when command is still running")
	}

	// TODO: find a better workaround
	// This is a hack to ensure the job-executor's signal handler has time to be set up.
	// Otherwise, job-executor misses the SIGTERM signal.
	time.Sleep(1 * time.Second)

	// stop job
	err = testJob.Stop()
	if err != nil {
		t.Errorf("error stopping job: %v", err)
	}

	// wait for job to completely stop
	_, err = io.ReadAll(testJob.Stream())
	if err != nil {
		t.Errorf("error reading output: %v", err)
	}

	status = testJob.Status()
	if status.State != JOB_STATUS_TERMINATED {
		t.Errorf("expected job state to be 'terminated', got '%s'", status.State)
	}

	if status.ExitCode != -1 {
		t.Errorf("expected job exit code to be -1, got %d", status.ExitCode)
	}

	if len(status.ExitReason) > 0 {
		t.Errorf("expected job exit reason to be nil with no errors while waiting for command to finish, but got %v", status.ExitReason)
	}
}
