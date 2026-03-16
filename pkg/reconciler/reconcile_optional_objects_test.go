package reconciler

import (
	"context"
	"testing"

	"github.com/kyma-project/application-connector-manager/api/v1alpha1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func networkPolicyToUnstructured(np *networkingv1.NetworkPolicy) unstructured.Unstructured {
	obj := unstructured.Unstructured{}
	obj.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "networking.k8s.io",
		Version: "v1",
		Kind:    "NetworkPolicy",
	})
	obj.SetName(np.Name)
	obj.SetNamespace(np.Namespace)
	obj.SetLabels(np.Labels)

	return obj
}

func buildFakeClient(scheme *runtime.Scheme, objs ...client.Object) client.Client {
	return fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(objs...).
		Build()
}

func buildScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	_ = networkingv1.AddToScheme(s)
	_ = v1alpha1.AddToScheme(s)
	return s
}

func managedNetworkPolicy(name, namespace string) *networkingv1.NetworkPolicy {
	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				managedByLabelKey: managedByLabelValue,
				moduleLabelKey:    moduleLabelValue,
			},
		},
	}
}

func buildFsm(fakeClient client.Client, optionalObjs []unstructured.Unstructured) *fsm {
	return &fsm{
		K8s: K8s{
			Client: fakeClient,
		},
		Cfg: Cfg{
			OptionalObjs: optionalObjs,
		},
	}
}

func TestsFnReconcileOptionalObjects_NetworkPoliciesEnabled_Applies(t *testing.T) {
	scheme := buildScheme()
	np := managedNetworkPolicy("test-np", "default")

	fakeClient := buildFakeClient(scheme)
	optionalObj := networkPolicyToUnstructured(np)

	r := buildFsm(fakeClient, []unstructured.Unstructured{optionalObj})

	s := &systemState{
		instance: v1alpha1.ApplicationConnector{
			Spec: v1alpha1.ApplicationConnectorSpec{
				NetworkPoliciesEnabled: true,
			},
		},
	}

	_, _, err := sFnReconcileOptionalObjects(context.Background(), r, s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result networkingv1.NetworkPolicyList
	if listErr := fakeClient.List(context.Background(), &result, client.InNamespace("default")); listErr != nil {
		t.Fatalf("failed to list NetworkPolicies: %v", listErr)
	}

	if len(result.Items) != 1 {
		t.Errorf("expected 1 NetworkPolicy, got %d", len(result.Items))
	}
}

func TestsFnReconcileOptionalObjects_NetworkPoliciesDisabled_Removes(t *testing.T) {
	scheme := buildScheme()

	np1 := managedNetworkPolicy("np-1", "default")
	np2 := managedNetworkPolicy("np-2", "kyma-system")
	unmanagedNp := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "unmanaged-np",
			Namespace: "default",
			Labels: map[string]string{
				"some-other-label": "value",
			},
		},
	}

	fakeClient := buildFakeClient(scheme, np1, np2, unmanagedNp)

	r := buildFsm(fakeClient, []unstructured.Unstructured{})

	s := &systemState{
		instance: v1alpha1.ApplicationConnector{
			Spec: v1alpha1.ApplicationConnectorSpec{
				NetworkPoliciesEnabled: false,
			},
		},
	}

	_, _, err := sFnReconcileOptionalObjects(context.Background(), r, s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result networkingv1.NetworkPolicyList
	if listErr := fakeClient.List(context.Background(), &result); listErr != nil {
		t.Fatalf("failed to list NetworkPolicies: %v", listErr)
	}

	for _, item := range result.Items {
		if item.Labels[managedByLabelKey] == managedByLabelValue {
			t.Errorf("expected managed NetworkPolicy %q to be deleted, but it still exists", item.Name)
		}
	}

	var unmanagedResult networkingv1.NetworkPolicy
	if getErr := fakeClient.Get(context.Background(), client.ObjectKey{Name: "unmanaged-np", Namespace: "default"}, &unmanagedResult); getErr != nil {
		t.Errorf("expected unmanaged NetworkPolicy to still exist, but got error: %v", getErr)
	}
}
