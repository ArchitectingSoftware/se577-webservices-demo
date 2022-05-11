package BCServer

import (
	"context"
	"log"

	bc "drexel.edu/bc-service/grpc/BCGrpc"
)

type Server struct {
	dummy int64
	bc.UnimplementedBCSolverServer
}

func (s *Server) BlockSolver(ctx context.Context, req *bc.BcRequest) (*bc.BcResponse, error) {
	log.Println("Trying to find the first solutions")
	return BlockFinder(req, uint64(0))
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

		rsp, err := BlockFinder(req, startPoint)
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
