package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

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

type hashResult struct {
	found bool
	nonce uint64
	hash  string
}

var (
	portNum = flag.String("port", "9095", "Port number where server will listen")
)

func ExceptionGenerator(exception bool, exit bool) {
	if exception {
		fmt.Println("Simulating an exception via a panic")
		panic(999)
	} else if exit {
		fmt.Println("Simulating a major crash")
		os.Exit(999)
	}
}

func bcHandler(ctx context.Context, goId uint64, lowerIndex uint64,
	upperIndex uint64, baseString string, complexityPrefix string, resChannel chan<- hashResult) {

	var hashBuffer bytes.Buffer

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
		// println("XXX "+hashBuffer.String()+" "+ blockHashString)
		if strings.HasPrefix(blockHashString, complexityPrefix) {
			resChannel <- hashResult{found: true, nonce: i, hash: blockHashString}
			break
		}
	}
	resChannel <- hashResult{found: false, nonce: upperIndex, hash: nullHash}
}

func ghProxy(c *gin.Context) {
	remote, err := url.Parse("https://api.github.com")
	if err != nil {
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.Director = func(req *http.Request) {
		req.Header = c.Request.Header
		req.Host = remote.Host
		req.URL.Scheme = remote.Scheme
		req.URL.Host = remote.Host
		req.URL.Path = c.Param("ghapi")
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}

func ghSecureProxy(c *gin.Context) {
	remote, err := url.Parse("https://api.github.com")
	if err != nil {
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(remote)
	//note your github personal key should be in the GITHUB_ACCESS_TOKEN environment
	//variable, im using a helper, see main(), and a .env file
	//
	//Still requires debugging
	authHeader := "Token " + os.Getenv(("GITHUB_ACCESS_TOKEN"))

	w := c.Writer
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	//w.Header().Set("Authorization", authHeader)

	proxy.Director = func(req *http.Request) {
		req.Header = c.Request.Header
		req.Header.Set("Authorization", authHeader)
		req.Header.Set("Access-Control-Allow-Origin", "*")

		req.Host = remote.Host
		req.URL.Scheme = remote.Scheme
		req.URL.Host = remote.Host
		req.URL.Path = c.Param("ghapi")

		log.Println(req.Header)
		return
	}

	proxy.ServeHTTP(w, c.Request)
}

func main() {

	println("STARTING")
	err := godotenv.Load()
	if err != nil {
		log.Panicln("Unable to load the .env file, proxy operations might not work")
	} else {
		log.Println("Environment file loaded!")
	}

	flag.Parse()
	r := gin.Default()
	r.Use(cors.Default())

	r.GET("/gh/*ghapi", ghProxy)
	r.GET("/ghsecure/*ghapi", ghSecureProxy)

	r.GET("/bc3", func(c *gin.Context) {
		q := c.Query("q") //query data
		p := c.Query("p") //parent hash
		b := c.Query("b") //block id
		x := c.Query("x") //max iterations
		m, _ := strconv.ParseUint(c.Query("m"), 10, 64)
		cr, _ := strconv.ParseBool(c.Query("crash"))
		ex, _ := strconv.ParseBool(c.Query("exception"))

		println("m2 = " + strconv.FormatUint(m, 10))
		println("q = " + q + "p = " + p + "b = " + b + "x = " + x + "m = " + strconv.FormatUint(m, 10))

		//Handle Default Values
		if m == 0 {
			m = 500000 //default value
		}
		if x == "" {
			x = "000"
		}

		//simulate a crash or an exception if one was indicated
		ExceptionGenerator(ex, cr)

		ctx, cancel := context.WithCancel(context.Background())

		baseHashString := b + q + p //All hashes will have these things followed by the nonce

		res := make(chan hashResult)
		startTime := time.Now()

		totalGoRoutines := uint64(runtime.NumCPU() - 1)
		fmt.Println("Num CPUs = ", runtime.NumCPU())
		window := m / totalGoRoutines

		for c := uint64(0); c < totalGoRoutines; c++ {
			lb := c * window
			ub := lb + window - 1
			go bcHandler(ctx, c, lb, ub, baseHashString, x, res)
		}

		var asyncDone hashResult
		for c := uint64(0); c < totalGoRoutines; c++ {
			asyncDone = <-res
			if asyncDone.found == false {
				continue
			} else {
				break
			}
		}

		cancel()
		durationTime := time.Now().Sub(startTime)

		c.JSON(200, gin.H{
			"query":           q,
			"blockHash":       string(asyncDone.hash),
			"nonce":           asyncDone.nonce,
			"executionTimeMs": durationTime.Nanoseconds() / 1e6, //convert to ms
			"found":           asyncDone.found,
			"parentHash":      p,
			"blockId":         b,
		})

	})

	r.GET("/bc2", func(c *gin.Context) {
		q := c.Query("q") //query data
		p := c.Query("p") //parent hash
		b := c.Query("b") //block id
		x := c.Query("x") //max iterations
		m, _ := strconv.ParseUint(c.Query("m"), 10, 64)
		cr, _ := strconv.ParseBool(c.Query("crash"))
		ex, _ := strconv.ParseBool(c.Query("exception"))

		println("m2 = " + strconv.FormatUint(m, 10))
		println("q = " + q + "p = " + p + "b = " + b + "x = " + x + "m = " + strconv.FormatUint(m, 10))

		//Handle Default Values
		if m == 0 {
			m = 500000 //default value
		}
		if x == "" {
			x = "000"
		}

		//simulate a crash or an exception if one was indicated
		ExceptionGenerator(ex, cr)

		solutionBlock := BCBlock{}

		var hashBuffer bytes.Buffer
		baseHashString := b + q + p //All hashes will have these things followed by the nonce

		startTime := time.Now()

		println("debug")
		observable := rxgo.Range(0, int(m), rxgo.WithBufferedChannel(100)).
			Filter(func(item interface{}) bool {
				i := item.(int)

				hashBuffer.Reset()
				hashBuffer.WriteString(baseHashString)
				hashBuffer.WriteString(strconv.FormatUint(uint64(i), 10))

				shash := sha256.Sum256(hashBuffer.Bytes())
				blockHashString := hex.EncodeToString(shash[:])
				if strings.HasPrefix(blockHashString, x) {
					println("****Found it! - ", i, blockHashString)
					return true
				} else {
					return false
				}
			}).First()

		for result := range observable.Observe() {
			resultNonce := 0
			resultFound := false

			if result.Error() {
				//println("not found")
				resultNonce = int(m)
				resultFound = false
			} else {
				resultNonce = result.V.(int)
				resultFound = true
				//fmt.Printf("observable success %v", result.V)
			}

			finalHashString := b + q + p + strconv.FormatUint(uint64(resultNonce), 10)
			hashBuffer.Reset()
			hashBuffer.WriteString(finalHashString)
			finalHash := sha256.Sum256(hashBuffer.Bytes())
			finalBlockHashString := hex.EncodeToString(finalHash[:])

			solutionBlock = BCBlock{
				BlockHash:  finalBlockHashString,
				Nonce:      uint64(resultNonce),
				Found:      resultFound,
				ParentHash: p,
				BlockId:    b,
			}
		}

		durationTime := time.Now().Sub(startTime)

		c.JSON(200, gin.H{
			"query":           q,
			"blockHash":       string(solutionBlock.BlockHash),
			"nonce":           solutionBlock.Nonce,
			"executionTimeMs": durationTime.Nanoseconds() / 1e6, //convert to ms
			"found":           solutionBlock.Found,
			"parentHash":      solutionBlock.ParentHash,
			"blockId":         solutionBlock.BlockId,
		})
	})

	r.GET("/bc", func(c *gin.Context) {
		q := c.Query("q") //query data
		p := c.Query("p") //parent hash
		b := c.Query("b") //block id
		x := c.Query("x") //max iterations
		m, _ := strconv.ParseUint(c.Query("m"), 10, 64)
		cr, _ := strconv.ParseBool(c.Query("crash"))
		ex, _ := strconv.ParseBool(c.Query("exception"))

		println("m1 = " + strconv.FormatUint(m, 10))
		println("q = " + q + "p = " + p + "b = " + b + "x = " + x + "m = " + strconv.FormatUint(m, 10))

		//Handle Default Values
		if m == 0 {
			m = 500000 //default value
		}
		if x == "" {
			x = "000"
		}

		//simulate a crash or an exception if one was indicated
		ExceptionGenerator(ex, cr)

		solutionBlock := BCBlock{}

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
				solutionBlock = BCBlock{
					BlockHash:  blockHashString,
					Nonce:      i,
					Found:      true,
					ParentHash: p,
					BlockId:    b,
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
			solutionBlock = BCBlock{
				BlockHash:  badBlockHashString,
				Nonce:      uint64(m),
				Found:      false,
				ParentHash: p,
				BlockId:    b,
			}
		}

		durationTime := time.Now().Sub(startTime)

		c.JSON(200, gin.H{
			"query":           q,
			"blockHash":       string(solutionBlock.BlockHash),
			"nonce":           solutionBlock.Nonce,
			"executionTimeMs": durationTime.Nanoseconds() / 1e6, //convert to ms
			"found":           solutionBlock.Found,
			"parentHash":      solutionBlock.ParentHash,
			"blockId":         solutionBlock.BlockId,
		})
	})

	host := os.Getenv("GO_BC_HOST")
	if len(host) == 0 {
		host = "0.0.0.0"
	}
	port := os.Getenv("GO_BC_PORT")
	if len(port) == 0 {
		port = *portNum
		println("port " + *portNum)
	}

	r.Run(host + ":" + port)
	//r.Run() // listen and serve on 0.0.0.0:8080
}
