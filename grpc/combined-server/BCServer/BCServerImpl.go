package BCServer

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"strconv"
	"strings"
	"time"

	bc "drexel.edu/bc-service/grpc/BCGrpc"
)

func BlockFinder(req *bc.BcRequest, startIdx uint64) (*bc.BcResponse, error) {
	log.Println("IN Finder Handler")
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
				Query:         q,
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
			Query:         q,
		}
	}

	durationTime := time.Now().Sub(startTime)
	solutionBlock.ExecTimeMs = durationTime.Nanoseconds() / 1e6 //convert to ms

	return solutionBlock, nil
}
