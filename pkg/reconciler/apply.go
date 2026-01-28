package reconciler

import (
	"context"
	"errors"

	"github.com/kyma-project/application-connector-manager/api/v1alpha1"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	ErrInstallationFailed = errors.New("installation failed")
)

func sFnApply(ctx context.Context, r *fsm, s *systemState) (stateFn, *ctrl.Result, error) {
	var isError bool
	for _, obj := range append(r.Objs, r.Deps...) {
		r.log.
			With("gvk", obj.GetObjectKind().GroupVersionKind()).
			With("name", obj.GetName()).
			With("ns", obj.GetNamespace()).
			Debug("applying")

		err := r.Patch(ctx, &obj, client.Apply, &client.PatchOptions{ //nolint:staticcheck
			Force:        ptr.To[bool](true),
			FieldManager: "application-connector-manager",
		})

		if err != nil {
			r.log.With("err", err).Error("apply error")
			isError = true
		}

		s.objs = append(s.objs, obj)
	}

	if !isError {
		return switchState(sFnVerify)
	}

	s.instance.UpdateStateFromErr(
		v1alpha1.ConditionTypeInstalled,
		v1alpha1.ConditionReasonApplyObjError,
		ErrInstallationFailed,
	)
	r.log.Error("Error during applying helm charts!")
	return stopWithErrorAndRequeue(ErrInstallationFailed) // exponential backoff
}
