package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/P-A-R-U-S/Go-Job-Worker-Service/pkg/proto"
	"github.com/google/uuid"
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
	ARGUMENT_HOST               = "host"
	ARGUMENT_CA_CERTIFICATE     = "ca-cert"
	ARGUMENT_CLIENT_CERTIFICATE = "client-cert"
	ARGUMENT_CLIENT_PRIVATE_KEY = "client-key"
)

const (
	START_COMMAND_FLAG_CPU                 = "cpu"
	START_COMMAND_FLAG_MEMORY              = "memory"
	START_COMMAND_FLAG_IO_BYTES_PER_SECOND = "io"
	START_COMMAND_FLAG_COMMAND             = "c"
)

var (
	ErrNoAbleToCreateClient = errors.New("not able to create client")
)

func main() {
	a := &cli.App{
		Name:  "Jow Worker Command Line Interface",
		Usage: "Connect to JobWorker Service to run arbitrary Linux command on remote hosts",
		Commands: []*cli.Command{
			{
				Name:  "start",
				Usage: "starting new job",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     START_COMMAND_FLAG_CPU,
						Value:    "0.5",
						Usage:    "approximate number of CPU cores to limit the job",
						Required: true,
					},
					// TODO: in future add format support for - e.g. 100MB, 1GB and etc
					&cli.StringFlag{
						Name:     START_COMMAND_FLAG_MEMORY,
						Value:    "1000000000",
						Usage:    "maximum amount of memory used by the job",
						Required: true,
					},
					&cli.StringFlag{
						Name:     START_COMMAND_FLAG_IO_BYTES_PER_SECOND,
						Value:    "1000000",
						Usage:    "maximum read and write on the device mounted / is mounted on",
						Required: true,
					},
					&cli.StringFlag{
						Name:     START_COMMAND_FLAG_COMMAND,
						Aliases:  []string{"command"},
						Usage:    "command to execute",
						Required: true,
					},
				},
				Action: func(cCtx *cli.Context) error {
					host := cCtx.String(ARGUMENT_HOST)
					caCert := cCtx.String(ARGUMENT_CA_CERTIFICATE)
					clientCert := cCtx.String(ARGUMENT_CLIENT_CERTIFICATE)
					clientKey := cCtx.String(ARGUMENT_CLIENT_PRIVATE_KEY)

					fmt.Printf("connecting to service %s\n", host)
					fmt.Printf("CA Certificate: %s\n", caCert)
					fmt.Printf("Client Certificate: %s\n", clientCert)
					fmt.Printf("Clint Private key: %s\n", clientKey)

					client, err := getClient(host, caCert, clientCert, clientKey)
					if err != nil {
						return ErrNoAbleToCreateClient
					}

					command := cCtx.String(START_COMMAND_FLAG_MEMORY)
					args := []string{}
					cpu := cCtx.Float64(START_COMMAND_FLAG_CPU)
					memory := cCtx.Int64(START_COMMAND_FLAG_MEMORY)
					io := cCtx.Int64(START_COMMAND_FLAG_IO_BYTES_PER_SECOND)

					start(client, command, args, cpu, memory, io)
					return nil
				},
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     ARGUMENT_HOST,
				Value:    "localhost",
				Usage:    "server IP:PORT to connect",
				Required: true,
			},
			&cli.StringFlag{
				Name:     ARGUMENT_CA_CERTIFICATE,
				Usage:    "client certificate authority (CA)",
				Required: true,
			},
			&cli.StringFlag{
				Name:     ARGUMENT_CLIENT_CERTIFICATE,
				Usage:    "client mTLS certificate",
				Required: true,
			},
			&cli.StringFlag{
				Name:     ARGUMENT_CLIENT_PRIVATE_KEY,
				Usage:    "client private key",
				Required: true,
			},
		},
		Action: func(cCtx *cli.Context) error {
			fmt.Printf("connecting to service %s\n", cCtx.String(ARGUMENT_HOST))
			fmt.Printf("CA Certificate: %s\n", cCtx.String(ARGUMENT_CA_CERTIFICATE))
			fmt.Printf("Client Certificate: %s\n", cCtx.String(ARGUMENT_CLIENT_CERTIFICATE))
			fmt.Printf("Clint Private key: %s\n", cCtx.String(ARGUMENT_CLIENT_PRIVATE_KEY))
			return nil
		},
	}

	if err := a.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func start(client proto.JobWorkerClient, command string, args []string, cpu float64, memory int64, io int64) error {
	return nil
}

func status(client proto.JobWorkerClient, command string, args []string, cpu float64, memory int64, io int64) error {
	return nil
}

func stream(client proto.JobWorkerClient, jobId uuid.UUID) error {
	// Initiate the stream with a context that supports cancellation.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	request := &proto.JobRequest{
		Uuid: jobId.String(),
	}
	stream, err := client.Stream(ctx, request)
	if err != nil {
		log.Fatalf("error creating stream: %v", err)
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
		output, err := stream.Recv()
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

func stop(client *proto.JobWorkerClient, command string, args []string, cpu float64, memory int64, io int64) error {
	return nil
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
