package predicate

import (
	"reflect"

	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type deploymentPredicate struct {
	predicate.ResourceVersionChangedPredicate
	log *zap.SugaredLogger
}

var NewDeploymentPredicate2 = predicate.TypedResourceVersionChangedPredicate[*unstructured.Unstructured]{
	TypedFuncs: predicate.TypedFuncs[*unstructured.Unstructured]{
		CreateFunc:  nil,
		DeleteFunc:  nil,
		UpdateFunc:  nil,
		GenericFunc: nil,
	},
}

func NewDeploymentPredicate(log *zap.SugaredLogger) predicate.Predicate {
	return &deploymentPredicate{
		log: log,
	}
}

func (p deploymentPredicate) Update(e event.UpdateEvent) bool {
	var oldDeployment appsv1.Deployment
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(e.ObjectOld.(*unstructured.Unstructured).Object, &oldDeployment); err != nil {
		p.log.Warnf("unable to convert old deployment: %w", err)
		return true
	}

	var newDeployment appsv1.Deployment
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(e.ObjectNew.(*unstructured.Unstructured).Object, &newDeployment); err != nil {
		p.log.Warnf("unable to convert new deployment: %w", err)
		return true
	}

	// check if status changed
	if statusEqual := reflect.DeepEqual(oldDeployment.Status, newDeployment.Status); !statusEqual {
		return true
	}

	// check if spec changed
	if specEqual := reflect.DeepEqual(oldDeployment.Spec, newDeployment.Spec); !specEqual {
		return true
	}

	// check if labels changed
	if labelsEqual := reflect.DeepEqual(oldDeployment.GetLabels(), newDeployment.GetLabels()); !labelsEqual {
		return true
	}

	// check if annotations changed
	if annotationsEqual := reflect.DeepEqual(oldDeployment.GetAnnotations(), newDeployment.GetAnnotations()); !annotationsEqual {
		return true
	}

	// check if namespace changed
	if oldDeployment.GetNamespace() != newDeployment.GetNamespace() {
		return true
	}

	return false
}
