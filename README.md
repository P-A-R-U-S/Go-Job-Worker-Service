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

    > Repo already contains pre-generate certificate, but you can generate your certificate
    > **make generate_certificate** 

4. Run server
    ```makefile
      make run_server
     ```
    > Server, using following address:port _0.0.0.0:8080_ (or _localhost:8080_) by default.

5. Run Client
    
    ```makefile
      make run_client_test
     ```
     following command below run simple command. For proper testing you can run command in following order
* **start command** - `./jwcli --host 'localhost:8080' --ca-cert './certs/ca-cert.pem' --client-cert './certs/client-1-cert.pem' --client-key './certs/client-1-key.pem' start --cpu 0.5 --memory 1000000000 --io 10000000 --c 'echo' 'hello world'`


* **get status** - `./jwcli --host 'localhost:8080' --ca-cert './certs/ca-cert.pem' --client-cert './certs/client-1-cert.pem' --client-key './certs/client-1-key.pem' status --id <JOB ID>`


* **get command output** -`./jwcli --host 'localhost:8080' --ca-cert './certs/ca-cert.pem' --client-cert './certs/client-1-cert.pem' --client-key './certs/client-1-key.pem' stream --id <JOB ID>`


* **stop command execution** - `./jwcli --host 'localhost:8080' --ca-cert './certs/ca-cert.pem' --client-cert './certs/client-1-cert.pem' --client-key './certs/client-1-key.pem' stop --id $(JOB_ID)`


### Library usage
An example to run a simple job: _echo "Hello, World!"_:

```go
package main

import (
	"fmt"
	"github.com/P-A-R-U-S/Go-Job-Worker-Service/pkg/jobWorker"
	"os"
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

	// start new job
	if err = newJob.Start(); err != nil {
		fmt.Printf("error starting job: %w", err)
		os.Exit(-1)
	}

	// get status
	jobStatus := newJob.job.Status()
	fmt.Printf("status:%s", jobStatus.State)

	// stop job
	if err := newJob.job.Stop(); err != nil {
		log.Errorf("error stopping job: %w", err)
		os.Exit(-1)
	}
	fmt.Printf("Status:%s", jobStatus.State)

	// get status
	jobStatus = newJob.job.Status()
	fmt.Println("status:%s, exitCode:%d, exitReason:%s", jobStatus.State, jobStatus.ExitCode, jobStatus.ExitReason)
}
```




