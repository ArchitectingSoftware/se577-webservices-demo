package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	bc "drexel.edu/bc-service/grpc/BCGrpc"
	BcServer "drexel.edu/bc-service/grpc/BCServer"
)

var (
	grpcPort = flag.Int("grpcPort", 9900, "The GRPC server port - default is 9990")
	httpPort = flag.Int("httpPort", 9901, "The HTTP server port - default is 9991")
	tls      = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	certFile = flag.String("cert", "./certs/server.crt", "Fully qualified path to the server certificate file - default ./certs/server.crt")
	keyFile  = flag.String("key", "./certs/server.key", "Fully qualified path to the server key file - default ./certs/server.key")
)

func GenerateTLSGrpc(pemPath, keyPath string) (*grpc.Server, error) {
	cred, err := credentials.NewServerTLSFromFile(pemPath, keyPath)
	if err != nil {
		return nil, err
	}

	s := grpc.NewServer(
		grpc.Creds(cred),
	)
	return s, nil
}

func GenerateGrpc() (*grpc.Server, error) {
	s := grpc.NewServer()
	return s, nil
}

func StartGrpcServer() {
	//setup connection from flags
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *grpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	var s *grpc.Server

	if *tls {
		log.Println("Starting TLS Server....")
		s, _ = GenerateTLSGrpc(*certFile, *keyFile)
	} else {
		log.Println("Starting Local (non-TLS) Server....")
		s, _ = GenerateGrpc()
	}

	serverConfig := &BcServer.Server{}

	bc.RegisterBCSolverServer(s, serverConfig)

	// Register server method (actions the server will do)
	// TODO

	log.Printf("GRPC server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func StartHttpServer() {
	r := gin.Default()
	r.Use(cors.Default())
	r.GET("/bc", BcServer.BlockWebSolver)
	r.Run(fmt.Sprintf(":%d", *httpPort))
}

func main() {
	flag.Parse()
	go StartGrpcServer()
	StartHttpServer()
}
