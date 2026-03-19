package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/kyma-project/application-connector-manager/api/v1alpha1"
	"github.com/kyma-project/application-connector-manager/pkg/unstructured"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	istio "istio.io/client-go/pkg/apis/networking/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"k8s.io/utils/ptr"
)

const (
	appGatewayDeploymentName      = "central-application-gateway"
	appConValidatorDeploymentName = "central-application-connectivity-validator"
	compassRtAgentDeploymentName  = "compass-runtime-agent"
	istioNamespace                = "istio-system"
	defaultTestTimeout            = 60 * time.Second
	testDomainName                = "testme"
)

var _ = Describe("ApplicationConnector controller", func() {

	appConWithNetworkPolicies := applicationConnector("app-conn-with-np", "kyma-system", v1alpha1.ApplicationConnectorSpec{
		ApplicationGatewaySpec: v1alpha1.AppGatewaySpec{
			LogLevel: v1alpha1.LogLevel("info"),
		},
		AppConValidatorSpec: v1alpha1.AppConnValidatorSpec{
			LogLevel:  v1alpha1.LogLevel("info"),
			LogFormat: v1alpha1.LogFormat("json"),
		},
		NetworkPoliciesEnabled: true,
	})

	appConWithoutNetworkPolicies := applicationConnector("app-conn-without-np", "kyma-system", v1alpha1.ApplicationConnectorSpec{
		ApplicationGatewaySpec: v1alpha1.AppGatewaySpec{
			LogLevel: v1alpha1.LogLevel("info"),
		},
		AppConValidatorSpec: v1alpha1.AppConnValidatorSpec{
			LogLevel:  v1alpha1.LogLevel("info"),
			LogFormat: v1alpha1.LogFormat("json"),
		},
		NetworkPoliciesEnabled: false,
	})

	Context("When creating fresh instance", func() {
		DescribeTable(
			"The application-connector is created properly with given specification (network policies enabled)",
			// the table function that will be executed for each entry
			testInstanceCreate,
			Entry("with default arguments", defaultTestTimeout, appConWithNetworkPolicies),
		)
		DescribeTable(
			"The application-connector is created properly with given specification (network policies disabled)",
			// the table function that will be executed for each entry
			testInstanceCreate,
			Entry("with default arguments", defaultTestTimeout, appConWithoutNetworkPolicies),
		)
		DescribeTable(
			"The application-connector is updated properly when network policies are disabled",
			// the table function that will be executed for each entry
			testInstanceDisableNetworkPolicy,
			Entry("with default arguments", defaultTestTimeout, appConWithNetworkPolicies),
		)
	})
})

func validateAppConState(ctx context.Context, expected State, key types.NamespacedName) error {
	state, err := getApplicationConnectorState(ctx, key)
	if err != nil {
		return err
	}
	if state != expected {
		return fmt.Errorf("invalid state")
	}
	return nil
}

func testInstanceCreate(t time.Duration, ac v1alpha1.ApplicationConnector) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTestTimeout)
	defer cancel()

	By(fmt.Sprintf("create application-connector instance: %s/%s", ac.Namespace, ac.Name))
	Expect(k8sClient.Create(ctx, &ac)).To(Succeed())

	testReconcile(ac, ctx, t)
}

func testInstanceDisableNetworkPolicy(t time.Duration, ac v1alpha1.ApplicationConnector) {
	testInstanceUpdate(t, ac, func(ac *v1alpha1.ApplicationConnector) {
		ac.Spec.NetworkPoliciesEnabled = false
	})
}

func testInstanceUpdate(t time.Duration, ac v1alpha1.ApplicationConnector, updateFunc func(*v1alpha1.ApplicationConnector)) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTestTimeout)
	defer cancel()

	By(fmt.Sprintf("create application-connector instance: %s/%s", ac.Namespace, ac.Name))
	Expect(k8sClient.Create(ctx, &ac)).To(Succeed())

	testReconcile(ac, ctx, t)

	var updatedAC v1alpha1.ApplicationConnector
	Expect(k8sClient.Get(ctx, types.NamespacedName{Name: ac.Name, Namespace: ac.Namespace}, &updatedAC)).To(Succeed())

	updateFunc(&updatedAC)

	Expect(k8sClient.Update(ctx, &updatedAC)).To(Succeed())

	testReconcile(ac, ctx, t)
}

func testReconcile(ac v1alpha1.ApplicationConnector, ctx context.Context, t time.Duration) {
	instanceNsName := types.NamespacedName{Name: ac.Name, Namespace: ac.Namespace}
	// both deployments should not be ready, the CR status should be in
	// processing state
	Eventually(validateAppConState).
		WithArguments(ctx, StateProcessing, instanceNsName).
		WithPolling(time.Second).
		WithTimeout(t).
		Should(Succeed())

	By("simulate k8s reaction - update application-gateway deployment and create replica-set")
	appGatewayNsName := types.NamespacedName{Name: appGatewayDeploymentName, Namespace: ac.Namespace}
	Expect(simulateK8sDeploymentRdy(ctx, appGatewayNsName)).To(Succeed())

	// application-connectivity-validator deployments should not be ready, the CR status should be in
	// processing state
	Eventually(validateAppConState).
		WithArguments(ctx, StateProcessing, instanceNsName).
		WithPolling(time.Second).
		WithTimeout(t).
		Should(Succeed())

	By("simulate k8s reaction - update application-connectivity-validator deployment and create replica-set")
	appConValidatorDeploymentName := types.NamespacedName{Name: appConValidatorDeploymentName, Namespace: ac.Namespace}
	Expect(simulateK8sDeploymentRdy(ctx, appConValidatorDeploymentName)).To(Succeed())

	// application-connectivity-validator deployments should not be ready, the CR status should be in
	// processing state
	Eventually(validateAppConState).
		WithArguments(ctx, StateProcessing, instanceNsName).
		WithPolling(time.Second).
		WithTimeout(t).
		Should(Succeed())

	By(fmt.Sprintf("simulate k8s reaction - update %s deployment and create replica-set", compassRtAgentDeploymentName))
	compassRtAgentNsName := types.NamespacedName{Name: compassRtAgentDeploymentName, Namespace: ac.Namespace}
	Expect(simulateK8sDeploymentRdy(ctx, compassRtAgentNsName)).To(Succeed())

	// all deployments should be ready, the CR status should be in
	// ready state
	Eventually(validateAppConState).
		WithArguments(ctx, StateReady, instanceNsName).
		WithPolling(time.Second).
		WithTimeout(t).
		Should(Succeed())

	// check if domain name was set
	Expect(validateGateway(ctx, testDomainName)).To(BeNil())
	Expect(k8sClient.Delete(ctx, &ac)).To(Succeed())
}

func validateGateway(ctx context.Context, expectedDomainName string) error {
	var u unstructured.Unstructured
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Kind:    "Gateway",
		Version: "v1beta1",
		Group:   "networking.istio.io",
	})

	gatewayKey := types.NamespacedName{
		Name:      "kyma-gateway-application-connector",
		Namespace: "kyma-system",
	}

	if err := k8sClient.Get(ctx, gatewayKey, &u); err != nil {
		return err
	}

	var gateway istio.Gateway
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &gateway); err != nil {
		return fmt.Errorf("conversion error: %w", err)
	}

	expectedHost := fmt.Sprintf("gateway.%s", expectedDomainName)

	for _, s := range gateway.Spec.Servers {
		for _, h := range s.Hosts {
			if expectedHost == h {
				return nil
			}
		}
	}

	return fmt.Errorf(`domain: "%s" name not propagated`, expectedDomainName)
}

func namespace(name string) corev1.Namespace {
	return corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

func secret(ns string) corev1.Secret {
	return corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "compass-agent-configuration",
			Namespace: ns,
		},
	}
}

func applicationConnector(name, nsName string, spec v1alpha1.ApplicationConnectorSpec) v1alpha1.ApplicationConnector {
	return v1alpha1.ApplicationConnector{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: nsName,
			Labels: map[string]string{
				"operator.kyma-project.io/kyma-name": "test",
			},
		},
		Spec: spec,
	}
}

func replicaSet(d appsv1.Deployment) appsv1.ReplicaSet {
	return appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-replica-set", d.Name),
			Namespace: d.Namespace,
			Labels: map[string]string{
				"app": d.Name,
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "apps/v1",
					Kind:       "Deployment",
					Name:       d.Name,
					UID:        d.GetUID(),
					Controller: ptr.To[bool](true),
				},
			},
		},
		// dummy values
		Spec: appsv1.ReplicaSetSpec{
			Replicas: ptr.To[int32](1),
			Selector: d.Spec.Selector,
			Template: d.Spec.Template,
		},
	}
}

func simulateK8sDeploymentRdy(ctx context.Context, key types.NamespacedName) error {
	var deployment appsv1.Deployment
	if err := k8sClient.Get(ctx, key, &deployment); err != nil {
		return err
	}

	deployment.Status.Conditions = append(deployment.Status.Conditions, appsv1.DeploymentCondition{
		Type:    appsv1.DeploymentAvailable,
		Status:  corev1.ConditionTrue,
		Reason:  "test-reason",
		Message: "test-message",
	})
	deployment.Status.Replicas = 1

	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		return k8sClient.Status().Update(ctx, &deployment)
	})
	if err != nil {
		return err
	}

	rs := replicaSet(deployment)
	if err := k8sClient.Create(ctx, &rs); err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	rs.Status.ReadyReplicas = 1
	rs.Status.Replicas = 1

	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		return k8sClient.Status().Update(ctx, &rs)
	})
}

func getApplicationConnectorState(ctx context.Context, key types.NamespacedName) (State, error) {
	var emptyState = State("")
	var connector v1alpha1.ApplicationConnector
	err := k8sClient.Get(ctx, key, &connector)
	if err != nil {
		return emptyState, err
	}
	return State(connector.Status.State), nil
}
