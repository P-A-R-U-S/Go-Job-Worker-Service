package jobWorker

import (
	"io"
	"log"
	"strings"
	"testing"
	"time"
)

func Test_Job_Running(t *testing.T) {
	// There is no problem to run test in parallel, but log output are confusing if you need to investigate anything.
	// TODO: Uncomment in final version when testing completely done.
	//t.Parallel()

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

func Test_Job_Second_Call_Stop_expected_not_send_SIGKIL_again(t *testing.T) {
	// There is no problem to run test in parallel, but log output are confusing if you need to investigate anything.
	// TODO: Uncomment in final version when testing completely done.
	//t.Parallel()

	config := JobConfig{
		Command:          "ping",
		Arguments:        []string{"google.com"}, //or localhost as an option {"-c", "1", "127.0.0.1"},
		CPU:              0.5,                    // half a CPU core
		IOBytesPerSecond: 100_000_000,            // 100 MB/s
		MemBytes:         1_000_000_000,          // 1 GB
	}

	testJob := NewJob(&config)

	// start the job
	err := testJob.Start()
	if err != nil {
		t.Fatalf("error starting job: %v", err)
	}

	if err = testJob.Stop(); err != nil {
		t.Fatalf("error stopping job: %v", err)
	}

	err = testJob.Stop()
	if err == nil || err.Error() != ErrJobAlreadyStopped.Error() {
		t.Fatalf("expected error(ErrJobAlreadyStopped): %v", err)
	}
}

func Test_Job_Prevents_NetworkRequests(t *testing.T) {
	//t.Parallel()
	//t.Skip()

	// Prove that the job-executor binary is not able to make network requests by showing that ping
	// to localhost fails since the loopback device is not turned on.
	config := JobConfig{
		Command:          "ping",
		Arguments:        []string{"127.0.0.1"}, //or localhost as an option {"-c", "1", "127.0.0.1"},
		CPU:              0.5,                   // half a CPU core
		IOBytesPerSecond: 100_000_000,           // 100 MB/s
		MemBytes:         1_000_000_000,         // 1 GB
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
		t.Errorf("expected job state to be '%s', got '%s'", JOB_STATUS_COMPLETED, status.State)
	}

	if status.ExitCode != 2 {
		t.Errorf("expected job exit code to be 1, got %d", status.ExitCode)
	}

	if len(status.ExitReason) != 0 {
		t.Errorf("expected job exit reason to be set when command errors, but got nil")
	}

	expectedPingOutput := "Operation not permitted" //"Network is unreachable"
	if !strings.Contains(string(output), expectedPingOutput) {
		t.Fatalf("expected output to contain %q, got %q", expectedPingOutput, output)
	}
}

func Test_Job_Stop_via_SIGKILL_expected_kill_all_child_processes(t *testing.T) {
	//t.Parallel()

	config := JobConfig{
		Command:          "/bin/sh",
		Arguments:        []string{"-c", "watch date > date.txt"},
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

	// stop job
	err = testJob.Stop()
	if err != nil {
		t.Errorf("error stopping job: %v", err)
	}

	status := testJob.Status()
	if len(status.ExitReason) > 0 {
		t.Errorf("expected no error return when stopping job with child processed: %s", status.ExitReason)
	}
}

func Test_Job_Stopping_Long_Lived_Command(t *testing.T) {
	//t.Parallel()

	config := JobConfig{
		Command:          "/bin/bash",
		Arguments:        []string{"-c", "while :; do  echo thinking; sleep 1; done"},
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
		t.Errorf("expected job state to be '%s', got '%s'", JOB_STATUS_RUNNING, status.State)
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
	time.Sleep(5 * time.Second)

	// stop job
	err = testJob.Stop()
	if err != nil {
		t.Errorf("error stopping job: %v", err)
	}

	var output []byte
	// wait for job to completely stop
	output, err = io.ReadAll(testJob.Stream())
	if err != nil {
		t.Errorf("error reading: %v", err)
	}
	log.Printf("output(%s)", string(output))

	status = testJob.Status()
	if status.State != JOB_STATUS_TERMINATED {
		t.Errorf("expected job state to be '%s', got '%s'", JOB_STATUS_TERMINATED, status.State)
	}

	if status.ExitCode != -1 {
		t.Errorf("expected job exit code to be -1, got %d", status.ExitCode)
	}

	if len(status.ExitReason) == 0 {
		// TODO: In Ubuntu exit reason signal:terminated, but can be different in other linux distro's. Need to research.
		t.Errorf("expected job exit reason not to be empty, because command had been terminated")
	}
}

func Test_Job_IOLimits(t *testing.T) {
	//t.Parallel()
	t.Skip()

	config := JobConfig{
		Command:          "dd",
		Arguments:        []string{"if=/dev/zero", "of=/tmp/file1", "bs=64M", "count=1", "oflag=direct"},
		CPU:              0.5,           // half a CPU core
		IOBytesPerSecond: 10_000_000,    // 10 MB/s
		MemBytes:         1_000_000_000, // 1 GB
	}

	testJob := NewJob(&config)

	// start the job
	err := testJob.Start()
	if err != nil {
		t.Fatalf("error starting job: %v", err)
	}

	// get job output
	output, err := io.ReadAll(testJob.Stream())
	if err != nil {
		t.Errorf("error reading output: %v", err)
	}

	// TODO: this test hasn't appeared flaky yet, but there's probably a better way to validate the output/IO limit
	expectedIOLimitOutput := "10.0 MB/s"
	if !strings.Contains(string(output), expectedIOLimitOutput) {
		t.Errorf("expected output to contain %q, got %q", expectedIOLimitOutput, output)
	}
}
