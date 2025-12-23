resolve-local:
	GO111MODULE=on go mod vendor -v

.PHONY: test
test: envtest ## Run tests.
	if [ -d "$(ARTIFACTS)" ]; then \
		KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" GOFIPS140=v1.0.0 go test ./... -v -coverprofile $(ARTIFACTS)/filtered.cov; \
	else \
		KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" GOFIPS140=v1.0.0 go test ./... -v; \
	fi


ENVTEST_K8S_VERSION = 1.26.0

LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

ENVTEST ?= $(LOCALBIN)/setup-envtest

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	test -s $(LOCALBIN)/setup-envtest || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest
