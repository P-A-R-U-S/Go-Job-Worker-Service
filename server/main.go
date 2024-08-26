package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"github.com/P-A-R-U-S/Go-Job-Worker-Service/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
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
	reflection.Register(serviceRegistrar)

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

func loadTLSCredentials(pemClientCACertificate, pemServerCertificate, pemServerPrivateKey string) (credentials.TransportCredentials, error) {
	// load certificate of the CA who signed server's certificate
	pemClientCA, err := os.ReadFile(pemClientCACertificate)
	if err != nil {
		return nil, err
	}

	// load client CA certificate
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemClientCA) {
		return nil, fmt.Errorf("failed to append client CA's certificates")
	}

	// load server certificate and private key
	serverCert, err := tls.LoadX509KeyPair(pemServerCertificate, pemServerPrivateKey)
	if err != nil {
		return nil, err
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
	}

	return credentials.NewTLS(config), nil
}
