package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"flag"
	"fmt"
	"github.com/P-A-R-U-S/Go-Job-Worker-Service/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"log"
	"net"
	"os"
)

var (
	ErrGetingPWD = errors.New("no able to retrieve working directory path")
)

func main() {
	port := flag.Int("port", 8080, "the server port")

	pwd, err := os.Getwd()
	if err != nil {
		os.Exit(1)
		log.Fatalf("cannot execute PWD: %s", ErrGetingPWD)
	}

	// TODO: Hardcode for complete assigment, but in production should be stored on secure storage (Database, AWS Secrets or etc)
	var pathCACert = fmt.Sprintf("%s/certs/ca-cert.pem", pwd)
	var pathServerCert = fmt.Sprintf("%s/certs/server-cert.pem", pwd)
	var pathServerPrivateKey = fmt.Sprintf("%s/certs/server-key.pem", pwd)

	pemClientCACertificate := flag.String("client-ca-cert", pathCACert, "the client CA certificate")
	pemServerCertificate := flag.String("server-cert", pathServerCert, "the server certificate")
	pemServerPrivateKey := flag.String("server-key", pathServerPrivateKey, "the server private key")

	flag.Parse()
	log.Printf("start server on port: %d", *port)

	tlsCredentials, err := loadTLSCredentials(*pemClientCACertificate, *pemServerCertificate, *pemServerPrivateKey)
	if err != nil {
		log.Fatalf("failed to load TLS credentials: %v", err)
	}

	serviceRegistrar := grpc.NewServer(grpc.Creds(tlsCredentials))
	server := NewJobWorkerServer()
	proto.RegisterJobWorkerServer(serviceRegistrar, server)

	address := fmt.Sprintf(":%d", *port)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("cannot create listener: %s", err)
	}
	err = serviceRegistrar.Serve(lis)
	if err != nil {
		log.Fatalf("cannot server to serve: %s", err)
	}
}

var ErrFailedToAppendCA = errors.New("failed to append CA certificate")

func loadTLSCredentials(pemClientCACertificate, pemServerCertificate, pemServerPrivateKey string) (credentials.TransportCredentials, error) {
	// load certificate of the CA who signed server's certificate
	pemClientCA, err := os.ReadFile(pemClientCACertificate)
	if err != nil {
		return nil, fmt.Errorf("failed to load CA certificate: %w", err)
	}

	// load client CA certificate
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemClientCA) {
		return nil, ErrFailedToAppendCA
	}

	// load server certificate and private key
	serverCert, err := tls.LoadX509KeyPair(pemServerCertificate, pemServerPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load server key pair: %w", err)
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
	}

	return credentials.NewTLS(config), nil
}
