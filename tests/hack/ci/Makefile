K3D_URL=https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh

PROJECT_ROOT  ?= ../../..
CLUSTER_NAME  ?= kyma
REGISTRY_PORT ?= 5001
REGISTRY_NAME ?= ${CLUSTER_NAME}-registry
IMG ?= k3d-${REGISTRY_NAME}:${REGISTRY_PORT}/${MANAGER_IMAGE_NAME}:${MANAGER_IMAGE_TAG}

# Operating system architecture
OS_ARCH ?= $(shell uname -m)

# Operating system type
OS_TYPE ?= $(shell uname)

.PHONY: application-connector-module-image
application-connector-module-image:
	@echo "::group::application-connector-module-image"
	@make -C ${PROJECT_ROOT}/hack/common module-image
	@echo "::endgroup::"

.PHONY: application-connector-deploy
application-connector-deploy:
	@echo "::group::application-connector-deploy"
	@make -C ${PROJECT_ROOT}/hack/common deploy \
		IMG=${IMG}
	@echo "::endgroup::"

.PHONY: install-application-connector
install-application-connector: application-connector-deploy apply-appcon
	@echo "::group::install-application-connector"
	kubectl wait -n kyma-system \
		applicationconnectors/applicationconnector-sample \
		--for=jsonpath='{.status.state}'=Ready \
		--timeout=300s
	@echo "::endgroup::"

.PHONY: install-istio
install-istio: create-kyma-system-ns
	@echo "::group::install-istio"
	kubectl apply -f https://github.com/kyma-project/istio/releases/latest/download/istio-manager.yaml
	kubectl apply -f https://github.com/kyma-project/istio/releases/latest/download/istio-default-cr.yaml
	kubectl wait -n kyma-system istios/default --for=jsonpath='{.status.state}'=Ready --timeout=300s
	@echo "::endgroup::"

.PHONY: create-kyma-system-ns
create-kyma-system-ns: ## Create kyma-system namespace.
	@echo "::group::create-kyma-system-ns"
	kubectl create ns kyma-system
	kubectl label namespaces kyma-system istio-injection=enabled --overwrite=true
	@echo "::endgroup::"

.PHONY: create-k3d
create-k3d: ## Create k3d with kyma CRDs.
	@echo "::group::create-k3d"
	k3d cluster create ${CLUSTER_NAME} \
		--api-port 6550 \
		-p 8080:80@loadbalancer \
		-p 8443:443@loadbalancer \
		-p 9090-9099:9090-9099@loadbalancer \
		--agents 2 \
		--registry-create k3d-${REGISTRY_NAME}:${REGISTRY_PORT}
	@echo "::endgroup::"

.PHONY: apply-appcon
apply-appcon: ## Apply the k3d application-connector CR
	@echo "::group::apply-appcon"
	@make -C ${PROJECT_ROOT}/hack/common apply-appcon
	@echo "::endgroup::"

.PHONY: apply-appcon-crd
apply-appcon-crd: ## Apply the application-connector CRD
	@echo "::group::apply-appcon-crd"
	kubectl apply -f ${PROJECT_ROOT}/tests/hack/ci/deps/applications.applicationconnector.crd.yaml
	@echo "::endgroup::"

.PHONY: apply-compass-connection-crd
apply-compass-connection-crd: ## Apply the compass-connection CRD
	@echo "::group::apply-compas-connection-crd"
	kubectl apply -f ${PROJECT_ROOT}/tests/hack/ci/deps/compass-connection.crd.yaml
	@echo "::endgroup::"

.PHONY: gateway-tests
gateway-tests: 
	@echo "::group::gateway-tests"
	@make -f ${PROJECT_ROOT}/tests/Makefile.test-application-gateway test
	@echo "::endgroup::"

.PHONY: k3d-gateway-tests
k3d-gateway-tests: create-k3d \
	install-istio \
	install-application-connector \
	apply-appcon-crd \
	gateway-tests

.PHONY: patch-validator
patch-validator:
	@echo "::group::patch-validator"
	@make -f ${PROJECT_ROOT}/tests/Makefile.test-application-conn-validator patch-validator
	@echo "::endgroup::"

.PHONY: patch-compass-runtime-agent
patch-compass-runtime-agent:
	@echo "::group::patch-compass-runtime-agent"
	@make -f ${PROJECT_ROOT}/tests/Makefile.test-compass-runtime-agent patch-compass-runtime-agent
	@echo "::endgroup::"

.PHONY: validator-tests
validator-tests: 
	@echo "::group::validator-tests"
	@make -f ${PROJECT_ROOT}/tests/Makefile.test-application-conn-validator test
	@echo "::endgroup::"

.PHONY: k3d-validator-tests
k3d-validator-tests: patch-validator \
	create-k3d \
	install-istio \
	install-application-connector \
	apply-appcon-crd \
	validator-tests

.PHONY: k3d-agent-tests
k3d-agent-tests: create-k3d \
	install-istio \
	patch-compass-runtime-agent \
	install-application-connector \
	apply-compass-connection-crd \
	apply-appcon-crd \
	compass-runtime-agent-test-image \
	agent-tests

.PHONY: agent-tests
agent-tests:
	@echo "::group::agent-tests"
	@make -f ${PROJECT_ROOT}/tests/Makefile.test-compass-runtime-agent test
	@echo "::endgroup::"

.PHONY: compass-runtime-agent-test-image
compass-runtime-agent-test-image:
	@echo "::group::compass-runtime-agent-test-image"
	@make -C ${PROJECT_ROOT}/tests compass-runtime-agent-test-image \
		DOCKER_PUSH_REPOSITORY=localhost:${REGISTRY_PORT} \
		DOCKER_TAG=002
	@echo "::endgroup::"

.PHONY: install-compass-runtime-agent
install-compass-runtime-agent: compass-runtime-agent-test-image
	@echo "::group::install-compass-runtime-agent"
	@helm template ${PROJECT_ROOT}/tests/hack/ci/resources/charts/compass-runtime-agent \
		-n kyma-system \
		| kubectl apply -f -
	@echo "::endgroup::"
