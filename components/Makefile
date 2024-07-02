CRA_DIR = $(shell pwd)/compass-runtime-agent
VALIDATOR_DIR = $(shell pwd)/central-application-connectivity-validator

.PHONY: test
test: test-cra test-validator

.PHONY: test-cra
test-cra: envtest ## Run tests.
	cd compass-runtime-agent
	if [ -d "$(ARTIFACTS)" ]; then \
		cd $(CRA_DIR) && KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test ./... -v -coverprofile $(ARTIFACTS)/filtered.cov; \
	else \
		cd $(CRA_DIR) && KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test ./... -v; \
	fi

.PHONY: fmt-validator
fmt-validator: ## Run go fmt against code.
	cd $(VALIDATOR_DIR) && go fmt ./...

.PHONY: vet-validator
vet-validator: ## Run go vet against code.
	cd $(VALIDATOR_DIR) && go vet ./...

.PHONY: test-validator
test-validator: fmt-validator vet-validator envtest ## Run tests.
	cd compass-runtime-agent
	if [ -d "$(ARTIFACTS)" ]; then \
		cd $(VALIDATOR_DIR) && KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test ./... -v -coverprofile $(ARTIFACTS)/filtered.cov; \
	else \
		cd $(VALIDATOR_DIR) && KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test ./... -v; \
	fi

ENVTEST_K8S_VERSION = 1.26.0

LOCALBIN ?= $(shell pwd)/dev/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

ENVTEST ?= $(LOCALBIN)/setup-envtest

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	test -s $(LOCALBIN)/setup-envtest || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest
