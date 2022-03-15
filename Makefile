GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
RESET  := $(shell tput -Txterm sgr0)
OS ?= $(shell go env GOOS)
ARCH ?= $(shell go env GOARCH)

IMAGE_NAME := "dodopizza/cert-manager-webhook-yandex"
IMAGE_TAG := "latest"
K8S_VERSION=1.21.2

.PHONY: all
all: help

build:
	docker buildx build --platform linux/$(ARCH) --tag $(IMAGE_NAME):$(IMAGE_TAG) .

.PHONY: tidy
tidy:
	go mod tidy -compat=1.17 -v

.PHONY: lint
lint:
	golangci-lint run

.PHONY: prepare
prepare: tidy lint

test-integration: _test/kubebuilder
	TEST_ASSET_ETCD=_test/kubebuilder/bin/etcd \
	TEST_ASSET_KUBE_APISERVER=_test/kubebuilder/bin/kube-apiserver \
	TEST_ASSET_KUBECTL=_test/kubebuilder/bin/kubectl \
	go test -v .

_test/kubebuilder:
	mkdir -p _test/kubebuilder
	curl -sSLo envtest-bins.tar.gz "https://go.kubebuilder.io/test-tools/${K8S_VERSION}/${OS}/${ARCH}"
	tar -C _test/kubebuilder --strip-components=1 -zvxf envtest-bins.tar.gz
	rm envtest-bins.tar.gz

clean: clean-kubebuilder

clean-kubebuilder:
	rm -Rf _test/kubebuilder

.PHONY: help
help:
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@echo "  ${YELLOW}build                     ${RESET} Build docker image for specific arch"
	@echo "  ${YELLOW}tidy                      ${RESET} Run tidy for go module to remove unused dependencies"
	@echo "  ${YELLOW}lint                      ${RESET} Run linters via golangci-lint"
	@echo "  ${YELLOW}prepare                   ${RESET} Run all available checks and generators"
	@echo "  ${YELLOW}test-integration          ${RESET} Run integration test located at main_test.go"
	@echo "  ${YELLOW}cleanup                   ${RESET} Cleanup all temporary artifacts"
