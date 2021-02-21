GO_VERSION ?= 1.16
DOCKER_IMAGE ?= ghcr.io/tinyzimmer/go-gst:$(GO_VERSION)

GOPATH ?= $(shell go env GOPATH)
GOBIN  ?= $(GOPATH)/bin
GOLANGCI_VERSION ?= v1.33.0
GOLANGCI_LINT    ?= $(GOBIN)/golangci-lint

PLUGIN_GEN ?= "$(shell go env GOPATH)/bin/gst-plugin-gen"

$(GOLANGCI_LINT):
        curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN) $(GOLANGCI_VERSION)

# TODO: Fails miserably but at least checks that it compiles
lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run -v