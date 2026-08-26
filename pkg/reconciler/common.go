package reconciler

import (
	ctrl "sigs.k8s.io/controller-runtime"
	"time"
)

func stopWithErrorAndRequeue(err error) (stateFn, *ctrl.Result, error) {
	return sFnUpdateStatus(nil, err), nil, nil
}

func stopWithRequeueAfter(duration time.Duration) (stateFn, *ctrl.Result, error) {
	return sFnUpdateStatus(&ctrl.Result{RequeueAfter: duration}, nil), nil, nil
}

func stopWithNoRequeue() (stateFn, *ctrl.Result, error) {
	return sFnUpdateStatus(nil, nil), nil, nil
}

func stopWithRequeue() (stateFn, *ctrl.Result, error) {
	return sFnUpdateStatus(&ctrl.Result{RequeueAfter: time.Second}, nil), nil, nil
}

func switchState(fn stateFn) (stateFn, *ctrl.Result, error) {
	return fn, nil, nil
}
