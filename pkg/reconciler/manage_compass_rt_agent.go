package reconciler

import (
	"context"
	"fmt"

	"github.com/kyma-project/application-connector-manager/api/v1alpha1"
	"github.com/kyma-project/application-connector-manager/pkg/unstructured"
	"golang.org/x/exp/slices"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	keyCompassAgentCfg = client.ObjectKey{
		Namespace: "kyma-system",
		Name:      "compass-agent-configuration",
	}
)

type craDTO struct {
	Domain   string
	Replicas int32
}

func buildUpdateCompassRuntimeAgent(ctx context.Context, r *fsm, _ *systemState) (func(i v1alpha1.ApplicationConnectorSpec, objs uList, _ uList) error, error) {
	var secret v1.Secret
	err := r.Get(ctx, keyCompassAgentCfg, &secret)

	var replicas int32 = 1
	if errors.IsNotFound(err) {
		replicas = 0
	}
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	return func(i v1alpha1.ApplicationConnectorSpec, objs uList, _ uList) error {
		u, err := unstructured.IsDeployment("compass-runtime-agent").First(objs)
		if err != nil {
			return err
		}

		if err := unstructured.Update(u, craDTO{Domain: i.DomainName, Replicas: replicas}, updateCompassRuntimeAgent); err != nil {
			return err
		}
		return nil
	}, nil
}

func updateCompassRuntimeAgent(d *appv1.Deployment, dto craDTO) error {
	if d == nil {
		return fmt.Errorf("invalid value: nil")
	}
	// find compass-runtime-agent container
	index := slices.IndexFunc(
		d.Spec.Template.Spec.Containers,
		func(c corev1.Container) bool { return c.Name == "compass-runtime-agent" })
	// return error if compass-runtime-agent container was not found
	if index == -1 {
		return fmt.Errorf("compass-runtime-agent container: %w", unstructured.ErrNotFound)
	}

	craEnvs := &d.Spec.Template.Spec.Containers[index].Env

	// define all update functions
	fns := []update{
		envVarUpdate(
			craEnvs,
			corev1.EnvVar{
				Name:  v1alpha1.EnvRuntimeAgentAppRuntimeEventsURL,
				Value: fmt.Sprintf("https://gateway.%s", dto.Domain),
			}),
		envVarUpdate(
			craEnvs,
			corev1.EnvVar{
				Name:  v1alpha1.EnvRuntimeAgnetAppRuntimeConsoleURL,
				Value: fmt.Sprintf("https://console.%s", dto.Domain),
			}),
	}
	// perform update
	for _, f := range fns {
		if err := f(); err != nil {
			return err
		}
	}
	d.Spec.Replicas = ptr.To(dto.Replicas)

	return nil
}
