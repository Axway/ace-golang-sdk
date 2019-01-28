all: lint test testrace

build: deps
	go build github.com/Axway/ace-golang-sdk/...

clean:
	go clean -i github.com/Axway/ace-golang-sdk/...

deps:
	dep ensure

test: deps
	go test github.com/Axway/ace-golang-sdk/...

testrace: deps
	go test -race -cpu 1,4 -timeout 7m github.com/Axway/ace-golang-sdk/...

lint: ## Lint the files
	@golint -set_exit_status ${GO_PKG_LIST}

.PHONY: \
	all \
	build \
	clean \
	deps \
	test \
	testrace \
	lint \
