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
	kubectl apply -f https://github.com/kyma-project/istio/releases/download/1.25.0/istio-manager.yaml
	kubectl apply -f https://github.com/kyma-project/istio/releases/download/1.25.0/istio-default-cr.yaml
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
	@echo "::group::create-k3d"
	k3d cluster create ${CLUSTER_NAME} \
		--image docker.io/rancher/k3s:v1.35.0-k3s1 \
		--api-port 6550 \
		-p 8080:80@loadbalancer \
		-p 8443:443@loadbalancer \
		-p 9090-9099:9090-9099@loadbalancer \
		--agents 2 \
		--registry-create k3d-${REGISTRY_NAME}:${REGISTRY_PORT}
	kubectl version
	@echo "::endgroup::"

.PHONY: delete-k3d
delete-k3d: ## Delete k3d with kyma CRDs.
delete-k3d: delete-k3d-cluster delete-k3d-registry

.PHONY: delete-k3d-registry
delete-k3d-registry: ## Delete k3d kyma registry.
	-k3d registry delete ${REGISTRY_NAME}

.PHONY: patch-with-istio-crds
patch-with-istio-crds: ## Delete k3d kyma cluster.
	kubectl apply -f ./k3d-patches/patch-istio-crds.yaml

##@ Module management

.PHONY: verify-appcon
verify-appcon: ## Wait for ApplicationConnector CR to be in Ready state.
	@./verify_appcon_status.sh

.PHONY: verify-kyma
verify-kyma: ## Wait for Kyma CR to be in Ready state.
	@./verify_kyma_status.sh

.PHONY: fix-template
fix-template: ## Create template-k3d.yaml based on moduletemplate.yaml with right URLs.
	@cat ${PROJECT_ROOT}/moduletemplate.yaml \
	| sed -e 's/remote/control-plane/g' \
		-e 's/${REGISTRY_PORT}/5000/g' \
	      	-e 's/localhost/k3d-${REGISTRY_NAME}.localhost/g' \
	> ${PROJECT_ROOT}/template-k3d.yaml

.PHONY: enable-module
enable-module: ## Patch application-connector module in the right Kyma CR.
	kubectl patch kymas.operator.kyma-project.io \
		default-kyma \
		-n kyma-system \
		--type=merge \
		--patch-file ./k3d-patches/patch-kyma.yaml

.PHONY: patch-mod-mgr-role
patch-mod-mgr-role: ## Path lifecycle-manager clusterrole.
	kubectl patch clusterrole \
		lifecycle-manager-manager-role \
		--patch-file=./k3d-patches/patch-k3d-lifecycle-manager-clusterrole.yaml

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

.PHONY: clean-charts
clean-charts:
	@rm -rf ${PROJECT_ROOT}/charts

.PHONY: deploy
deploy:
	@make -C ${PROJECT_ROOT} deploy \
		IMG=k3d-${REGISTRY_NAME}:${REGISTRY_PORT}/${MANAGER_IMAGE_NAME}:${MANAGER_IMAGE_TAG}

.PHONY: yq-install
yq-install:
	@make -C ${PROJECT_ROOT} yq
