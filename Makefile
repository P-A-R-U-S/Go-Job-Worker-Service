run_test_namespace:
	sudo suse
	go mod tidy
	gofmt -w pkg/jobWorker/namespaces/*.gols
	go test -v -race pkg/jobWorker/namespaces/*.go -run "^Test_CGroup"

run_test_output:
	go mod tidy
	gofmt -w pkg/jobWorker/*.go
	go test -v -race pkg/jobWorker/*.go -run "^Test_CommandOutput"
	go test -v -race pkg/jobWorker/*.go -run "^Test_OutputReadCloser"

run_test_job:
	go mod tidy
	gofmt -w pkg/jobWorker/*.go
	go test -v -race pkg/jobWorker/*.go -run "^Test_CommandOutput"