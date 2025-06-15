/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"k8s.io/apimachinery/pkg/labels"
	"reflect"

	"github.com/kyma-project/application-connector-manager/api/v1alpha1"
	acm_predicate "github.com/kyma-project/application-connector-manager/pkg/common/controller-runtime/predicate"
	"github.com/kyma-project/application-connector-manager/pkg/reconciler"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	v2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type State string

// Valid CustomObject States.
const (
	// StateReady signifies application-connector is ready and has been installed successfully.
	StateReady State = "Ready"

	// StateProcessing signifies application-connector is reconciling and is in the process of installation.
	// Processing can also signal that the Installation previously encountered an error and is now recovering.
	StateProcessing State = "Processing"

	// StateError signifies an error for application-connector. This signifies that the Installation
	// process encountered an error.
	// Contrary to Processing, it can be expected that this state should change on the next retry.
	StateError State = "Error"

	// StateDeleting signifies application-connector is being deleted. This is the state that is used
	// when a deletionTimestamp was detected and Finalizers are picked up.
	StateDeleting State = "Deleting"
)

type ApplicationConnetorReconciler interface {
	reconcile.Reconciler
	SetupWithManager(mgr ctrl.Manager) error
}

type Watch = func(src source.TypedSource[reconcile.Request]) error

type applicationConnectorReconciler struct {
	log *zap.SugaredLogger
	reconciler.Cfg
	reconciler.K8s
	Watch
	DepsACK bool
}

func NewApplicationConnetorReconciler(c client.Client, r record.EventRecorder, log *zap.SugaredLogger, o1 []unstructured.Unstructured, o2 []unstructured.Unstructured) ApplicationConnetorReconciler {
	return &applicationConnectorReconciler{
		log: log,
		Cfg: reconciler.Cfg{
			Finalizer: v1alpha1.Finalizer,
			Objs:      o1,
			Deps:      o2,
		},
		K8s: reconciler.K8s{
			Client:        c,
			EventRecorder: r,
		},
	}
}

func (r *applicationConnectorReconciler) mapFunction(_ context.Context, object client.Object) []reconcile.Request {
	var applicationConnectors v1alpha1.ApplicationConnectorList
	err := r.List(context.Background(), &applicationConnectors)

	if apierrors.IsNotFound(err) {
		return nil
	}

	if err != nil {
		r.log.Error(err)
		return nil
	}

	if len(applicationConnectors.Items) < 1 {
		return nil
	}

	// instance is being deleted, do not notify it about changes
	instanceIsBeingDeleted := !applicationConnectors.Items[0].GetDeletionTimestamp().IsZero()
	if instanceIsBeingDeleted {
		return nil
	}

	// make sure only 1 controller will handle change
	return []ctrl.Request{
		{
			NamespacedName: types.NamespacedName{
				Namespace: applicationConnectors.Items[0].Namespace,
				Name:      applicationConnectors.Items[0].Name,
			},
		},
	}
}

var ommitStatusChanged = predicate.Or(
	predicate.LabelChangedPredicate{},
	predicate.AnnotationChangedPredicate{},
	predicate.GenerationChangedPredicate{},
)

var hpaResourceVersionChangedPredicate2 = predicate.TypedResourceVersionChangedPredicate[*unstructured.Unstructured]{
	TypedFuncs: predicate.TypedFuncs[*unstructured.Unstructured]{
		CreateFunc:  nil,
		DeleteFunc:  nil,
		UpdateFunc:  nil,
		GenericFunc: nil,
	},
}

type hpaResourceVersionChangedPredicate struct {
	predicate.TypedResourceVersionChangedPredicate[*unstructured.Unstructured]
	log *zap.SugaredLogger
}

func (h hpaResourceVersionChangedPredicate) Update(e event.TypedUpdateEvent[*unstructured.Unstructured]) bool {
	if update := h.TypedResourceVersionChangedPredicate.Update(e); !update {
		return false
	}

	var newObj v2.HorizontalPodAutoscaler
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(e.ObjectNew.Object, &newObj); err != nil {
		return true
	}

	var oldObj v2.HorizontalPodAutoscaler
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(e.ObjectOld.Object, &oldObj); err != nil {
		return true
	}

	conditionsEqual := reflect.DeepEqual(oldObj.Status.Conditions, newObj.Status.Conditions)
	replicasEqual := oldObj.Status.CurrentReplicas == newObj.Status.CurrentReplicas

	result := !conditionsEqual || !replicasEqual
	if result {
		h.log.With("conditionsEqual", conditionsEqual, "replicasEqual", replicasEqual).Debugf("reconciliation triggered by HPA: %s/%s", oldObj.Namespace, oldObj.Name)
	}
	return result
}

var (
	hpaGroupVersionKind = schema.GroupVersionKind{
		Group:   v2.GroupName,
		Version: "v2",
		Kind:    "HorizontalPodAutoscaler",
	}

	deploymentVersionKind = schema.GroupVersionKind{
		Group:   appsv1.GroupName,
		Version: appsv1.SchemeGroupVersion.Version,
		Kind:    "Deployment",
	}
)

func LabelSelectorPredicate(s metav1.LabelSelector) (predicate.TypedPredicate[*unstructured.Unstructured], error) {
	selector, err := metav1.LabelSelectorAsSelector(&s)
	if err != nil {
		return predicate.TypedFuncs[*unstructured.Unstructured]{}, err
	}

	//
	return predicate.NewTypedPredicateFuncs[*unstructured.Unstructured](func(u *unstructured.Unstructured) bool {
		return selector.Matches(labels.Set(u.GetLabels()))
	}), nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *applicationConnectorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	//labelSelectorPredicate, err := predicate.LabelSelectorPredicate[*unstructured.Unstructured](
	//	metav1.LabelSelector{
	//		MatchLabels: map[string]string{
	//			"app.kubernetes.io/part-of": "application-connector-manager",
	//		},
	//	},
	//)

	//MatchLabels: map[string]string{
	//	"app.kubernetes.io/part-of": "application-connector-manager",
	//}

	var labelSelectorPredicate predicate.TypedLabelChangedPredicate[*unstructured.Unstructured]

	b := ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.ApplicationConnector{}, builder.WithPredicates(ommitStatusChanged))

	// create functtion to register wached objects
	watchFn := func(u unstructured.Unstructured) {
		objPredicate := predicate.TypedResourceVersionChangedPredicate[*unstructured.Unstructured]{}

		if u.GroupVersionKind() == hpaGroupVersionKind {
			objPredicate = hpaResourceVersionChangedPredicate2
		}

		if u.GroupVersionKind() == deploymentVersionKind {
			objPredicate = acm_predicate.NewDeploymentPredicate2
		}

		r.log.With("gvk", u.GroupVersionKind().String()).Infoln("adding watcher")
		allLabelSelectors := predicate.And(labelSelectorPredicate, objPredicate)

		kind := source.Kind(mgr.GetCache(), &unstructured.Unstructured{}, &handler.TypedEnqueueRequestForObject[*unstructured.Unstructured]{}, allLabelSelectors)
		b = b.WatchesRawSource(kind)

		//b = b.WatchesRawSource(
		//	source.Kind(mgr.GetCache(), &u),
		//	handler.EnqueueRequestsFromMapFunc(r.mapFunction),
		//	builder.WithPredicates(
		//		predicate.And(
		//			labelSelectorPredicate,
		//			objPredicate,
		//		),
		//	),
		//)

	}
	// register watch for each managed type of object
	if err := registerWatchDistinct(r.Objs, watchFn); err != nil {
		return err
	}
	// define predicate for compass-runtime-agent secret
	craSecretPredicate := predicateCompassRtAgentSecret{
		objectName: "compass-agent-configuration",
		namespace:  "kyma-system",
		log:        r.log,
	}
	// register watch for compass-runtime-agent secret
	b = b.Watches(
		&corev1.Secret{},
		handler.EnqueueRequestsFromMapFunc(r.mapFunction),
		builder.WithPredicates(&craSecretPredicate),
	)

	controller, err := b.Build(r)
	if err != nil {
		return err
	}

	r.Watch = controller.Watch
	r.Cache = mgr.GetCache()
	return nil
}

var requCounter = 0

func (r *applicationConnectorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var instance v1alpha1.ApplicationConnector
	if err := r.Get(ctx, req.NamespacedName, &instance); err != nil {
		return ctrl.Result{
			Requeue: true,
		}, client.IgnoreNotFound(err)
	}

	stateFSM := reconciler.NewFsm(
		r.log.With("reqID", requCounter),
		r.Cfg,
		reconciler.K8s{
			Client:        r.Client,
			EventRecorder: r.EventRecorder,
			Watch:         r.Watch,
			MapFunc:       r.mapFunction,
			Cache:         r.Cache,
		},
		&r.DepsACK)
	requCounter++
	return stateFSM.Run(ctx, instance)
}
