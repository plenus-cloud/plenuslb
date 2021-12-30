.PHONY: help
.DEFAULT_GOAL := help

all-bin: build-controller build-operator

lint-install:
	@echo "  >  Installing code linter"
	@wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.42.1

lint:
	@echo "  >  Linting code"
	@golangci-lint --version
	@golangci-lint cache clean
	@golangci-lint -v run

test-code: ## runs the code test
	@echo "  >  Generating code"
	@go test `go list ./pkg/... | grep -v *pkg/apis | grep -v *pkg/client` -v -coverprofile=coverage.out 

test: test-code ## runs the code test and shows the coverage
	@go tool cover -func=coverage.out
	@go tool cover -html=coverage.out


generate: vendor ## generates all the stubs for CRDs
	@echo "  >  Generating stubs"
	@./hack/update-codegen.sh

vendor: ## creates vendors folder
	@echo "  >  Creating vendors folder"
	@export GO111MODULE=on
	@go mod vendor

build-operator: clean ## build operator binary
	@echo "  >  Building operator locally"
	@export GO111MODULE=on
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/operator ./pkg/cmd/operator/main.go

build-controller: clean ## build controller binary
	@echo "  >  Building controller locally"
	@export GO111MODULE=on
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/controller ./pkg/cmd/controller/main.go

tidy: ## tidy go deps
	@echo "  >  Tidying project dependencies"
	@go mod tidy

install-deps: ## download go deps
	@echo "  >  Downloading project dependencies"
	@go mod download

clean: ## clean buil cache
	@echo "  >  Cleaning build cache"
	@go clean

help:
	@echo
	@echo " Choose a command run in "$(PROJECTNAME)":"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "     \033[36m%-30s\033[0m %s\n", $$1, $$2}'
	