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

func main() {
	port := flag.Int("port", 8080, "the server port")

	pemClientCACertificate := flag.String("client-ca-cert", "../certs/ca-cert.pem", "the client CA certificate")
	pemServerCertificate := flag.String("server-cert", "../certs/server-cert.pem", "the server certificate")
	pemServerPrivateKey := flag.String("server-key", "../certs/server-key.pem", "the server private key")

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
