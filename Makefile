generate_certificate:
	sh generate_certs.sh

generate_grpc_code:
	 protoc --go_out=. \
	 		--go_opt=paths=source_relative \
	 		--go-grpc_out=require_unimplemented_servers=false:. \
	 		--go-grpc_opt=paths=source_relative \
	 		./pkg/proto/jobWorker.proto

run_server:
	go run server/*.go -port 8080

build: generate_certificate generate_grpc_code