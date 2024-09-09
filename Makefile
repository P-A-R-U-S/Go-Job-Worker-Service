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