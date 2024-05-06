GATEWAY_TEST_IMAGE   = "$(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/gateway-test:$(DOCKER_TAG)"
VALIDATOR_TEST_IMAGE = "$(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/connectivity-validator-test:$(DOCKER_TAG)"
COMPASS_TEST_IMAGE   = "$(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/compass-runtime-agent-test:$(DOCKER_TAG)"
MOCK_APP_IMAGE       = "$(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/mock-app:$(DOCKER_TAG)"

REGISTRY_PORT ?= 5001

publish-gateway-test: image-gateway-test
	docker push $(GATEWAY_TEST_IMAGE)

image-gateway-test:
	docker build -t $(GATEWAY_TEST_IMAGE) -f Dockerfile.gateway .

publish-mock-app: image-mock-app
	docker push $(MOCK_APP_IMAGE)

image-mock-app:
	docker build -t $(MOCK_APP_IMAGE) -f Dockerfile.mockapp .

publish-validator-test: image-validator-test
	docker push $(VALIDATOR_TEST_IMAGE)

image-validator-test:
	docker build -t $(VALIDATOR_TEST_IMAGE) -f Dockerfile.connectivity-validator .

publish-compass-runtime-agent-test: image-compass-runtime-agent-test
	@echo "::group::publish-compass-runtime-agent-test"
	docker push $(COMPASS_TEST_IMAGE)
	@echo "::endgroup::"

.PHONY: image-compass-runtime-agent-test
image-compass-runtime-agent-test:
	@echo "::group::image-compass-runtime-agent-test"
	@docker build -t $(COMPASS_TEST_IMAGE) -f ${PWD}/tests/Dockerfile.compass-runtime-agent .
	@echo "::endgroup::"

.PHONY: compass-runtime-agent-test-image
compass-runtime-agent-test-image: \
	image-compass-runtime-agent-test \
	publish-compass-runtime-agent-test
