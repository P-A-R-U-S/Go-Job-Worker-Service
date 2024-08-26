# Job Worker Service
 Job worker service provides an API to run arbitrary Linux processes on remote host

### Build
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

*  Run server

    ```makefile
      make run_server
     ```
    > **Note:** Server using following address:port _0.0.0.0:8080_ (or _localhost:8080_) by default.

* Run Client
    ```makefile
      make run_client
     ```
 




