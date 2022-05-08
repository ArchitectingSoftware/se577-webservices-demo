package BCServer

import (
	"strconv"

	bc "drexel.edu/bc-service/grpc/BCGrpc"
	"github.com/gin-gonic/gin"
)

func BlockWebSolver(c *gin.Context) {
	q := c.Query("q") //query data
	p := c.Query("p") //parent hash
	b := c.Query("b") //block id
	x := c.Query("x") //max iterations
	m, _ := strconv.ParseUint(c.Query("m"), 10, 64)

	println("m1 = " + strconv.FormatUint(m, 10))
	println("q = " + q + "p = " + p + "b = " + b + "x = " + x + "m = " + strconv.FormatUint(m, 10))

	//Handle Default Values
	if m == 0 {
		m = 500000 //default value
	}
	if x == "" {
		x = "000"
	}

	req := &bc.BcRequest{
		Query:         q,
		ParentBlockId: p,
		BlockId:       b,
		Complexity:    x,
		MaxTries:      m,
	}

	resp, err := BlockFinder(req)
	if err != nil {
		//handle error here
	}

	c.JSON(200, gin.H{
		"query":           resp.Query,
		"blockHash":       resp.BlockHash,
		"nonce":           resp.Nonce,
		"executionTimeMs": resp.ExecTimeMs,
		"found":           resp.Found,
		"parentHash":      resp.ParentBlockId,
		"blockId":         resp.BlockId,
	})
}
