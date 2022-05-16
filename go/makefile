SHELL := /bin/bash

.PHONY: help
help:
	@echo "Usage make <TARGET>"
	@echo ""
	@echo "  Targets:"
	@echo "	   build-server			Build the server executable"
	@echo "	   run-server			Run the server executable"
	@echo "	   run					Run the server from code"
	@echo "	   docker-server		Containerize the server"
	@echo "	   docker-run			Run server container"
	@echo "	   kind					Add to kind so that we can setup in ingress controller"
	@echo "	   test-server			Test server via curl"


.PHONY: build-server
build-server:
	go build -o ./bin/bcservice ./src

.PHONY: run
run:
	go run src/bcservice.go

.PHONY: run-server
run-server:
	(cd ./bin; ./bcservice)

.PHONY: docker-server
docker-server:
	docker build --rm=true -t architecting-software/bc-service-go .
	docker image prune --filter label=stage=builder

.PHONY: docker-run
docker-run:
	docker run --rm -p 9095:9095 architecting-software/bc-service-go

.PHONY: kind
kind:
	kind load docker-image architecting-software/bc-service-go 
	
.PHONY: test-server
test-server:
	curl "localhost:9095/bc?q=hello&p=00000000000000000000000000000000&b=18f3a534-3f69-469d-bc22-981a01d12ac2&x=000&m=1000000" | json_pp 