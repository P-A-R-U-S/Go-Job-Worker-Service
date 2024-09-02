package jobWorker

import (
	"io"
	"strings"
	"testing"
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
