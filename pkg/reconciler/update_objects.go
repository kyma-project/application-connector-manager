package reconciler

import (
	"context"
	"fmt"
	"strings"

	"github.com/kyma-project/application-connector-manager/api/v1alpha1"
	"github.com/kyma-project/application-connector-manager/pkg/unstructured"
	"golang.org/x/exp/slices"
	istio "istio.io/client-go/pkg/apis/networking/v1beta1"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

type update = func() error

type uList = []unstructured.Unstructured

type defaultingOption func(*v1alpha1.ApplicationConnectorSpec) error

func applyDefaults(spec v1alpha1.ApplicationConnectorSpec, ops ...defaultingOption) (v1alpha1.ApplicationConnectorSpec, error) {
	var specCopy v1alpha1.ApplicationConnectorSpec
	spec.DeepCopyInto(&specCopy)

	for _, opt := range ops {
		if err := opt(&specCopy); err != nil {
			return v1alpha1.ApplicationConnectorSpec{}, err
		}
	}
	return specCopy, nil
}

func sFnUpdate(ctx context.Context, r *fsm, s *systemState) (stateFn, *ctrl.Result, error) {
	updatedSpec, err := applyDefaults(s.instance.Spec, func(spec *v1alpha1.ApplicationConnectorSpec) error {
		if s.domainName != "" {
			spec.DomainName = s.domainName
		}
		return nil
	})

	if err != nil {
		return stopWithErrorAndRequeue(fmt.Errorf("defaults application failed: %w", err))
	}

	updateCRA, err := buildUpdateCompassRuntimeAgent(ctx, r, s)
	if err != nil {
		return stopWithErrorAndRequeue(fmt.Errorf("unable to build CRA update function: %w", err))
	}

	for _, f := range []func(v1alpha1.ApplicationConnectorSpec, uList, uList) error{
		updateCRA,
		updateCentralApplicationGateway,
		updateAppConnectivityValidator,
		updateGateways,
		updateVirtualServices,
	} {
		if err := f(updatedSpec, r.Objs, r.Deps); err != nil {
			return stopWithErrorAndRequeue(err)
		}
	}
	return switchState(sFnApply)
}

func envVarUpdate(envs *[]corev1.EnvVar, newEnv corev1.EnvVar) update {
	return func() error {
		if envs == nil {
			return fmt.Errorf("invalid value: nil")
		}
		envIndex := slices.IndexFunc(
			*envs,
			func(env corev1.EnvVar) bool { return newEnv.Name == env.Name })
		// return error if env variable was not found
		if envIndex == -1 {
			*envs = append(*envs, newEnv)
			return nil
		}

		(*envs)[envIndex] = newEnv
		return nil
	}
}

func updateAppConnectivityValidator(i v1alpha1.ApplicationConnectorSpec, objs uList, _ uList) error {

	u, err := unstructured.IsDeployment("central-application-connectivity-validator").First(objs)
	if err != nil {
		return err
	}

	if err := unstructured.Update(u, i.AppConValidatorSpec, updateAppConnValidatorEnvs); err != nil {
		return err
	}
	return nil
}

func updateAppConnValidatorEnvs(d *appv1.Deployment, v v1alpha1.AppConnValidatorSpec) error {

	index := slices.IndexFunc(
		d.Spec.Template.Spec.Containers,
		func(c corev1.Container) bool { return c.Name == "central-application-connectivity-validator" })

	if index == -1 {
		return fmt.Errorf("central-application-connectivity-validator: %w", unstructured.ErrNotFound)
	}

	validatorEnvs := &d.Spec.Template.Spec.Containers[index].Env

	fns := []update{
		envVarUpdate(
			validatorEnvs,
			corev1.EnvVar{
				Name:  v1alpha1.EnvAppConnValidatorLogLevel,
				Value: string(v.LogLevel),
			}),
		envVarUpdate(
			validatorEnvs,
			corev1.EnvVar{
				Name:  v1alpha1.EnvAppConnValidatorLogFormat,
				Value: string(v.LogFormat),
			}),
	}

	// perform update
	for _, f := range fns {
		if err := f(); err != nil {
			return err
		}
	}
	return nil
}

func updateCentralApplicationGateway(i v1alpha1.ApplicationConnectorSpec, objs uList, _ uList) error {
	u, err := unstructured.IsDeployment("central-application-gateway").First(objs)
	if err != nil {
		return err
	}

	if err := unstructured.Update(u, i.ApplicationGatewaySpec, updateCAG); err != nil {
		return err
	}
	return nil
}

func updateCAG(d *appv1.Deployment, v v1alpha1.AppGatewaySpec) error {
	if d == nil {
		return fmt.Errorf("invalid value: nil")
	}
	// find compass-runtime-agent container
	index := slices.IndexFunc(
		d.Spec.Template.Spec.Containers,
		func(c corev1.Container) bool { return c.Name == "central-application-gateway" })
	// return error if compass-runtime-agent container was not found
	if index == -1 {
		return fmt.Errorf("central-application-gateway container: %w", unstructured.ErrNotFound)
	}
	cAppG8wayArgs := &d.Spec.Template.Spec.Containers[index].Args

	// define all update functions
	fns := []update{
		argValueUpdate(cAppG8wayArgs, v1alpha1.ArgCentralAppGatewayRequestTimeout, fmt.Sprintf("%.0f", v.RequestTimeout.Seconds())),
		argValueUpdate(cAppG8wayArgs, v1alpha1.ArgCentralAppGatewayProxyTimeout, fmt.Sprintf("%.0f", v.ProxyTimeout.Seconds())),
		argValueUpdate(cAppG8wayArgs, v1alpha1.ArgLogLevel, fmt.Sprintf("%v", v.LogLevel)),
	}
	// perform update
	for _, f := range fns {
		if err := f(); err != nil {
			return err
		}
	}
	return nil
}

func argValueUpdate(args *[]string, key string, newValue any) update {
	return func() error {
		newArg := fmt.Sprintf("%s=%v", key, newValue)
		// assume argument is distinct
		argIndex := slices.IndexFunc(*args, func(s string) bool {
			return strings.HasPrefix(s, key)
		})
		// return error if env variable was not found
		if argIndex == -1 {
			*args = append(*args, newArg)
			return nil
		}

		(*args)[argIndex] = newArg
		return nil
	}
}

func updateG8(g *istio.Gateway, domainName string) error {
	if g == nil {
		return fmt.Errorf("invalid value: nil")
	}

	for i := range g.Spec.Servers {
		g.Spec.Servers[i].Hosts = []string{fmt.Sprintf("gateway.%s", domainName)}
	}

	return nil
}

func updateGateways(i v1alpha1.ApplicationConnectorSpec, _ uList, deps uList) error {
	us, err := unstructured.IsGatewayKind().All(deps)
	if err != nil {
		return err
	}

	for _, u := range us {
		if err := unstructured.Update(u, i.DomainName, updateG8); err != nil {
			return err
		}
	}

	return nil
}

func updateVS(g *istio.VirtualService, domainName string) error {
	if g == nil {
		return fmt.Errorf("invalid value: nil")
	}

	g.Spec.Hosts = []string{fmt.Sprintf("gateway.%s", domainName)}

	return nil
}

func updateVirtualServices(i v1alpha1.ApplicationConnectorSpec, _ uList, deps uList) error {
	us, err := unstructured.IsVirtualService().All(deps)
	if err != nil {
		return err
	}

	for _, u := range us {
		if err := unstructured.Update(u, i.DomainName, updateVS); err != nil {
			return err
		}
	}

	return nil
}
