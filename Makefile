VERSION := 0.1.0

LDFLAGS := -X main.Version=$(VERSION)
GOFLAGS := -ldflags "$(LDFLAGS)"
GOOS ?= $(shell uname | tr A-Z a-z)
GOARCH ?= $(subst x86_64,amd64,$(patsubst i%86,386,$(shell uname -m)))
SUFFIX ?= $(GOOS)-$(GOARCH)
ARCHIVE ?= $(BINARY)-$(VERSION).$(SUFFIX).tar.gz
BINARY := azure_exporter-$(VERSION).$(SUFFIX)

./dist/$(BINARY):
	mkdir -p ./dist
	go build $(GOFLAGS) -o $@

.PHONY: test
test:
	go test -v -race $$(go list ./... | grep -v /vendor/)

.PHONY: lint
lint:
	golint -min_confidence .1 ./azure
	golint -min_confidence .1

.PHONY: clean
clean:
	rm -rf ./dist

.PHONY: format
format:
	find . -name "*.go" |grep -v vendor | xargs goimports -w

.PHONY: docker
docker:
	docker run --rm -v "$$PWD":/go/src/github.com/iamseth/azure_exporter -w /go/src/github.com/iamseth/azure_exporter golang:1.6 bash -c make
