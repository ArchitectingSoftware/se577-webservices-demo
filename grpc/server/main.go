package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	bc "drexel.edu/bc-service/grpc/BCGrpc"
	BcServer "drexel.edu/bc-service/grpc/BCServer"
)

var (
	port     = flag.Int("port", 9990, "The server port - default is 9990")
	tls      = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	certFile = flag.String("cert", "./certs/server.crt", "Fully qualified path to the server certificate file - default ./certs/server.crt")
	keyFile  = flag.String("key", "./certs/server.key", "Fully qualified path to the server key file - default ./certs/server.key")
)

/*
type Server struct {
	bc.UnimplementedBCSolverServer
	dummy int64
}
*/

func GenerateTLSApi(pemPath, keyPath string) (*grpc.Server, error) {
	cred, err := credentials.NewServerTLSFromFile(pemPath, keyPath)
	if err != nil {
		return nil, err
	}

	s := grpc.NewServer(
		grpc.Creds(cred),
	)
	return s, nil
}

func GenerateLocalApi() (*grpc.Server, error) {
	s := grpc.NewServer()
	return s, nil
}

func main() {
	// parse arguments from the command line
	// this lets us define the port for the server
	flag.Parse()

	//setup connection from flags
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	var s *grpc.Server

	if *tls {
		log.Println("Starting TLS Server....")
		s, _ = GenerateTLSApi(*certFile, *keyFile)
	} else {
		log.Println("Starting Local (non-TLS) Server....")
		s, _ = GenerateLocalApi()
	}

	serverConfig := &BcServer.Server{}

	bc.RegisterBCSolverServer(s, serverConfig)

	// Register server method (actions the server will do)
	// TODO

	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
