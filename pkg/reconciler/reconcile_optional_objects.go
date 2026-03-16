package reconciler

import (
	"context"
	"github.com/kyma-project/application-connector-manager/api/v1alpha1"
	"github.com/kyma-project/application-connector-manager/pkg/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"

	networkingv1 "k8s.io/api/networking/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	managedByLabelKey   = "app.kubernetes.io/managed-by"
	managedByLabelValue = "application-connector-manager"
	moduleLabelKey      = "kyma-project.io/module"
	moduleLabelValue    = "application-connector"
)

func sFnReconcileOptionalObjects(ctx context.Context, r *fsm, s *systemState) (stateFn, *ctrl.Result, error) {
	if s.instance.Spec.NetworkPoliciesEnabled {
		for _, obj := range r.OptionalObjs {
			if err := patchObject(ctx, r, obj); err != nil {
				s.instance.UpdateStateFromErr(
					v1alpha1.ConditionTypeInstalled,
					v1alpha1.ConditionReasonOptionalManifestsReconciliationErr,
					ErrInstallationFailed,
				)
				return stopWithErrorAndRequeue(ErrInstallationFailed)
			}
		}
	} else {
		if err := removeNetworkPolicies(ctx, r.Client); err != nil {
			s.instance.UpdateStateFromErr(
				v1alpha1.ConditionTypeInstalled,
				v1alpha1.ConditionReasonOptionalManifestsReconciliationErr,
				ErrInstallationFailed,
			)
			return stopWithErrorAndRequeue(ErrInstallationFailed)
		}
	}
	return switchState(sFnVerify)
}

func patchObject(ctx context.Context, c client.Client, obj unstructured.Unstructured) error {
	bytes, err := obj.MarshalJSON()
	if err != nil {
		return err
	}

	return c.Patch(ctx, &obj, client.RawPatch(types.ApplyPatchType, bytes), &client.PatchOptions{
		Force:        ptr.To[bool](true),
		FieldManager: "application-connector-manager",
	})
}

func removeNetworkPolicies(ctx context.Context, c client.Client) error {
	var networkPolicyList networkingv1.NetworkPolicyList
	if err := c.List(ctx, &networkPolicyList, client.MatchingLabels{
		managedByLabelKey: managedByLabelValue,
		moduleLabelKey:    moduleLabelValue,
	}); err != nil {
		return err
	}

	for i := range networkPolicyList.Items {
		if err := c.Delete(ctx, &networkPolicyList.Items[i]); client.IgnoreNotFound(err) != nil {
			return err
		}
	}
	return nil
}
