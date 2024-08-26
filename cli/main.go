package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/urfave/cli"
	"google.golang.org/grpc/credentials"
	"log"
	"os"
)

func main() {
	a := cli.NewApp()
	initFlags(a)
	err := a.Run(os.Args)

	if err != nil {
		log.Fatal(err)
	}
}

func initFlags(a *cli.App) {
	a.Name = "Jow Worker Command Line Interface"
	a.Usage = "Connect to JobWorker Service to run arbitrary Linux command on remote hosts"
	a.Email = "ValentynPonomarenko@gmail.com"

	a.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "start",
			Usage: "jw start --client-cert <PATH_TO_CLIENT_CERT> --cpu 0.5 --memory 500000 --io 1000000 --c $(which date)",
		},
		cli.StringFlag{
			Name:  "status",
			Usage: "jw status --client-cert <PATH_TO_CLIENT_CERT> --id <UUID>",
		},
		cli.StringFlag{
			Name:  "stream",
			Usage: "jw stream --client-cert <PATH_TO_CLIENT_CERT> --id <UUID>",
		},
		cli.StringFlag{
			Name:  "stop",
			Usage: "jw stop --client-cert <PATH_TO_CLIENT_CERT> --id <UUID>",
		},
	}

	a.Action = func(c *cli.Context) error {
		var err error

		if len(c.Args()) == 0 || c.IsSet("h") || c.IsSet("help") {
			err = cli.ShowAppHelp(c)
		}

		if c.IsSet("start") {

		}

		if c.IsSet("status") {

		}

		if c.IsSet("stream") {

		}

		if c.IsSet("stop") {

		}
		return err
	}
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
