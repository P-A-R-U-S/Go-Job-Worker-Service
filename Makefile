generate_certificate:
	sh generate_certs.sh

generate_grpc_code:
	 protoc --go_out=. \
	 		--go_opt=paths=source_relative \
	 		--go-grpc_out=require_unimplemented_servers=false:. \
	 		--go-grpc_opt=paths=source_relative \
	 		./pkg/proto/jobWorker.proto

run_build:
	go mod tidy
	gofmt -w pkg/tls/tls.go
	gofmt -w pkg/proto/jobWorker.pb.go,pkg/proto/jobWorker_grpc.pb.go
	gofmt -w server/*.go
	gofmt -w build cli/*.go

	go build -o server -race server/*.go
	go build -o cli -race cli/main.go

run_test_cgroup:
	go mod tidy
	gofmt -w pkg/jobWorker/namespaces/*.go
	# "to create cgroup we need root permission"
	sudo go test -v -race pkg/jobWorker/namespaces/*.go -run "^Test_CGroup"

run_test_output:
	go mod tidy
	gofmt -w pkg/jobWorker/*.go
	go test -v -race pkg/jobWorker/*.go -run "^Test_CommandOutput"
	go test -v -race pkg/jobWorker/*.go -run "^Test_OutputReadCloser"

run_test_job:
	go mod tidy
	gofmt -w pkg/jobWorker/*.go
	# "to create cgroup we need root permission"
	sudo go test -v -race pkg/jobWorker/*.go -run "^Test_Job"

run_server:
	#GOOS=linux GOARCH=amd64 	# linux
	#GOOS=darwin GOARCH=arm64 	# Mac OS (Apple Silicon)
	#GOOS=windows GOARCH=amd64 	# Windows
	go build -o ./jwsrv server/*.go
	sudo ./jwsrv -port 8080

run_client_test:
	#GOOS=linux GOARCH=amd64 	# linux
	#GOOS=darwin GOARCH=arm64 	# Mac OS (Apple Silicon)
	#GOOS=windows GOARCH=amd64 	# Windows
	go build -o ./jwcli cli/main.go

	# run simple job
	./jwcli --host 'localhost:8080'\
			--ca-cert './certs/ca-cert.pem' \
			--client-cert './certs/client-1-cert.pem' \
			--client-key './certs/client-1-key.pem' \
			start \
			--cpu 0.5 \
			--memory 1000000000 \
			--io 10000000 \
			--command 'echo' 'hello world'

	# run short lived job
	#./jwcli --host 'localhost:8080'\
# 		--ca-cert 'certs/ca-cert.pem' \
# 		--client-cert 'certs/client-1-cert.pem' \
# 		--client-key 'certs/client-1-key.pem' \
# 		start \
# 		--cpu 0.5 \
# 		--memory 1000000000 \
# 		--io 10000000 \
# 		--command '/bin/bash' '-c' 'while :; do echo thinking; sleep 1; done'

 	# sleep for 5 second to let job generate some output


