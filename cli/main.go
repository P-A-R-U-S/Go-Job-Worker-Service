package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/P-A-R-U-S/Go-Job-Worker-Service/pkg/proto"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

const (
	COMMAND_FLAG_HOST                = "host"
	COMMAND_FLAG_CERTIFICATE         = "ca-cert"
	COMMAND_FLAG_CLIENT_CERTIFICATE  = "client-cert"
	COMMAND_FLAG_CLIENT_PRIVATE_KEY  = "client-key"
	COMMAND_FLAG_CPU                 = "cpu"
	COMMAND_FLAG_MEMORY              = "memory"
	COMMAND_FLAG_IO_BYTES_PER_SECOND = "io"
	COMMAND_FLAG_COMMAND             = "c"
	COMMAND_FLAG_ID                  = "id"
)

var (
	ErrNoAbleToCreateClient = errors.New("not able to create client")
)

func main() {
	a := &cli.App{
		Name:  "Jow Worker Command Line Interface",
		Usage: "Connect to JobWorker Service to run arbitrary Linux command on remote hosts",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     COMMAND_FLAG_HOST,
				Value:    "localhost",
				Usage:    "server IP:PORT to connect",
				Required: true,
			},
			&cli.StringFlag{
				Name:     COMMAND_FLAG_CERTIFICATE,
				Usage:    "client certificate authority (CA)",
				Required: true,
			},
			&cli.StringFlag{
				Name:     COMMAND_FLAG_CLIENT_CERTIFICATE,
				Usage:    "client mTLS certificate",
				Required: true,
			},
			&cli.StringFlag{
				Name:     COMMAND_FLAG_CLIENT_PRIVATE_KEY,
				Usage:    "client private key",
				Required: true,
			},
		},
		Action: func(cCtx *cli.Context) error {
			fmt.Printf("connecting to service %s\n", cCtx.String(COMMAND_FLAG_HOST))
			fmt.Printf("CA Certificate: %s\n", cCtx.String(COMMAND_FLAG_CERTIFICATE))
			fmt.Printf("Client Certificate: %s\n", cCtx.String(COMMAND_FLAG_CLIENT_CERTIFICATE))
			fmt.Printf("Clint Private key: %s\n", cCtx.String(COMMAND_FLAG_CLIENT_PRIVATE_KEY))
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:  "start",
				Usage: "starting new job",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     COMMAND_FLAG_CPU,
						Value:    "0.5",
						Usage:    "approximate number of CPU cores to limit the job",
						Required: true,
					},
					// TODO: in future add format support for - e.g. 100MB, 1GB and etc
					&cli.StringFlag{
						Name:     COMMAND_FLAG_MEMORY,
						Value:    "1000000000",
						Usage:    "maximum amount of memory used by the job",
						Required: true,
					},
					&cli.StringFlag{
						Name:     COMMAND_FLAG_IO_BYTES_PER_SECOND,
						Value:    "1000000",
						Usage:    "maximum read and write on the device mounted / is mounted on",
						Required: true,
					},
					&cli.StringFlag{
						Name:     COMMAND_FLAG_COMMAND,
						Aliases:  []string{"command"},
						Usage:    "command to execute",
						Required: true,
					},
				},
				Action: func(cCtx *cli.Context) error {
					client, err := createClient(cCtx)
					if err != nil {
						return ErrNoAbleToCreateClient
					}

					command := cCtx.String(COMMAND_FLAG_MEMORY)
					args := []string{}
					cpu := cCtx.Float64(COMMAND_FLAG_CPU)
					memory := cCtx.Int64(COMMAND_FLAG_MEMORY)
					io := cCtx.Int64(COMMAND_FLAG_IO_BYTES_PER_SECOND)

					start(client, command, args, cpu, memory, io)
					return nil
				},
			},
			{
				Name:  "status",
				Usage: "request job's status",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     COMMAND_FLAG_ID,
						Usage:    "job id",
						Required: true,
					},
				},
				Action: func(cCtx *cli.Context) error {
					client, err := createClient(cCtx)
					if err != nil {
						return ErrNoAbleToCreateClient
					}

					jobId := cCtx.String(COMMAND_FLAG_ID)
					fmt.Printf("job id: %s\n", jobId)

					return status(client, jobId)
				},
			},
			{
				Name:  "stream",
				Usage: "request job's output stream",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     COMMAND_FLAG_ID,
						Usage:    "job id",
						Required: true,
					},
				},
				Action: func(cCtx *cli.Context) error {
					client, err := createClient(cCtx)
					if err != nil {
						return ErrNoAbleToCreateClient
					}

					jobId := cCtx.String(COMMAND_FLAG_ID)
					fmt.Printf("job id: %s\n", jobId)

					return stream(client, jobId)
				},
			},
			{
				Name:  "stop",
				Usage: "stop job execution",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     COMMAND_FLAG_ID,
						Usage:    "job id",
						Required: true,
					},
				},
				Action: func(cCtx *cli.Context) error {
					client, err := createClient(cCtx)
					if err != nil {
						return ErrNoAbleToCreateClient
					}

					jobId := cCtx.String(COMMAND_FLAG_ID)
					fmt.Printf("job id: %s\n", jobId)

					return stop(client, jobId)
				},
			},
		},
	}

	if err := a.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func start(client proto.JobWorkerClient, command string, args []string, cpu float64, memBytes int64, ioBytesPerSecond int64) error {
	// Initiate the stream with a context that supports cancellation.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	request := &proto.JobCreateRequest{
		CPU:              cpu,
		MemBytes:         memBytes,
		IoBytesPerSecond: ioBytesPerSecond,
		Command:          command,
		Args:             args,
	}

	response, err := client.Start(ctx, request)
	if err != nil {
		return fmt.Errorf("error creating stream: %v", err)
	}

	log.Printf("Jod:%s created.", response.Id)
	return nil
}

func status(client proto.JobWorkerClient, jobId string) error {
	job := &proto.JobRequest{
		Id: jobId,
	}

	response, err := client.Status(context.Background(), job)
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	fmt.Printf("Jod:%s has status: %s. ExitCode:%d, ExitReason:%s\n",
		jobId,
		response.GetStatus(),
		response.GetExitCode(),
		response.GetExitReason())

	return nil
}

func stream(client proto.JobWorkerClient, jobId string) error {
	// Initiate the stream with a context that supports cancellation.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	request := &proto.JobRequest{
		Id: jobId,
	}
	response, err := client.Stream(ctx, request)
	if err != nil {
		return fmt.Errorf("error creating stream: %v", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case <-ctx.Done():
			return
		case s := <-sigCh:
			log.Printf("got signal %v, attempting graceful shutdown", s)
			cancel()
		}
	}()

	for {
		output, err := response.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Print(string(output.GetContent()))
				break
			}
			return fmt.Errorf("failed to receive output: %w", err)
		}

		fmt.Print(string(output.GetContent()))
	}

	return nil
}

func stop(client proto.JobWorkerClient, jobId string) error {
	job := &proto.JobRequest{
		Id: jobId,
	}

	response, err := client.Status(context.Background(), job)
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	fmt.Printf("Jod:%s has status: %s. ExitCode:%d, ExitReason:%s\n",
		jobId,
		response.GetStatus(),
		response.GetExitCode(),
		response.GetExitReason())

	return nil
}

func createClient(cCtx *cli.Context) (proto.JobWorkerClient, error) {
	host := cCtx.String(COMMAND_FLAG_HOST)
	caCert := cCtx.String(COMMAND_FLAG_CERTIFICATE)
	clientCert := cCtx.String(COMMAND_FLAG_CLIENT_CERTIFICATE)
	clientKey := cCtx.String(COMMAND_FLAG_CLIENT_PRIVATE_KEY)

	fmt.Printf("connecting to service %s\n", host)
	fmt.Printf("CA Certificate: %s\n", caCert)
	fmt.Printf("Client Certificate: %s\n", clientCert)
	fmt.Printf("Clint Private key: %s\n", clientKey)

	return getClient(host, caCert, clientCert, clientKey)
}

func getClient(host, caCertPath, clientCertPath, clientKeyPath string) (proto.JobWorkerClient, error) {
	tlsCredentials, err := loadTLSCredentials(caCertPath, clientCertPath, clientKeyPath)
	if err != nil {
		log.Fatalf("failed to load TLS credentials: %v", err)
	}

	clientConnection, err := grpc.NewClient(host, grpc.WithTransportCredentials(tlsCredentials))
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	client := proto.NewJobWorkerClient(clientConnection)
	if client == nil {
		return nil, fmt.Errorf("failed to create JobWorkerClient")
	}

	return client, nil
}

func loadTLSCredentials(pemClientCACertificate, pemClientCertificate, pemClientPrivateKey string) (credentials.TransportCredentials, error) {
	// load certificate of the CA who signed server's certificate
	perServerCA, err := os.ReadFile(pemClientCACertificate)
	if err != nil {
		return nil, err
	}

	// load client CA certificate
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(perServerCA) {
		return nil, fmt.Errorf("failed to append client CA's certificates")
	}

	// load server certificate and private key
	clientCert, err := tls.LoadX509KeyPair(pemClientCertificate, pemClientPrivateKey)
	if err != nil {
		return nil, err
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
	}

	return credentials.NewTLS(config), nil
}
