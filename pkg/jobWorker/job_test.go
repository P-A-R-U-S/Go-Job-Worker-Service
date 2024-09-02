package jobWorker

import (
	"io"
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

	if status.State != "completed" {
		t.Errorf("expected job state to be 'completed', got '%s'", status.State)
	}

	if status.ExitCode != 0 {
		t.Errorf("expected job exit code to be 0, got %d", status.ExitCode)
	}

	if status.State != JOB_STATUS_COMPLETED {
		t.Error("expected job to have exited")
	}

	if len(status.ExitReason) == 0 {
		t.Errorf("expected job exit reason to be empty, got %v", status.ExitReason)
	}

	if string(output) != "hello world\n" {
		t.Fatalf("expected output to be 'hello world\n', got '%s'", output)
	}
}

//func getRootPhysicalDevice(t *testing.T) string {
//	t.Helper()
//
//	rootDeviceMajMin := os.Getenv("ROOT_DEVICE_MAJ_MIN")
//	if rootDeviceMajMin == "" {
//		t.Fatal("ROOT_DEVICE_MAJ_MIN environment variable must be set")
//	}
//
//	return rootDeviceMajMin
//}
