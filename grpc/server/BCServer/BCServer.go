package BCServer

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	bc "drexel.edu/bc-service/grpc/BCGrpc"
)

type Server struct {
	dummy int64
	bc.UnimplementedBCSolverServer
}

func (s *Server) BlockSolver(ctx context.Context, req *bc.BcRequest) (*bc.BcResponse, error) {
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

	var solutionBlock *bc.BcResponse

	var hashBuffer bytes.Buffer
	baseHashString := b + q + p //All hashes will have these things followed by the nonce

	startTime := time.Now()
	//Use the looping variable to find the nonce
	for i := uint64(0); i < uint64(m); i++ {
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
