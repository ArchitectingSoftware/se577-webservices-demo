package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/soheilhy/cmux"

	bc "drexel.edu/bc-service/grpc/BCGrpc"
	BcServer "drexel.edu/bc-service/grpc/BCServer"
)

var (
	servicePort = flag.Int("servicePort", 9905, "The combined server port - default is 9990")
	//httpPort    = flag.Int("httpPort", 9901, "The HTTP server port - default is 9991")
	//tlsFlg      = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
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

func StartGrpcServer(l net.Listener, withTLS bool) {
	var s *grpc.Server

	if withTLS {
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

	log.Printf("GRPC server listening at %v", l.Addr())
	if err := s.Serve(l); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func StartHttpServer(l net.Listener) {
	r := gin.Default()
	r.Use(cors.Default())
	r.GET("/bc", BcServer.BlockWebSolver)
	r.RunListener(l)
}

func GetTlsListener() (net.Listener, error) {
	cer, err := tls.LoadX509KeyPair(*certFile, *keyFile)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	config := &tls.Config{Certificates: []tls.Certificate{cer}}
	ln, err := tls.Listen("tcp", fmt.Sprintf(":%d", *servicePort), config)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return ln, nil
}

func main() {
	flag.Parse()

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", *servicePort))
	defer l.Close()
	if err != nil {
		log.Fatal(err)

	}

	tcpm := cmux.New(l)

	// Declare the match for different services required.
	// Match connections in order:
	// First grpc, then HTTP, and otherwise grpcTls.
	grpcL := tcpm.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))
	httpL := tcpm.Match(cmux.HTTP1Fast())
	grpcTls := tcpm.Match(cmux.Any())

	go StartHttpServer(httpL)
	//for this demo we are starting 2 GRPC servers, one that can handle insecure HTTP2 connectios over localhost
	//for testing and one that requires TLS, which is the preferred way to do GRPC
	go StartGrpcServer(grpcL, false)
	go StartGrpcServer(grpcTls, true)

	log.Println("grpc server started.")
	log.Println("http server started.")
	log.Println("tls:grpc server started")
	log.Println("Server listening on ", *servicePort)

	// Start cmux serving.
	if err := tcpm.Serve(); !strings.Contains(err.Error(),
		"use of closed network connection") {
		log.Fatal(err)
	}
}
