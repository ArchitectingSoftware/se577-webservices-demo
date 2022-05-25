SHELL := /bin/bash

.PHONY: help
help:
	@echo "Usage make <TARGET>"
	@echo ""
	@echo "  Targets:"
	@echo "	   build			Build and package the lambda"
	@echo "	   deploy			Deploy to aws"
	@echo "	   test-local		Test the lambda locally"



.PHONY: build
build:
	sam build

.PHONY: deploy
deploy:
	sam deploy

.PHONY: test-local
test-local:
	sam local start-api
