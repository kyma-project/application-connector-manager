CLUSTER_NAME ?= kyma
REGISTRY_PORT ?= 5001
REGISTRY_NAME ?= ${CLUSTER_NAME}-registry

MANAGER_IMAGE_NAME ?= appcon-manager-dev-local
MANAGER_IMAGE_TAG ?= 0.0.1

PROJECT_ROOT ?= ../..

KYMA ?= ${PROJECT_ROOT}/bin/kyma-unstable

##@ General

.PHONY: help
## Display this help.
help:
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: install-istio
install-istio:
	kubectl apply -f https://github.com/kyma-project/istio/releases/latest/download/istio-manager.yaml
	kubectl apply -f https://github.com/kyma-project/istio/releases/latest/download/istio-default-cr.yaml
	kubectl wait -n kyma-system istios/default --for=jsonpath='{.status.state}'=Ready --timeout=300s

.PHONY: create-cra-config
create-cra-config:
	kubectl create secret generic \
		compass-agent-configuration \
		-n kyma-system \
		--from-literal=CONNECT_URL="https://test.me" \
		--from-literal=RUNTIME_ID="test-runtime-id" \
		--from-literal=TENANT="test-tenant" \
		--from-literal=USER="test-user"

.PHONY: run-without-lifecycle-manager
run-without-lifecycle-manager: ## Create k3d cluster and deploy module without the lifecycle-manager.
run-without-lifecycle-manager: \
        kyma \
	create-k3d \
	create-kyma-system-ns \
	install-istio \
	create-cra-config \
	module-image \
	deploy \
	apply-appcon \
	verify-appcon

##@ K3D

.PHONY: create-k3d
create-k3d: ## Create k3d with kyma CRDs.
	${KYMA} provision k3d --registry-port ${REGISTRY_PORT} --name ${CLUSTER_NAME} -p 8080:80@loadbalancer -p 8443:443@loadbalancer --ci

.PHONY: create-k3d
create-k3d: ## Create k3d with kyma CRDs.
	${KYMA} provision k3d --registry-port ${REGISTRY_PORT} --name ${CLUSTER_NAME} -p 8080:80@loadbalancer -p 8443:443@loadbalancer --ci

##@ Module management

.PHONY: verify-appcon
verify-appcon: ## Wait for ApplicationConnector CR to be in Ready state.
	@./verify_appcon_status.sh

.PHONY: enable-module
enable-module: ## Patch application-connector module in the right Kyma CR.
	kubectl patch kymas.operator.kyma-project.io \
		default-kyma \
		-n kyma-system \
		--type=merge \
		--patch-file ./k3d-patches/patch-kyma.yaml

.PHONY: install-module
install-module: ## Kyma alpha deploy with k3d moduletemplate.
	${KYMA} alpha deploy \
		--ci \
		--force-conflicts
	kubectl apply -f ${PROJECT_ROOT}/template-k3d.yaml

.PHONY: apply-appcon
apply-appcon: ## Apply the k3d application-connector CR
	kubectl apply -n kyma-system \
		-f ${PROJECT_ROOT}/hack/common/test_sample_applicationconnector.yaml

.PHONY: create-kyma-system-ns
create-kyma-system-ns: ## Create kyma-system namespace.
	kubectl create ns kyma-system

### Internal Dependencies

.PHONY: kyma
kyma:
	@make -C ${PROJECT_ROOT} kyma

.PHONY: module-image
module-image:
	@make -C ${PROJECT_ROOT} module-image \
		IMG=localhost:${REGISTRY_PORT}/${MANAGER_IMAGE_NAME}:${MANAGER_IMAGE_TAG}

.PHONY: deploy
deploy:
	@make -C ${PROJECT_ROOT} deploy \
		IMG=k3d-${REGISTRY_NAME}:${REGISTRY_PORT}/${MANAGER_IMAGE_NAME}:${MANAGER_IMAGE_TAG}