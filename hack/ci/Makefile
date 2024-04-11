PROJECT_ROOT ?= ../..
PROJECT_COMMON ?= ../common

K3D_URL=https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh
DEFAULT_K3D_VERSION=v5.6.0

include ${PROJECT_ROOT}/.env

ifndef MODULE_SHA
    MODULE_SHA = ${shell git rev-parse --short HEAD}
endif

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: integration-test
integration-test: ## Run integration tests
	@echo "\n Not implemented \n"

.PHONY: k3d-integration-test
k3d-integration-test: ## Run integration tests.
k3d-integration-test: run-without-lifecycle-manager integration-test

.PHONY: k3d-k8s-compatibility-test
k3d-k8s-compatibility-test: ## K8s compatibility tests not implemented yet.
	@echo "k8s compatibility tests not implemented yet"

.PHONY: hyperscalers-compatibility-test
hyperscalers-compatibility-test: ## Hyperscalers compatibility tests not implemented yet.
	@echo "hyperscalers compatibility tests not implemented yet"

.PHONY: k3d-upgrade-test
k3d-upgrade-test: ## Installs application-connector from latest and upgrades to local.
k3d-upgrade-test:
	@echo "upgrade tests not implemented yet"

.PHONY: render-manifest
render-manifest:
	@make -C ${PROJECT_ROOT} render-manifest

.PHONY: install-k3d
install-k3d:
	curl --silent --fail ${K3D_URL} | TAG=${DEFAULT_K3D_VERSION} bash

.PHONY: run-without-lifecycle-manager
run-without-lifecycle-manager: install-k3d
	@make -C ${PROJECT_COMMON} run-without-lifecycle-manager

