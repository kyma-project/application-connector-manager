package reconciler

import (
	"fmt"

	"github.com/kyma-project/application-connector-manager/api/v1alpha1"
	"github.com/kyma-project/application-connector-manager/pkg/unstructured"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"golang.org/x/exp/slices"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type craMatcher struct {
	expected   craDTO
	actualEnvs []corev1.EnvVar
	expectEnvs []corev1.EnvVar
}

func haveRuntimeAgentDefaults(v craDTO) types.GomegaMatcher {
	return &craMatcher{expected: v}
}

func (m *craMatcher) Match(actual any) (success bool, err error) {
	u, ok := actual.(unstructured.Unstructured)
	if !ok {
		return false, fmt.Errorf("stateFnNameMatcher expects unstructured.Unstructured")
	}

	var actualDeployment appsv1.Deployment
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &actualDeployment); err != nil {
		return false, fmt.Errorf("conversion error: %w", err)
	}

	index := slices.IndexFunc(actualDeployment.Spec.Template.Spec.Containers, func(c corev1.Container) bool {
		return c.Name == "compass-runtime-agent"
	})
	if index == -1 {
		return false, fmt.Errorf("compass-runtime-agent container not found")
	}
	// create expected command argumets
	m.expectEnvs = []corev1.EnvVar{
		{
			Name:  v1alpha1.EnvRuntimeAgnetAppRuntimeConsoleURL,
			Value: fmt.Sprintf("https://console.%s", m.expected.Domain),
		},
		{
			Name:  v1alpha1.EnvRuntimeAgentAppRuntimeEventsURL,
			Value: fmt.Sprintf("https://gateway.%s", m.expected.Domain),
		},
	}
	m.actualEnvs = actualDeployment.Spec.Template.Spec.Containers[index].Env
	containsArgs := gomega.ContainElements(m.expectEnvs)
	return containsArgs.Match(m.actualEnvs)
}

func (m *craMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%s\nto be contain\n\t%s", m.actualEnvs, m.expectEnvs)
}

func (m *craMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%s\nnot to contain\n\t%s", m.actualEnvs, m.expectEnvs)
}
