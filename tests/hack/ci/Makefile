K3D_URL=https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh
K3S ?= "docker.io/rancher/k3s:v1.29.5-k3s1"

PROJECT_ROOT  ?= ../../..
CLUSTER_NAME  ?= kyma
REGISTRY_PORT ?= 5001
REGISTRY_NAME ?= ${CLUSTER_NAME}-registry

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
		IMG=k3d-${REGISTRY_NAME}:${REGISTRY_PORT}/${MANAGER_IMAGE_NAME}:${MANAGER_IMAGE_TAG}
	@echo "::endgroup::"

.PHONY: install-application-connector
install-application-connector: application-connector-module-image application-connector-deploy patch-api-server-networkpolicy apply-appcon
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
		--port 80:80@loadbalancer \
		--port 443:443@loadbalancer \
		--agents 2 \
		--k3s-arg "--flannel-backend=none@server:*" \
		--k3s-arg "--disable=traefik@server:0" \
		--k3s-arg "--tls-san=host.docker.internal@server:*" \
		--image rancher/k3s:v1.31.7-k3s1 \
		--registry-create k3d-${REGISTRY_NAME}:${REGISTRY_PORT}
	@echo "::endgroup::"

.PHONY: install-calico
install-calico: ## Install Calico CNI for proper NetworkPolicy support.
	@echo "::group::install-calico"
	@echo "Waiting for Kubernetes API server..."
	@until kubectl get nodes >/dev/null 2>&1; do sleep 2; done
	kubectl create -f https://raw.githubusercontent.com/projectcalico/calico/v3.31.3/manifests/operator-crds.yaml
	kubectl create -f https://raw.githubusercontent.com/projectcalico/calico/v3.31.3/manifests/tigera-operator.yaml
	kubectl create -f https://raw.githubusercontent.com/projectcalico/calico/v3.31.3/manifests/custom-resources.yaml
	kubectl rollout status -n kube-system deployment coredns --timeout=300s
	kubectl patch installation default --type=merge \
		-p '{"spec":{"cni":{"binDir":"/var/lib/rancher/k3s/data/cni","confDir":"/var/lib/rancher/k3s/agent/etc/cni/net.d"}}}'
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

.PHONY: patch-api-server-networkpolicy
patch-api-server-networkpolicy: ## Patch NetworkPolicy to use k3d API server port (6443 instead of 443)
	@echo "::group::patch-api-server-networkpolicy"
	@echo "Patching NetworkPolicy to use API server targetPort 6443 for k3d..."
	kubectl patch networkpolicy -n kyma-system \
		application-connector-kyma-project.io--acm-module-to-api-server \
		--type='json' \
		-p='[{"op": "replace", "path": "/spec/egress/0/ports/0/port", "value": 6443}]'
	@echo "::endgroup::"

.PHONY: patch-validator-networkpolicy
patch-validator-networkpolicy: ## Patch NetworkPolicy to allow test namespace to access validator and validator to reach echoserver
	@echo "::group::patch-validator-networkpolicy"
	@echo "Patching NetworkPolicy to allow ingress from test namespace..."
	kubectl patch networkpolicy -n kyma-system \
		application-connector-kyma-project.io--acm-validator-allow-from-istio-ingressgateway \
		--type='json' \
		-p='[{"op": "add", "path": "/spec/ingress/0/from/-", "value": {"namespaceSelector": {"matchLabels": {"kubernetes.io/metadata.name": "test"}}}}]'
	@echo "Patching NetworkPolicy to allow egress to echoserver in test namespace..."
	kubectl patch networkpolicy -n kyma-system \
		application-connector-kyma-project.io--acm-validator-allow-to-eventing \
		--type='json' \
		-p='[{"op": "add", "path": "/spec/egress/-", "value": {"to": [{"namespaceSelector": {"matchLabels": {"kubernetes.io/metadata.name": "test"}}}], "ports": [{"protocol": "TCP", "port": 80}]}}]'
	@echo "::endgroup::"

.PHONY: gateway-tests
gateway-tests: 
	@echo "::group::gateway-tests"
	@make -f ${PROJECT_ROOT}/tests/Makefile.test-application-gateway test
	@echo "::endgroup::"

.PHONY: k3d-gateway-tests
k3d-gateway-tests: create-k3d \
	install-calico \
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
	install-calico \
	install-istio \
	install-application-connector \
	patch-validator-networkpolicy \
	apply-appcon-crd \
	validator-tests

.PHONY: k3d-agent-tests
k3d-agent-tests: create-k3d \
	install-calico \
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
