package reconciler

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

var gardenerCM types.NamespacedName = types.NamespacedName{
	Name:      "shoot-info",
	Namespace: "kube-system",
}

func sFnDetectDomain(ctx context.Context, r *fsm, s *systemState) (stateFn, *ctrl.Result, error) {
	if s.instance.Spec.DomainName != "" {
		return switchState(sFnUpdate)
	}

	// check if domain name was already fetched
	if s.domainName != "" {
		return switchState(sFnUpdate)
	}

	// try to fetch domain name from config map
	var cm v1.ConfigMap
	if err := r.Get(ctx, gardenerCM, &cm); err != nil {
		return stopWithErrorAndRequeue(fmt.Errorf("unable to detect domain: %w", err))
	}

	domainName, found := cm.Data["domain"]
	if !found {
		return stopWithErrorAndRequeue(fmt.Errorf("domain not found"))
	}

	s.domainName = domainName
	return switchState(sFnUpdate)
}
