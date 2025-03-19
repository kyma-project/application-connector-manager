package reconciler

import (
	"context"
	"errors"
	"github.com/kyma-project/application-connector-manager/pkg/unstructured"
	"time"

	"github.com/kyma-project/application-connector-manager/api/v1alpha1"
	acm_predicate "github.com/kyma-project/application-connector-manager/pkg/common/controller-runtime/predicate"
	"github.com/kyma-project/application-connector-manager/pkg/common/types"
	"golang.org/x/exp/slices"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var ErrIstioNotFound = errors.New("ISTIO not found")

func checkDeps(crds []apiextensionsv1.CustomResourceDefinition, gks ...schema.GroupVersionKind) error {
	var ackCount int
	for _, crd := range crds {
		isGK := func(gk schema.GroupVersionKind) bool {
			return gk.Group == crd.Spec.Group && gk.Kind == crd.Spec.Names.Kind
		}
		// check if one expected dependency version is served on cluster
		isVersion := func(v apiextensionsv1.CustomResourceDefinitionVersion) bool {
			for _, gkv := range gks {
				matches := v.Served && gkv.Version == v.Name
				if matches {
					return true
				}
			}
			return false
		}

		isServedVersion := slices.ContainsFunc(crd.Spec.Versions, isVersion)
		isOneOfGKS := slices.ContainsFunc(gks, isGK)

		if !isOneOfGKS || !isServedVersion {
			continue
		}
		ackCount++
	}

	dependencyCount := len(types.Dependencies)
	if ackCount != dependencyCount {
		return ErrIstioNotFound
	}
	return nil
}

func sFnCheckDependencies(ctx context.Context, r *fsm, s *systemState) (stateFn, *ctrl.Result, error) {
	var crds apiextensionsv1.CustomResourceDefinitionList
	if err := r.List(ctx, &crds); err != nil {
		s.instance.UpdateStateFromErr(v1alpha1.ConditionTypeInstalled, v1alpha1.ConditionReasonApplyObjError, err)
		return stopWithErrorAndRequeue(err)
	}

	if err := checkDeps(crds.Items, types.Dependencies...); err != nil {
		s.instance.UpdateStateFromErr(v1alpha1.ConditionTypeInstalled, v1alpha1.ConditionReasonApplyObjError, err)
		return stopWithRequeueAfter(time.Second * 10)
	}

	if *r.dependencyACK {
		return switchState(sFnInitialize)
	}

	return switchState(sFnRegisterDependencyWatch)
}

func sFnRegisterDependencyWatch(_ context.Context, r *fsm, s *systemState) (stateFn, *ctrl.Result, error) {
	labelSelectorPredicate, err := predicate.LabelSelectorPredicate(
		metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app.kubernetes.io/part-of": "application-connector-manager",
			},
		},
	)

	if err != nil {
		s.instance.UpdateStateFromErr(v1alpha1.ConditionTypeInstalled, v1alpha1.ConditionReasonApplyObjError, err)
		return stopWithErrorAndRequeue(err)
	}

	for _, u := range r.Deps {
		r.log.With("gvk", u.GroupVersionKind()).Info("adding watch")

		var objPredicate predicate.Predicate = predicate.GenerationChangedPredicate{}

		if u.GetObjectKind().GroupVersionKind() == types.VirtualService {
			objPredicate = acm_predicate.NewVirtualServicePredicate(r.log)
		}

		if u.GetObjectKind().GroupVersionKind() == types.Gateway {
			objPredicate = acm_predicate.NewGatewayPredicate(r.log)
		}

		kind := source.Kind(r.Cache, &unstructured.Unstructured{}, &handler.TypedEnqueueRequestForObject[*unstructured.Unstructured]{}, predicate.And(labelSelectorPredicate, objPredicate))
		//kind, handler.EnqueueRequestsFromMapFunc(r.MapFunc), predicate.And(labelSelectorPredicate, objPredicate)
		err := r.Watch(kind)

		if err != nil {
			s.instance.UpdateStateFromErr(v1alpha1.ConditionTypeInstalled, v1alpha1.ConditionReasonApplyObjError, err)
			return stopWithErrorAndRequeue(err)
		}
	}

	*r.dependencyACK = true
	r.log.Info("dependency ack")

	return stopWithRequeue()
}
