package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

const maxTries = 1000000
const zeroHash = "0000000000000000000000000000000000000000000000000000000000000000"
const nullHash = "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF"
const cancelHash = "CCCC0000CCCC0000CCCC0000CCCC0000CCCC0000CCCC0000CCCC0000CCCC0000"

type hashResult struct {
	found    bool
	nonce    uint64
	hash     string
	execTime int64
}

//BcResponse is a container for the REST response back to the caller
type BcResponse struct {
	Query           string `json:"query"`
	BlockHash       string `json:"blockHash"`
	Nonce           uint64 `json:"nonce"`
	ExecutionTimeMs int64  `json:"executionTimeMs"`
	Found           bool   `json:"found"`
	ParentHash      string `json:"parentHash"`
	BlockID         string `json:"blockId"`
}

// Response is of type APIGatewayProxyResponse since we're leveraging the
// AWS Lambda Proxy Request functionality (default behavior)
//
// https://serverless.com/framework/docs/providers/aws/events/apigateway/#lambda-proxy-integration
type Response events.APIGatewayProxyResponse

func powSolver(lb uint64, ub uint64, prefix string, complexity string) hashResult {
	var hashBuffer bytes.Buffer
	startTime := time.Now()

	//Use the looping variable to find the nonce
	for i := lb; i < ub; i++ {
		hashBuffer.Reset()
		hashBuffer.WriteString(prefix)
		hashBuffer.WriteString(strconv.FormatUint(i, 10))

		shash := sha256.Sum256(hashBuffer.Bytes())

		blockHashString := hex.EncodeToString(shash[:])
		// println("XXX "+hashBuffer.String()+" "+ blockHashString)
		if strings.HasPrefix(blockHashString, complexity) {
			durationTime := time.Now().Sub(startTime)
			return hashResult{
				found:    true,
				nonce:    i,
				hash:     blockHashString,
				execTime: durationTime.Nanoseconds() / 1e6,
			}
		}
	}

	durationTime := time.Now().Sub(startTime)
	return hashResult{
		found:    false,
		nonce:    ub,
		hash:     nullHash,
		execTime: durationTime.Nanoseconds() / 1e6,
	}
}

func exceptionGenerator(exception bool, exit bool) {
	if exception {
		fmt.Println("Simulating an exception via a panic")
		panic(999)
	} else if exit {
		fmt.Println("Simulating a major crash")
		os.Exit(999)
	}
}

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context, request events.APIGatewayProxyRequest) (Response, error) {
	q := request.QueryStringParameters["q"] //query data
	p := request.QueryStringParameters["p"] //parent hash
	b := request.QueryStringParameters["b"] //block id
	x := request.QueryStringParameters["x"] //max iterations
	m, _ := strconv.ParseUint(request.QueryStringParameters["m"], 10, 64)
	cr, _ := strconv.ParseBool(request.QueryStringParameters["crash"])
	ex, _ := strconv.ParseBool(request.QueryStringParameters["exception"])

	if m == 0 {
		m = 500000 //default value
	}
	if x == "" {
		x = "000"
	}

	exceptionGenerator(ex, cr)

	baseHashString := b + q + p

	res := powSolver(uint64(0), m, baseHashString, x)

	resp := BcResponse{
		Query:           q,
		BlockHash:       res.hash,
		Nonce:           res.nonce,
		ExecutionTimeMs: res.execTime,
		Found:           res.found,
		ParentHash:      p,
		BlockID:         b,
	}

	body, err := json.Marshal(resp)
	if err != nil {
		return Response{}, err
	}

	return Response{
		StatusCode:      200,
		IsBase64Encoded: false,
		Body:            string(body),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}, nil
}

func main() {
	lambda.Start(Handler)
}
