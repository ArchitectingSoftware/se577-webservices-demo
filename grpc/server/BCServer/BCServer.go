package BCServer

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"strconv"
	"strings"
	"time"

	bc "drexel.edu/bc-service/grpc/BCGrpc"
)

type Server struct {
	dummy int64
	bc.UnimplementedBCSolverServer
}

func (s *Server) BlockSolverAll(req *bc.BcRequest, stream bc.BCSolver_BlockSolverAllServer) error {
	log.Println("Trying to find all solutions ", req.MaxTries)
	var startPoint = uint64(0)

	//set to true to get into the loop
	var rsp = &bc.BcResponse{
		Found: true,
	}

	for {
		//exit condition
		if !rsp.Found || startPoint >= req.MaxTries {
			log.Println("IN EXIT")
			break
		}

		rsp, err := s.blockFinder(req, startPoint)
		if err != nil {
			log.Println("Error solving block ", err)
			return err
		}

		//only send back if solution found
		if rsp.Found {
			if err := stream.Send(rsp); err != nil {
				return err
			}
		}

		//now update counter, we found a solution, start looking n+1
		startPoint = rsp.Nonce + 1
	}
	return nil
}

func (s *Server) BlockSolver(ctx context.Context, req *bc.BcRequest) (*bc.BcResponse, error) {
	log.Println("Trying to find the first solutions")
	return s.blockFinder(req, uint64(0))
}

func (s *Server) blockFinder(req *bc.BcRequest, startIdx uint64) (*bc.BcResponse, error) {
	q := req.Query
	p := req.ParentBlockId
	b := req.BlockId
	m := req.MaxTries
	x := req.Complexity

	//Handle Default Values
	if m == 0 {
		m = 500000 //default value
	}
	if x == "" {
		x = "000"
	}

	//initial state to allocate the variable
	var solutionBlock = &bc.BcResponse{
		Found: false,
	}

	var hashBuffer bytes.Buffer
	baseHashString := b + q + p //All hashes will have these things followed by the nonce

	startTime := time.Now()
	//Use the looping variable to find the nonce
	for i := startIdx; i < uint64(m); i++ {
		hashBuffer.Reset()
		hashBuffer.WriteString(baseHashString)
		hashBuffer.WriteString(strconv.FormatUint(i, 10))

		shash := sha256.Sum256(hashBuffer.Bytes())

		blockHashString := hex.EncodeToString(shash[:])
		// println("XXX "+hashBuffer.String()+" "+ blockHashString)
		if strings.HasPrefix(blockHashString, x) {
			println("****Found it - ", i, blockHashString)
			solutionBlock = &bc.BcResponse{
				BlockHash:     blockHashString,
				Nonce:         i,
				Found:         true,
				ParentBlockId: p,
				BlockId:       b,
			}
			break
		}
	}

	if solutionBlock.Found == false {
		//recalc hash based on maximum search value m
		finalHashString := b + q + p + strconv.FormatUint(uint64(m), 10)
		hashBuffer.Reset()
		hashBuffer.WriteString(finalHashString)
		badHash := sha256.Sum256(hashBuffer.Bytes())
		badBlockHashString := hex.EncodeToString(badHash[:])
		solutionBlock = &bc.BcResponse{
			BlockHash:     badBlockHashString,
			Nonce:         uint64(m),
			Found:         false,
			ParentBlockId: p,
			BlockId:       b,
		}
	}

	durationTime := time.Now().Sub(startTime)
	solutionBlock.ExecTimeMs = durationTime.Nanoseconds() / 1e6 //convert to ms

	return solutionBlock, nil
}

func (s *Server) Ping(ctx context.Context, req *bc.PingRequest) (*bc.PingResponse, error) {
	retMessage := "HELLO FROM " + req.PingMessage

	resp := &bc.PingResponse{
		PongResponse: retMessage,
	}

	return resp, nil
}
