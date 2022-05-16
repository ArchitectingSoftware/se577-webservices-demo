SHELL := /bin/bash

.PHONY: help
help:
	@echo "Usage make <TARGET>"
	@echo ""
	@echo "  Targets:"
	@echo "	   build-server			Build the server; typescript to node"
	@echo "	   run					Run the server from javascript code"
	@echo "	   docker-server		Containerize the server"
	@echo "	   docker-run			Run server container"
	@echo "	   kind					Add to kind so that we can setup in ingress controller"
	@echo "	   test-server			Test server via curl"


.PHONY: build-server
build-server:
	tsc

.PHONY: run
run:
	node ./build/server.js

.PHONY: docker-server
docker-server:
	docker build --rm=true -t architecting-software/bc-service-node-fastify .

.PHONY: docker-run
docker-run:
	docker run --rm -p 9094:9094 architecting-software/bc-service-node-fastify

.PHONY: kind
kind:
	kind load docker-image architecting-software/bc-service-node-fastify

.PHONY: test-server
test-server:
	curl "localhost:9094/bc?q=hello&p=00000000000000000000000000000000&b=18f3a534-3f69-469d-bc22-981a01d12ac2&x=000&m=1000000" | json_pp 