package reconciler

import (
	"context"
	"fmt"
	"time"

	"github.com/kyma-project/application-connector-manager/api/v1alpha1"
	commontypes "github.com/kyma-project/application-connector-manager/pkg/common/types"
	modtest "github.com/kyma-project/application-connector-manager/pkg/reconciler/testing"
	"github.com/kyma-project/application-connector-manager/pkg/unstructured"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("ACM sFnUpdate", func() {

	var (
		gvkDeployment    = schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"}
		defaulDomainName = "test123"
	)

	var testData map[string][]unstructured.Unstructured
	updateTimeout := time.Second * 5

	defaultState := &systemState{
		domainName: defaulDomainName,
		instance: v1alpha1.ApplicationConnector{
			Spec: v1alpha1.ApplicationConnectorSpec{
				ApplicationGatewaySpec: v1alpha1.AppGatewaySpec{
					ProxyTimeout:   metav1.Duration{Duration: time.Second * 101},
					RequestTimeout: metav1.Duration{Duration: time.Second * 102},
					LogLevel:       v1alpha1.LogLevelFatal,
				},
				AppConValidatorSpec: v1alpha1.AppConnValidatorSpec{
					LogLevel:  "debug",
					LogFormat: "text",
				},
				DomainName: defaulDomainName,
			},
		},
	}

	testData, err := modtest.LoadTestData(modtest.SfnUpdate)
	Expect(err).Should(BeNil(), fmt.Errorf("unable to extract test data: %s", err))

	ctx, cancel := context.WithTimeout(context.Background(), updateTimeout)
	defer cancel()

	DescribeTable(
		"update state function",
		testUpdate,
		Entry(
			"happy path",
			ctx,
			&fsm{
				Cfg: Cfg{
					Objs: testData[modtest.TdUpdateAcmValid],
					Deps: testData[modtest.TdUpdateDepsValid],
				},
				K8s: K8s{
					Client: fake.NewFakeClient(),
				},
			},
			defaultState,
			testUpdateOptions{
				MatchExpectedErr: BeNil(),
				MatchNextFnState: equalStateFunction(sFnApply),
				StateMatch: map[schema.GroupVersionKind]map[string]types.GomegaMatcher{
					gvkDeployment: {
						"central-application-gateway":                haveAppGatewaySpec(defaultState.instance.Spec.ApplicationGatewaySpec),
						"central-application-connectivity-validator": haveAppConnValidatorSpec(defaultState.instance.Spec.AppConValidatorSpec),
						"compass-runtime-agent":                      haveRuntimeAgentDefaults(craDTO{Domain: defaulDomainName, Replicas: 1}),
					},
					commontypes.Gateway: {
						"kyma-gateway-application-connector": haveDomainNamePropagatedInGateway(fmt.Sprintf("gateway.%s", defaultState.instance.Spec.DomainName)),
					},
					commontypes.VirtualService: {
						"central-application-connectivity-validator": haveDomainNamePropagatedInVirtualService(fmt.Sprintf("gateway.%s", defaultState.instance.Spec.DomainName)),
					},
				},
			},
		),
		Entry(
			"no deployment",
			ctx,
			&fsm{
				K8s: K8s{
					Client: fake.NewFakeClient(),
				},
			},
			defaultState,
			testUpdateOptions{
				MatchExpectedErr: BeNil(),
				MatchNextFnState: equalStateFunction(sFnUpdateStatus(nil, nil)),
			},
		),
	)
})

type testUpdateOptions struct {
	MatchExpectedErr types.GomegaMatcher
	MatchNextFnState types.GomegaMatcher
	StateMatch       map[schema.GroupVersionKind]map[string]types.GomegaMatcher
}

func testUpdate(ctx context.Context, r *fsm, s *systemState, ops testUpdateOptions) {
	sFn, _, err := sFnUpdate(ctx, r, s)
	Expect(err).To(ops.MatchExpectedErr)
	Expect(sFn).To(ops.MatchNextFnState)
	// match state
	for gvk, nameMatcherPairs := range ops.StateMatch {
		for name, match := range nameMatcherPairs {
			u, err := unstructured.IsNamedGroupVersionKind(name, gvk).First(append(r.Objs, r.Deps...))
			Expect(err).Should(BeNil())
			Expect(*u).Should(match)
		}
	}
}
