generate_certificate:
	sh generate_certs.sh

generate_grpc_code:
	 protoc --go_out=. \
	 		--go_opt=paths=source_relative \
	 		--go-grpc_out=require_unimplemented_servers=false:. \
	 		--go-grpc_opt=paths=source_relative \
	 		./pkg/proto/jobWorker.proto

run_build:
	#go build pkg/tls.go
	#go build pgk/proto/*.go
	go build server/*.go
	go server/*.go
	go cli/*.go

run_server:
	go run server/*.go -port 8080

run_client:
	go run cli/*.go \
		--host localost:8080 \
		--ca-cert "../certs/ca-cert.pem" \
		--client-cert "../certs/client-1-cert.pem" \
		--client-key "../certs/client-1-key.pem" \
		start --c "echo $${PATH}"


build: generate_certificate generate_grpc_code