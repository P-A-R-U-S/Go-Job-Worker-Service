# Job Worker Service
 Job worker service provides client  and server API to run arbitrary command on 
 Linux 64-bit systems in an isolated manner

# Purpose
This library is intended to be used to run arbitrary commands on Linux 64-bit systems. 
To help prevent disruptions to other processes running on the system, the job is run in an isolated manner. 

This isolation includes:
* new PID namespace to prevent killing other processes on the host
* new mount namespace to mount a new proc filesystem so that the job can't see other processes on the host
* new network namespace to prevent the job from accessing the local network and internet
* configurable CPU, memory, and IO limits to throttle the job using cgroups v2


### Build and Run
1. Install [Go](https://go.dev/doc/install) language
2. Clone github repo
    ```
    $ git clone https://github.com/P-A-R-U-S/Go-Job-Worker-Service.git
    ```
3. Install [protoc and appropriate gRPC plugins](https://grpc.io/docs/languages/go/quickstart/)
    ```
    $ go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    $ go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
    ```
   
    > [!NOTE]
    > Repo already contains pre-generate certificate, but you can generate your certificate 
    > 
   > **make generate_certificate** 

4. Run server
    ```makefile
      make generate_grpc_code
      make run_server
     ```
    >[!NOTE] 
    > Server, using following address:port _0.0.0.0:8080_ 
    > (or _localhost:8080_) by default.

5. Run Client
    ```makefile
      make run_client_test
     ```

### Library usage
An example to run a job simple job: echo "Hello, World!":

```go
package main

import (
	"errors"
	"fmt"
	"github.com/P-A-R-U-S/Go-Job-Worker-Service/pkg/jobWorker"
)

func main() {

	// create a job config
	config := jobWorker.JobConfig{
		CPU:              "0.5",
		IOBytesPerSecond: 100_000,
		MemBytes:         1_000_000,
		Command:          "echo",
		Arguments:        []string{"Hello", "World!"},
	}

	newJob := jobWorker.NewJob(&config)
	
	// get status
	jobStatus := newJob.job.Status()
	fmt.Println("status:%s", jobStatus.State)
	
	// stop job
	if err := newJob.job.Stop(); err != nil {
		fmt.Errorf("error stopping job: %w", err)
	}
	fmt.Println("Status:%s", jobStatus.State)

	// get status
	jobStatus = newJob.job.Status()
	fmt.Println("status:%s, exitCode:%d, exitReason:%s", jobStatus.State, jobStatus.ExitCode, jobStatus.ExitReason)
}
```




