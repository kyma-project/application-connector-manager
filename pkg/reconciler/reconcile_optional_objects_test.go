package reconciler

import (
	"context"
	"github.com/kyma-project/application-connector-manager/pkg/yaml"
	"github.com/stretchr/testify/require"
	"os"
	"testing"

	"github.com/kyma-project/application-connector-manager/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestFnReconcileOptionalObjects_NetworkPoliciesEnabled_Applies(t *testing.T) {
	patchObject = func(ctx context.Context, c client.Client, obj unstructured.Unstructured) error {
		return c.Create(ctx, &obj)
	}

	scheme := buildScheme()

	file, err := os.Open("./testdata/application-connector-optional.yaml")
	require.NoError(t, err)

	data, err := yaml.LoadData(file)
	require.NoError(t, err)

	err = file.Close()
	require.NoError(t, err)

	fakeClient := buildFakeClient(scheme)

	r := buildFsm(fakeClient, data)

	s := &systemState{
		instance: v1alpha1.ApplicationConnector{
			Spec: v1alpha1.ApplicationConnectorSpec{
				NetworkPoliciesEnabled: true,
			},
		},
	}

	_, _, err = sFnReconcileOptionalObjects(context.Background(), r, s)
	assert.NoError(t, err)

	var result networkingv1.NetworkPolicyList
	err = fakeClient.List(context.Background(), &result, client.InNamespace("kyma-system"))
	assert.NoError(t, err)
	assert.Len(t, result.Items, 1)
}

func TestFnReconcileOptionalObjects_NetworkPoliciesDisabled_Removes(t *testing.T) {
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
	assert.NoError(t, err)

	var result networkingv1.NetworkPolicyList
	err = fakeClient.List(context.Background(), &result)
	assert.NoError(t, err)

	for _, item := range result.Items {
		assert.NotEqual(t, managedByLabelValue, item.Labels[managedByLabelKey],
			"expected managed NetworkPolicy %q to be deleted, but it still exists", item.Name)
	}

	var unmanagedResult networkingv1.NetworkPolicy
	err = fakeClient.Get(context.Background(), client.ObjectKey{Name: "unmanaged-np", Namespace: "default"}, &unmanagedResult)
	assert.NoError(t, err, "expected unmanaged NetworkPolicy to still exist")
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
