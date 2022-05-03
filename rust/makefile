SHELL := /bin/bash

.PHONY: help
help:
	@echo "Usage make <TARGET>"
	@echo ""
	@echo "  Targets:"
	@echo "	   build-server			Build the server executable"
	@echo "	   run-server			Run the server executable"
	@echo "	   docker-server		Containerize the server"
	@echo "	   docker-run			Run server container"
	@echo "	   test-server			Test server via curl command"


.PHONY: build-server
build-server:
	cargo build --release


.PHONY: run-server
run-server:
	cargo run --release

.PHONY: docker-server
docker-server:
	docker build --rm=true -t architecting-software/bc-service-rust .
	docker image prune -f

.PHONY: docker-run
docker-run:
	docker run --rm -p 9099:9099 architecting-software/bc-service-rust

.PHONY: test-server
test-server:
	curl "localhost:9099/bc?q=hello&p=00000000000000000000000000000000&b=18f3a534-3f69-469d-bc22-981a01d12ac2&x=000&m=1000000" | json_pp 