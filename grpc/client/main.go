package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	bc "drexel.edu/bc-service/grpc/BCGrpc"
)

var (
	port       = flag.Int("port", 9991, "The server port - default is 9990")
	tls        = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	all        = flag.Bool("all", false, "If set finds all of the solutions")
	serverCert = flag.String("cert", "./certs/server.crt", "Fully qualified path to the server certificate file - default ./certs/server.crt")
	query      = flag.String("query", "Hello World!", "Query String to send to GRPC Server")
	complexity = flag.String("complex", "0000", "Block solver complexity")
	totalTries = flag.Uint64("tries", 1000000, "Block solver complexity")
)

/*
type Server struct {
	bc.UnimplementedBCSolverServer
	dummy int64
}
*/

func main() {
	// parse arguments from the command line
	// this lets us define the port for the server
	flag.Parse()

	log.Println("TLS ", *tls)

	var opts []grpc.DialOption
	//now the server options
	if *tls {
		creds, err := credentials.NewClientTLSFromFile(*serverCert, "")

		if err != nil {
			log.Fatalf("Failed to TLS files: %v", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	log.Println("Connecting over Port: ", *port)
	cc, err := grpc.Dial(fmt.Sprintf(":%d", *port), opts...)
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	defer cc.Close()

	cli := bc.NewBCSolverClient(cc)

	if *all {
		doBCSolveAll(cli, *query)
	} else {
		doBCSolve(cli, *query)
	}
}

func doBCSolve(c bc.BCSolverClient, q string) {
	log.Println("Searching for the first solutions")
	req := &bc.BcRequest{
		Query:         q,
		ParentBlockId: "0101010101010101010101010101010101010101010101010101010101010101",
		BlockId:       "9090909090909090909090909090909090909090909090909090909090909090",
		MaxTries:      *totalTries,
		Complexity:    *complexity,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rsp, err := c.BlockSolver(ctx, req)
	if err != nil {
		log.Fatalf("Error %+v\n", err)
	}

	log.Printf("Nonce: %d Hash:  %s\n", rsp.Nonce, rsp.BlockHash)
}

func doBCSolveAll(c bc.BCSolverClient, q string) {
	log.Println("Searching for all solutions")
	req := &bc.BcRequest{
		Query:         q,
		ParentBlockId: "0101010101010101010101010101010101010101010101010101010101010101",
		BlockId:       "9090909090909090909090909090909090909090909090909090909090909090",
		MaxTries:      *totalTries,
		Complexity:    *complexity,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	stream, err := c.BlockSolverAll(ctx, req)
	if err != nil {
		log.Fatalf("Error %+v\n", err)
	}

	for {
		rsp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Error on rpc stream %v", err)
		}
		log.Printf("Nonce: %d Hash:  %s\n", rsp.Nonce, rsp.BlockHash)
	}
}
