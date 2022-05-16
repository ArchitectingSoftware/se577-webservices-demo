package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	ghproxy "drexel.edu/bc-service/go/src/GHProxy"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/reactivex/rxgo/v2"
)

const maxTries = 500000
const zeroHash = "0000000000000000000000000000000000000000000000000000000000000000"
const nullHash = "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF"
const cancelHash = "CCCC0000CCCC0000CCCC0000CCCC0000CCCC0000CCCC0000CCCC0000CCCC0000"

type BCBlock struct {
	BlockHash  string
	Nonce      uint64
	Found      bool
	ParentHash string
	BlockId    string
}

type BCBlockRequest struct {
	Query          string
	ParentBlock    string
	BlockId        string
	MaxTries       uint64
	StartPosition  uint64
	SolutionPrefix string
	CrashSim       bool
	ExceptionSim   bool
}

type hashResult struct {
	found bool
	nonce uint64
	hash  string
}

var (
	portNum = flag.String("port", "9095", "Port number where server will listen")
)

//Process the request parameters
func processRequestParms(c *gin.Context) BCBlockRequest {

	req := BCBlockRequest{}

	req.Query = c.Query("q")          //query data
	req.ParentBlock = c.Query("p")    //parent hash
	req.BlockId = c.Query("b")        //block id
	req.SolutionPrefix = c.Query("x") //max iterations

	//now parse the typed parameters
	m, _ := strconv.ParseUint(c.Query("m"), 10, 64)
	cr, _ := strconv.ParseBool(c.Query("crash"))
	ex, _ := strconv.ParseBool(c.Query("exception"))

	req.StartPosition = 0
	req.MaxTries = m
	req.CrashSim = cr
	req.ExceptionSim = ex

	//Handle Default Values
	if req.MaxTries == 0 {
		req.MaxTries = 500000 //default value
	}
	if req.SolutionPrefix == "" {
		req.SolutionPrefix = "000"
	}

	return req
}

//simulates generating exception
func ExceptionGenerator(exception bool, exit bool) {
	if exception {
		fmt.Println("Simulating an exception via a panic")
		panic(999)
	} else if exit {
		fmt.Println("Simulating a major crash")
		os.Exit(999)
	}
}

//Simple handler, loops to find values
func basicBcHandler(c *gin.Context) {
	reqP := processRequestParms(c)

	//JUST TO BE SAFE, for this StartPos should be zero
	reqP.StartPosition = 0

	//simulate a crash or an exception if one was indicated
	ExceptionGenerator(reqP.ExceptionSim, reqP.CrashSim)

	solutionBlock := BCBlock{}

	var hashBuffer bytes.Buffer
	//All hashes will have these things followed by the nonce
	baseHashString := reqP.BlockId + reqP.Query + reqP.ParentBlock

	startTime := time.Now()
	//Use the looping variable to find the nonce
	for i := reqP.StartPosition; i < reqP.MaxTries; i++ {
		hashBuffer.Reset()
		hashBuffer.WriteString(baseHashString)
		hashBuffer.WriteString(strconv.FormatUint(i, 10))

		shash := sha256.Sum256(hashBuffer.Bytes())

		blockHashString := hex.EncodeToString(shash[:])
		// println("XXX "+hashBuffer.String()+" "+ blockHashString)
		if strings.HasPrefix(blockHashString, reqP.SolutionPrefix) {
			log.Println("****Found it - ", i, blockHashString)
			solutionBlock = BCBlock{
				BlockHash:  blockHashString,
				Nonce:      i,
				Found:      true,
				ParentHash: reqP.ParentBlock,
				BlockId:    reqP.BlockId,
			}
			break
		}
	}

	if solutionBlock.Found == false {
		//recalc hash based on maximum search value m
		finalHashString := baseHashString + strconv.FormatUint(reqP.MaxTries, 10)
		hashBuffer.Reset()
		hashBuffer.WriteString(finalHashString)
		badHash := sha256.Sum256(hashBuffer.Bytes())
		badBlockHashString := hex.EncodeToString(badHash[:])
		solutionBlock = BCBlock{
			BlockHash:  badBlockHashString,
			Nonce:      reqP.MaxTries,
			Found:      false,
			ParentHash: reqP.ParentBlock,
			BlockId:    reqP.BlockId,
		}
	}

	durationTime := time.Now().Sub(startTime)

	c.JSON(200, gin.H{
		"query":           reqP.Query,
		"blockHash":       string(solutionBlock.BlockHash),
		"nonce":           solutionBlock.Nonce,
		"executionTimeMs": durationTime.Nanoseconds() / 1e6, //convert to ms
		"found":           solutionBlock.Found,
		"parentHash":      solutionBlock.ParentHash,
		"blockId":         solutionBlock.BlockId,
	})
}

//Simple handler, loops to find values
func observableBcHandler(c *gin.Context) {
	reqP := processRequestParms(c)

	//JUST TO BE SAFE, for this StartPos should be zero
	reqP.StartPosition = 0

	//simulate a crash or an exception if one was indicated
	ExceptionGenerator(reqP.ExceptionSim, reqP.CrashSim)

	solutionBlock := BCBlock{}

	var hashBuffer bytes.Buffer
	//All hashes will have these things followed by the nonce
	baseHashString := reqP.BlockId + reqP.Query + reqP.ParentBlock

	startTime := time.Now()

	observable := rxgo.Range(int(reqP.StartPosition), int(reqP.MaxTries), rxgo.WithBufferedChannel(100)).
		Filter(func(item interface{}) bool {
			i := item.(int)

			hashBuffer.Reset()
			hashBuffer.WriteString(baseHashString)
			hashBuffer.WriteString(strconv.FormatUint(uint64(i), 10))

			shash := sha256.Sum256(hashBuffer.Bytes())
			blockHashString := hex.EncodeToString(shash[:])
			if strings.HasPrefix(blockHashString, reqP.SolutionPrefix) {
				log.Println("****Found it! - ", i, blockHashString)
				return true
			} else {
				return false
			}
		}).First()

	for result := range observable.Observe() {
		resultNonce := 0
		resultFound := false

		if result.Error() {
			resultNonce = int(reqP.MaxTries)
			resultFound = false
		} else {
			resultNonce = result.V.(int)
			resultFound = true
		}
		finalHashString := baseHashString + strconv.FormatUint(reqP.MaxTries, 10)
		hashBuffer.Reset()
		hashBuffer.WriteString(finalHashString)
		finalHash := sha256.Sum256(hashBuffer.Bytes())
		finalBlockHashString := hex.EncodeToString(finalHash[:])

		solutionBlock = BCBlock{
			BlockHash:  finalBlockHashString,
			Nonce:      uint64(resultNonce),
			Found:      resultFound,
			ParentHash: reqP.ParentBlock,
			BlockId:    reqP.BlockId,
		}
	}

	durationTime := time.Now().Sub(startTime)

	c.JSON(200, gin.H{
		"query":           reqP.Query,
		"blockHash":       string(solutionBlock.BlockHash),
		"nonce":           solutionBlock.Nonce,
		"executionTimeMs": durationTime.Nanoseconds() / 1e6, //convert to ms
		"found":           solutionBlock.Found,
		"parentHash":      solutionBlock.ParentHash,
		"blockId":         solutionBlock.BlockId,
	})
}

//Simple handler, loops to find values
func concurrentBcHandler(c *gin.Context) {
	reqP := processRequestParms(c)

	//JUST TO BE SAFE, for this StartPos should be zero
	reqP.StartPosition = 0

	//simulate a crash or an exception if one was indicated
	ExceptionGenerator(reqP.ExceptionSim, reqP.CrashSim)

	//All hashes will have these things followed by the nonce
	baseHashString := reqP.BlockId + reqP.Query + reqP.ParentBlock

	// Concurrency Setup
	ctx, cancel := context.WithCancel(context.Background())

	//Calculate the concurrency based on the number of CPUs - this is
	//calculation heavy so we would like to get a solver running on each CPU
	res := make(chan hashResult)
	startTime := time.Now()

	totalGoRoutines := uint64(runtime.NumCPU() - 1)
	log.Println("Num CPUs = ", runtime.NumCPU())
	window := reqP.MaxTries / totalGoRoutines

	//Loop to start determined number of goroutines
	for c := uint64(0); c < totalGoRoutines; c++ {
		lb := c * window
		ub := lb + window - 1
		go bcHandler(ctx, c, lb, ub, baseHashString, reqP.SolutionPrefix, res)
	}

	//As the go routines finish, if one finds a solution then
	//you can break out, worse case all goroutines will end
	//and no solutions will be found
	var asyncDone hashResult
	for c := uint64(0); c < totalGoRoutines; c++ {
		asyncDone = <-res
		if asyncDone.found == false {
			continue
		} else {
			break
		}
	}

	//Cancel any running go routines, this would be the case
	//if a routine finishes, but others are still running, this
	//will cause the running routines to end because we already
	//found a soluton
	cancel()

	//Now standard stuff to send back result
	durationTime := time.Now().Sub(startTime)

	c.JSON(200, gin.H{
		"query":           reqP.Query,
		"blockHash":       string(asyncDone.hash),
		"nonce":           asyncDone.nonce,
		"executionTimeMs": durationTime.Nanoseconds() / 1e6, //convert to ms
		"found":           asyncDone.found,
		"parentHash":      reqP.ParentBlock,
		"blockId":         reqP.BlockId,
	})
}

// Does the bc solver calulaction, helper for the concurrent calculator
func bcHandler(ctx context.Context, goId uint64, lowerIndex uint64,
	upperIndex uint64, baseString string, complexityPrefix string, resChannel chan<- hashResult) {

	var hashBuffer bytes.Buffer

	//Loop around on the assigned subset of the overall search space
	for i := lowerIndex; i < upperIndex; i++ {
		select {
		case <-ctx.Done():
			resChannel <- hashResult{found: false, nonce: i, hash: cancelHash}
			return
		default:
			break
		}

		hashBuffer.Reset()
		hashBuffer.WriteString(baseString)
		hashBuffer.WriteString(strconv.FormatUint(i, 10))

		shash := sha256.Sum256(hashBuffer.Bytes())

		blockHashString := hex.EncodeToString(shash[:])

		//Solution found, write that back on the result channel
		if strings.HasPrefix(blockHashString, complexityPrefix) {
			resChannel <- hashResult{found: true, nonce: i, hash: blockHashString}
			break
		}
	}

	//Soltion not found, write that back on the result channel
	resChannel <- hashResult{found: false, nonce: upperIndex, hash: nullHash}
}

func main() {

	//Load the config, will be needed for gh api key
	err := godotenv.Load()
	if err != nil {
		log.Println("Unable to load the .env file, secure proxy operations might not work")
	} else {
		log.Println("Environment file loaded!")
	}

	//setup the API handler
	flag.Parse()
	r := gin.Default()
	r.Use(cors.Default())

	//setup proxies for github
	r.GET("/gh/*ghapi", ghproxy.GhProxy)
	r.GET("/ghsecure/*ghapi", ghproxy.GhSecureProxy)

	//Now the solver options
	r.GET("/bc", basicBcHandler)       //Basic
	r.GET("/bco", observableBcHandler) //Observable Demo
	r.GET("/bcc", concurrentBcHandler) //Concurrent Demo

	//Being able to set the host and port via the environment is helpful if we
	//containerize
	host := os.Getenv("GO_BC_HOST")
	if len(host) == 0 {
		host = "0.0.0.0"
	}
	port := os.Getenv("GO_BC_PORT")
	if len(port) == 0 {
		port = *portNum
		println("port " + *portNum)
	}

	//Finally, lets run
	r.Run(host + ":" + port)
}
