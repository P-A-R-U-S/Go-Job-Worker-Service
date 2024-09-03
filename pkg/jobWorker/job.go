package jobWorker

import (
	"errors"
	"fmt"
	ns "github.com/P-A-R-U-S/Go-Job-Worker-Service/pkg/jobWorker/namespaces"
	"github.com/google/uuid"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	ErrJobAlreadyStarted       = errors.New("job already started")
	ErrJobNotStarted           = errors.New("job not started")
	ErrInvalidCommand          = errors.New("command must be provided")
	ErrInvalidCPU              = errors.New("CPU must be greater than 0")
	ErrInvalidIOBytesPerSecond = errors.New("IOBytesPerSecond must be greater than 0")
	ErrInvalidMemBytes         = errors.New("MemBytes must be greater than 0")
)

type State string

const (
	JOB_STATUS_NOT_STARTED State = "NotStarted"
	JOB_STATUS_RUNNING     State = "Running"
	JOB_STATUS_COMPLETED   State = "Completed"
	JOB_STATUS_TERMINATED  State = "Terminated"
)

// JobStatus represent current status of the Job
type JobStatus struct {
	State State
	// ExitCode is the exit code of the job if it has exited via exit().
	ExitCode int
	// ExitReason is the reason the job has errored if it has errored during execution or cleanup.
	ExitReason string
}

// JobConfig represent job configuration settings (all fields are required)
type JobConfig struct {
	// RootPhysicalDevice is the major and minor number of the root physical device to apply IOBytesPerSecond limit to.
	RootPhysicalDevice string
	// CPU is the number of CPU cores to limit the job to such as 0.5 for half a CPU core.
	CPU float64
	// MemBytes is the number of bytes to limit the job to use, such as 1_000_000_000 for 1 GB.
	MemBytes int64
	// IOBytesPerSecond is the number of bytes per second to limit the job to read/write on the
	//					provided RootPhysicalDevice, such as 100_000_000 for 100 MB/s.
	IOBytesPerSecond int64
	// Command is the command to run.
	Command string
	// Arguments are the arguments to pass to the command, if any.
	Arguments []string
}

func (jobConfig *JobConfig) isValid() error {
	if jobConfig.Command == "" {
		return ErrInvalidCommand
	}

	if jobConfig.CPU <= 0 {
		return ErrInvalidCPU
	}

	if jobConfig.IOBytesPerSecond <= 0 {
		return ErrInvalidIOBytesPerSecond
	}

	if jobConfig.MemBytes <= 0 {
		return ErrInvalidMemBytes
	}

	return nil
}

type Job struct {
	UUID   uuid.UUID
	cmd    *exec.Cmd
	mutex  sync.Mutex
	output *CommandOutput
	config *JobConfig
	status *JobStatus
	// processState holds information about the process once it completes
	// 				and has `nil` until the job has completed running
	processState *os.ProcessState
}

func (job *Job) String() string {
	return fmt.Sprintf("id:%s, state:%s with command:%s %s", job.UUID, job.status.State, job.config.Command, strings.Join(job.config.Arguments, " "))
}

func NewJob(config *JobConfig) *Job {
	output := NewCommandOutput()
	job := &Job{
		UUID:   uuid.New(),
		config: config,
		status: &JobStatus{
			State:    JOB_STATUS_NOT_STARTED,
			ExitCode: -1, // TODO: Probably we need to set default exist code e.g. 999999 to avoid confusions.
		},
		output: output,
	}
	log.Printf("creted job:%s", job)
	return job
}

// Start - starting the Job in a semi-isolated environment (creating new PID, mount and network and also creates a new control group for the process limiting CPU, IO, and memory)
// The user running Start() should be the root user or have the necessary permissions to create namespaces and control groups
//
// ErrJobAlreadyStarted is returned, if the Job has already been started.
// ErrInvalidCommand, ErrInvalidCPU, ErrInvalidIOBytesPerSecond, ErrInvalidMemBytes is returned, if provided configuration is invalid
func (job *Job) Start() error {
	job.mutex.Lock()
	defer job.mutex.Unlock()

	// validate configuration
	log.Printf("validate job:%s", job)
	if err := job.config.isValid(); err != nil {
		log.Printf("validate job error:%v", err)
		return err
	}

	// validate job hasn't been run
	if job.status.State != JOB_STATUS_NOT_STARTED {
		return ErrJobAlreadyStarted
	}

	cmd := exec.Command(job.config.Command, job.config.Arguments...)
	// combine the stdout and stderr so that the stdout and stderr are combined in the order they are written
	cmd.Stderr = job.output
	cmd.Stdout = job.output
	cmd.SysProcAttr = &syscall.SysProcAttr{
		// CLONE_NEWPID:  creates a new PID namespace preventing the process from seeing/killing host processes
		// CLONE_NEWNET:  creates a new network namespace preventing the process from accessing the internet or local network
		// CLONE_NEWNS:   creates a new mount namespace preventing the process from impacting host mounts
		// CLONE_NEWUTS:  creates a new UTS namespaces provide isolation between two system identifiers: the hostname and the NIS domain name
		// CLONE_NEWPID:  crates new PID namespaces isolate the process ID number space, meaning that processes in different PID namespaces can have the same PID
		// CLONE_NEWUSER: creates new namespaces to isolate security-related identifiers and attributes, in particular, user IDs and group IDs
		Cloneflags: syscall.CLONE_NEWNS |
			syscall.CLONE_NEWIPC |
			syscall.CLONE_NEWNET |
			syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWUSER,
		UidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      os.Getuid(),
				Size:        1,
			},
		},
		GidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      os.Getgid(),
				Size:        1,
			},
		},
		// Also, enables mounting a new proc filesystem so that command such as `ps -ef` only see the processes in the PID namespace
		Unshareflags: syscall.CLONE_NEWNS,
		// instruct cmd.Run to use the control group file descriptor, so that Job Command does not
		// have to manually add the new PID to the control group
		UseCgroupFD: true,
	}

	formatedUUID := strings.Replace(job.UUID.String(), "-", "", -1)

	cgroupName := formatedUUID
	cgroupDir := ns.GetCGroupPath(cgroupName)

	err := ns.CreateCGroup(cgroupDir, job.config.RootPhysicalDevice, job.config.CPU, job.config.IOBytesPerSecond, job.config.MemBytes)
	if err != nil {
		return fmt.Errorf("error creating cgroup: %w", err)
	}

	// open the cgroup.procs file so cmd.Run can automatically add the new PID to the control group
	cgroupTasksDir := filepath.Join(cgroupDir, "tasks")

	procsFile, err := os.OpenFile(cgroupTasksDir, os.O_RDONLY, 0)
	if err != nil {
		return fmt.Errorf("error opening cgroup.procs: %w", err)
	}

	// provide the file descriptor to cmd.Run so that it can add the new PID to the control group
	cmd.SysProcAttr.CgroupFD = int(procsFile.Fd())

	if err := ns.MountProc(); err != nil {
		fmt.Printf("Error mounting /proc - %s\n", err)
		os.Exit(1)
	}

	//if err := ns.PivotRoot(rootfs); err != nil {
	//	fmt.Printf("Error running pivot_root - %s\n", err)
	//	os.Exit(1)
	//}

	log.Printf("starting job:%s", job)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error starting command: %w", err)
	}
	job.cmd = cmd
	job.status.State = JOB_STATUS_RUNNING

	// run the command in a Goroutine so that Start can return immediately
	go func() {
		// Use cmd.Process.Wait() instead of cmd.Wait() since cmd.Wait() is not thread safe
		// and we do not want to hold the mutex while waiting for the process to exit.
		// So instead we use cmd.Process.Wait() and store the result in j.processState to mimic what
		// cmd.Wait would do.
		// This prevents concurrency issues when a user calls Start(), the command quickly exits (updating the
		// process state), and the user invokes Status().
		processState, err := job.cmd.Process.Wait()
		job.mutex.Lock()
		defer job.mutex.Unlock()

		job.processState = processState
		job.status.ExitCode = job.processState.ExitCode()
		if job.status.State == JOB_STATUS_RUNNING {
			job.status.State = JOB_STATUS_COMPLETED
		}

		if err != nil {
			job.status.ExitReason = job.status.ExitReason + fmt.Sprintf("error running command: %s\n", err)
		}

		if err == nil && !job.processState.Success() {
			job.status.ExitReason = job.status.ExitReason + fmt.Sprintf("error running command: %s\n", &exec.ExitError{ProcessState: job.processState})
		}

		// close the output, so that any readers of the output know the process has exited and will no longer
		// block waiting for new output
		if err = job.output.Close(); err != nil {
			job.status.ExitReason = job.status.ExitReason + fmt.Sprintf("error closing output: %s\n", err)
		}

		// do not close the cgroup.procs file until after the process has exited
		if err = procsFile.Close(); err != nil {
			job.status.ExitReason = job.status.ExitReason + fmt.Sprintf("error closing cgroup.procs: %w", err)
		}

		// do not close the cgroup.procs file until after the process has exited
		if err = ns.CleanupCGroup(cgroupName); err != nil {
			job.status.ExitReason = job.status.ExitReason + fmt.Sprintf("error closing cgroup: %s\n", err)
		}

		if err = ns.UnmountProc(); err != nil {
			log.Printf("error unmounting /proc - %s\n", err)
		}
	}()

	return nil
}

// Status returns the current Status of the Job.
func (job *Job) Status() *JobStatus {
	job.mutex.Lock()
	defer job.mutex.Unlock()
	log.Printf("get job status:%s", job)
	return job.status
}

// Stream returns an OutputReadCloser that streams the combined stdout and stderr of the Job.
func (job *Job) Stream() *OutputReadCloser {
	log.Printf("get job stream:%s", job)
	return NewOutputReadCloser(job.output)
}

func (job *Job) Stop() error {
	job.mutex.Lock()
	defer job.mutex.Unlock()

	log.Printf("stop job :%s", job)
	if job.cmd == nil {
		return ErrJobNotStarted
	}

	if err := job.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("error sending SIGTERM: %w", err)
	}

	cmdWait := make(chan error, 1)
	timer := time.NewTimer(10 * time.Second)
	// stop timer in case process exits before 10 seconds
	// it's safe to stop timer even if stopped already
	defer timer.Stop()

	go func() {
		_, err := job.cmd.Process.Wait()
		cmdWait <- err
	}()

	select {
	case <-cmdWait:
		// command exited before timer expired, so nothing to do
	case <-timer.C:
		// send SIGKILL if process is still running after timer expires
		log.Printf("send SIGKILL to job:%s", job)
		if err := job.cmd.Process.Signal(syscall.SIGKILL); err != nil {
			return fmt.Errorf("error sending SIGKILL: %w", err)
		}
	}

	job.status.State = JOB_STATUS_TERMINATED

	return nil
}
