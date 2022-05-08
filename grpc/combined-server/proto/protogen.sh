#!/bin/bash
protoc --go_out=../BCGrpc --go_opt=paths=source_relative \
    --go-grpc_out=../BcGrpc --go-grpc_opt=paths=source_relative \
    bc.proto