package application_gateway

import (
	"encoding/json"
	"regexp"
	"strconv"

	"github.com/kyma-project/application-connector-manager/components/central-application-gateway/pkg/apis/applicationconnector/v1alpha1"
	"github.com/pkg/errors"

	test_api "github.com/kyma-project/application-connector-manager/tests/internal/testkit/test-api"
)

func getExpectedHTTPCode(service v1alpha1.Service) (int, error) {
	re := regexp.MustCompile(`\d+`)
	if codeStr := re.FindString(service.Description); len(codeStr) > 0 {
		return strconv.Atoi(codeStr)
	}
	return 0, errors.New("Bad configuration")
}

func gatewayURL(app, service string) string {
	return "http://central-application-gateway.kyma-system:8080/" + app + "/" + service
}

func unmarshalBody(body []byte) (test_api.EchoResponse, error) {
	res := test_api.EchoResponse{}
	err := json.Unmarshal(body, &res)
	return res, err
}
