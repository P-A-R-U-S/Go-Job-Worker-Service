package jobWorker

import (
	"context"
	"errors"
	"fmt"
	ns "github.com/P-A-R-U-S/Go-Job-Worker-Service/pkg/jobWorker/namespaces"
	"github.com/google/uuid"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
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
	ErrJobAlreadyStopped       = errors.New("Job already stopped")
)

type State string

const (
	JOB_STATUS_NOT_STARTED State = "NotStarted"
	JOB_STATUS_RUNNING     State = "Running"
	JOB_STATUS_COMPLETED   State = "Completed"
	JOB_STATUS_TERMINATED  State = "Terminated"
)

const (
	STOP_GRACE_PERIOD = 10 * time.Second
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
	// processState holds information about the process once it completes
	// 				and has `nil` until the job has completed running
	processState *os.ProcessState
	// isTerminated is true if the job has been started
	isStarted bool
	// isCompleted is true if the job has been successfully completed
	isCompleted bool
	// isTerminated is true if the job has been terminated via Stop()
	isTerminated bool
	// exitReason is the reason the job has errored if it has errored during execution or cleanup
	exitReason error
}

func (job *Job) getCGroupName() string {
	return strings.Replace(job.UUID.String(), "-", "", -1)
}

func (job *Job) String() string {
	return fmt.Sprintf("id:%s with command:%s %s", job.UUID, job.config.Command, strings.Join(job.config.Arguments, " "))
}

func NewJob(config *JobConfig) *Job {
	output := NewCommandOutput()
	job := &Job{
		UUID:   uuid.New(),
		config: config,
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
	if job.isStarted {
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
		Setsid: true,
		// Also, enables mounting a new proc filesystem so that command such as `ps -ef` only see the processes in the PID namespace
		Unshareflags: syscall.CLONE_NEWNS,
		// instruct cmd.Run to use the control group file descriptor, so that Job Command does not
		// have to manually add the new PID to the control group
		UseCgroupFD: true,
	}

	cleanCGroup := make(chan bool)
	go func() {
		select {
		case <-cleanCGroup:
			{
				// do not close the cgroup.procs file until after the process has exited
				if err := ns.DeleteCGroup(job.getCGroupName()); err != nil {
					log.Printf("error closing cgroup: %s\n", err)
					job.exitReason = errors.Join(job.exitReason, fmt.Errorf("error closing cgroup: %w\n", err))
				}
			}
		}
	}()

	err := ns.CreateCGroup(job.getCGroupName())
	if err != nil {
		cleanCGroup <- true
		return fmt.Errorf("error creating cgroup: %w", err)
	}

	if err = ns.AddResourceControl(job.getCGroupName(), ns.CPU_WEIGHT_File, strconv.Itoa(int(job.config.CPU*100))); err != nil {
		cleanCGroup <- true
		log.Printf("could not add resources into controller:%s, %v", ns.CPU_WEIGHT_File, err)
		return fmt.Errorf("error starting command: %w", err)
	}
	if err = ns.AddResourceControl(job.getCGroupName(), ns.MEMORY_HIGH_File, strconv.FormatInt(job.config.MemBytes, 10)); err != nil {
		cleanCGroup <- true
		return fmt.Errorf("could not add resources into controller:%s, %v", ns.MEMORY_HIGH_File, err)
	}
	//if err = ns.AddResourceControl(job.getCGroupName(), ns.IO_WEIGHT_File, strconv.FormatInt(job.config.IOBytesPerSecond, 10)); err != nil {
	//	return fmt.Errorf("could not add resources into controller:%s, %v", ns.IO_WEIGHT_File, err)
	//}

	//provide the file descriptor to cmd.Run so that it can add the new PID to the control group
	if err = ns.AddProcess(job.getCGroupName(), cmd); err != nil {
		cleanCGroup <- true
		return fmt.Errorf("Error AddProcess /proc - %w\n", err)
	}

	unmount := make(chan bool)
	go func() {
		select {
		case <-unmount:
			{
				if err := ns.UnmountProc(); err != nil {
					log.Printf("error unmounting /proc - %s\n", err)
					job.exitReason = errors.Join(job.exitReason, fmt.Errorf("error unmounting /proc - %w\n", err))
				}
			}
		}
	}()

	if err = ns.MountProc(); err != nil {
		cleanCGroup <- true
		return fmt.Errorf("Error mounting /proc - %w\n", err)
	}

	log.Printf("starting job:%s, cmd:%s", job, cmd.String())
	if err = cmd.Start(); err != nil {
		cleanCGroup <- true
		unmount <- true
		return fmt.Errorf("error starting command: %w", err)
	}
	job.cmd = cmd
	job.isStarted = true

	// run the command in a Goroutine so that Start can return immediately
	go func() {
		// Use cmd.Process.Wait() instead of cmd.Wait() since cmd.Wait() is not thread safe
		// and we do not want to hold the mutex while waiting for the process to exit.
		// So instead we use cmd.Process.Wait() and store the result in j.processState to mimic what
		// cmd.Wait would do.
		// This prevents concurrency issues when a user calls Start(), the command quickly exits (updating the
		// process state), and the user invokes Status().
		processState, err := job.cmd.Process.Wait()
		job.processState = processState

		job.mutex.Lock()
		defer job.mutex.Unlock()

		// at this stage job in completed (successfully or not we can detect from checking job.exitReason and isTerminated )
		job.isCompleted = true

		defer func() {
			cleanCGroup <- true
			unmount <- true
		}()

		if err != nil {
			job.exitReason = errors.Join(job.exitReason, fmt.Errorf("error running command: %w\n", err))
		}

		if err == nil && !job.processState.Success() {
			job.exitReason = errors.Join(job.exitReason, &exec.ExitError{ProcessState: job.processState})
		}

		// close the output, so that any readers of the output know the process has exited and will no longer
		// block waiting for new output
		if err = job.output.Close(); err != nil {
			job.exitReason = errors.Join(job.exitReason, fmt.Errorf("error closing output: %w\n", err))
		}
	}()

	return nil
}

// Status returns the current Status of the Job.
func (job *Job) Status() *JobStatus {
	job.mutex.Lock()
	defer job.mutex.Unlock()

	if !job.isStarted {
		return &JobStatus{
			State:    JOB_STATUS_NOT_STARTED,
			ExitCode: -1,
		}
	}

	if job.isStarted && !job.isCompleted {
		return &JobStatus{
			State:    JOB_STATUS_RUNNING,
			ExitCode: -1,
		}
	}

	if job.isTerminated {
		return &JobStatus{
			State:      JOB_STATUS_TERMINATED,
			ExitCode:   job.processState.ExitCode(),
			ExitReason: job.exitReason.Error(),
		}
	}

	return &JobStatus{
		State:      JOB_STATUS_COMPLETED,
		ExitCode:   job.processState.ExitCode(),
		ExitReason: job.exitReason.Error(),
	}
}

// Stream returns an OutputReadCloser (implements io.ReadCloser)  that streams the combined stdout and stderr of the Job.
func (job *Job) Stream() io.ReadCloser {
	log.Printf("get job stream:%s", job)
	return NewOutputReadCloser(job.output)
}

func (job *Job) Stop() error {
	if job.isTerminated || job.isCompleted {
		return ErrJobAlreadyStopped
	}

	job.mutex.Lock()
	defer job.mutex.Unlock()

	log.Printf("stop job :%s", job)
	if job.cmd == nil {
		return ErrJobNotStarted
	}

	if err := job.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("error sending SIGTERM: %w", err)
	}

	// Set up timeout
	killCtx, cancel := context.WithTimeout(context.Background(), STOP_GRACE_PERIOD)
	defer cancel()

	cmdWait := make(chan error, 1)
	go func() {
		_, err := job.cmd.Process.Wait()
		cmdWait <- err
	}()

	select {
	case <-cmdWait:
		{
			// command exited before timer expired, so nothing to do
			log.Printf("process compeled :%s", job)
			job.isCompleted = true
		}
	case <-killCtx.Done():
		{
			//send SIGKILL if process is still running after timer expires
			log.Printf("send SIGKILL to job:%s", job)
			if err := syscall.Kill(-job.cmd.Process.Pid, syscall.SIGKILL); err != nil {
				return fmt.Errorf("error sending SIGKILL: %w", err)
			}
			job.isTerminated = true
		}
	}

	return nil
}
