package main

import (
	"fmt"
	"github.com/P-A-R-U-S/Go-Job-Worker-Service/pkg/proto"
	"google.golang.org/grpc"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"testing"
)

var (
	host             = "localhost:8080"
	caCert           = "%s/certs/ca-cert.pem"
	clientCert       = "%s/certs/client-1-cert.pem"
	clientPrivateKey = "%s/certs/client-1-key.pem"
)

func createTestClient() (proto.JobWorkerClient, *grpc.ClientConn, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, nil, fmt.Errorf("cannot retrive rooted path name")
	}

	pwd = strings.Replace(pwd, "/cli", "", -1)

	return getClient(host,
		fmt.Sprintf(caCert, pwd),
		fmt.Sprintf(clientCert, pwd),
		fmt.Sprintf(clientPrivateKey, pwd))
}

func captureOutput(f func() error) (string, error) {
	orig := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	log.SetOutput(os.Stdout)
	err := f()
	os.Stdout = orig
	log.SetOutput(os.Stdout)
	w.Close()
	out, _ := io.ReadAll(r)
	return string(out), err
}

func extractJobID(out string) string {
	r := regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}")
	matches := r.FindStringSubmatch(out)
	return matches[0]
}

// positive case

// Use Case #1: Starting simple job, check job status and let the job complete.
func Test_Client_Should_start_simple_job_and_return_correct_status(t *testing.T) {
	var err error
	var client proto.JobWorkerClient
	var conn *grpc.ClientConn
	var stdOut string

	t.Parallel()

	client, conn, err = createTestClient()
	if err != nil {
		t.Error(ErrNoAbleToCreateClient)
	}
	defer conn.Close()

	command := "echo"
	args := []string{"hello", "world"}
	cpu := 1.0
	memory := int64(1000000000)
	io := int64(10000000)

	testFunction := func() error {
		err = start(client, command, args, cpu, memory, io)
		if err != nil {
			t.Error(ErrNoAbleToCreateClient)
		}
		return err
	}

	stdOut, err = captureOutput(testFunction)
	if err != nil {
		t.Error(err)
	}

	jobId := extractJobID(stdOut)
	if jobId == "" {
		t.Error("should return jobId")
	}
}
