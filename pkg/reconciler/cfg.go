package reconciler

import (
	"github.com/kyma-project/application-connector-manager/pkg/unstructured"
)

// module specific configuuration
type Cfg struct {
	// the Finalizer identifies the module and is is used to delete
	// the module resources
	Finalizer string
	// the objects are module component parts; objects are applied
	// on the cluster one by one with given order
	Objs         []unstructured.Unstructured `json:"objs"`
	Deps         []unstructured.Unstructured
	OptionalObjs []unstructured.Unstructured
}
