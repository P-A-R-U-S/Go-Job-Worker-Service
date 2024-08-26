# Go-Job-Worker-Service
 Job worker service that provides an API to run arbitrary Linux processes

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

*  Build server

```makefile
  make -C pkg/proto
 ```



