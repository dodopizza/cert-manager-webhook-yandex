GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
RESET  := $(shell tput -Txterm sgr0)
OS ?= $(shell go env GOOS)
ARCH ?= $(shell go env GOARCH)

IMAGE_NAME := "dodopizza/cert-manager-webhook-yandex"
IMAGE_TAG := "latest"

.PHONY: all
all: help

build:
	docker buildx build --platform linux/$(ARCH) --tag $(IMAGE_NAME):$(IMAGE_TAG) .

.PHONY: help
help:
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@echo "  ${YELLOW}build                   ${RESET} Build docker image for specific arch"
