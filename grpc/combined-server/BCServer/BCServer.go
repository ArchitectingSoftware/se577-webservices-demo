package BCServer

import (
	"context"

	bc "drexel.edu/bc-service/grpc/BCGrpc"
)

type Server struct {
	dummy int64
	bc.UnimplementedBCSolverServer
}

func (s *Server) BlockSolver(ctx context.Context, req *bc.BcRequest) (*bc.BcResponse, error) {
	return BlockFinder(req)
}
