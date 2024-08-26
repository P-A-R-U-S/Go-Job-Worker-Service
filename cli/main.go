package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/P-A-R-U-S/Go-Job-Worker-Service/pkg/proto"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"log"
	"os"
)

const (
	ARGUMENT_HOST               = "host"
	ARGUMENT_CA_CERTIFICATE     = "ca-cert"
	ARGUMENT_CLIENT_CERTIFICATE = "client-cert"
	ARGUMENT_CLIENT_PRIVATE_KEY = "client-key"
)

func main() {
	a := &cli.App{
		Name:  "Jow Worker Command Line Interface",
		Usage: "Connect to JobWorker Service to run arbitrary Linux command on remote hosts",
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

func start(client *proto.JobWorkerClient, command string, args []string, cpu float64, memory int64, io int64) error {
	return nil
}

func status(client *proto.JobWorkerClient, command string, args []string, cpu float64, memory int64, io int64) error {
	return nil
}

func stream(client *proto.JobWorkerClient, command string, args []string, cpu float64, memory int64, io int64) error {
	return nil
}

func stop(client *proto.JobWorkerClient, command string, args []string, cpu float64, memory int64, io int64) error {
	return nil
}

//func initFlags(a *cli.App) {
//	a.Name = "Jow Worker Command Line Interface"
//	a.Usage = "Connect to JobWorker Service to run arbitrary Linux command on remote hosts"
//	a.Email = "ValentynPonomarenko@gmail.com"
//
//	a.Flags = []cli.Flag{
//		cli.StringFlag{
//			Name:  "start",
//			Usage: "start --client-cert <PATH_TO_CLIENT_CERT> --cpu 0.5 --memory 500000 --io 1000000 --c $(which date)",
//		},
//		cli.StringFlag{
//			Name:  "status",
//			Usage: "jw status --client-cert <PATH_TO_CLIENT_CERT> --id <UUID>",
//		},
//		cli.StringFlag{
//			Name:  "stream",
//			Usage: "stream --client-cert <PATH_TO_CLIENT_CERT> --id <UUID>",
//		},
//		cli.StringFlag{
//			Name:  "stop",
//			Usage: "stop --client-cert <PATH_TO_CLIENT_CERT> --id <UUID>",
//		},
//	}
//
//	a.Action = func(c *cli.Context) error {
//		var err error
//
//		if len(c.Args()) == 0 || c.IsSet("h") || c.IsSet("help") {
//			err = cli.ShowAppHelp(c)
//		}
//
//		if c.IsSet("start") {
//			client, err := getClient(c.)
//		}
//
//		if c.IsSet("status") {
//
//		}
//
//		if c.IsSet("stream") {
//
//		}
//
//		if c.IsSet("stop") {
//
//		}
//		return err
//	}
//}

func getClient(caCertPath, clientCertPath, clientKeyPath string) (proto.JobWorkerClient, error) {
	// TODO: make the address configurable
	host := "localhost"
	port := "8080"

	tlsCredentials, err := loadTLSCredentials(caCertPath, clientCertPath, clientKeyPath)
	if err != nil {
		log.Fatalf("failed to load TLS credentials: %v", err)
	}

	address := fmt.Sprintf("%d:%d", host, port)
	clientConnection, err := grpc.NewClient(address, grpc.WithTransportCredentials(tlsCredentials))
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
