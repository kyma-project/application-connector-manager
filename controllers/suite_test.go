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
	"os"
	"path/filepath"
	"testing"

	"github.com/kyma-project/application-connector-manager/pkg/reconciler"
	"github.com/kyma-project/application-connector-manager/pkg/yaml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	uzap "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	operatorv1alpha1 "github.com/kyma-project/application-connector-manager/api/v1alpha1"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	config    *rest.Config
	k8sClient client.Client
	testEnv   *envtest.Environment

	ctx    context.Context
	cancel context.CancelFunc

	externalDependencyDataPath = "../hack/common/k3d-patches/patch-istio-crds.yaml"
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Suite")
}

func builZapLogger() (*uzap.Logger, error) {
	config := uzap.NewDevelopmentConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.Encoding = "json"
	config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("Jan 02 15:04:05.000000000")

	return config.Build()
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "config", "crd", "bases"),
			externalDependencyDataPath,
		}, ErrorIfCRDPathMissing: true,
	}

	var err error
	// config is defined in this file globally.
	config, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(config).NotTo(BeNil())

	err = operatorv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(config, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	k8sManager, err := ctrl.NewManager(config, ctrl.Options{
		Scheme: scheme.Scheme,
	})
	Expect(err).ToNot(HaveOccurred())

	mgrLogger, err := builZapLogger()
	Expect(err).NotTo(HaveOccurred())

	objsFile, err := os.Open("../application-connector.yaml")
	Expect(err).ShouldNot(HaveOccurred())

	objsData, err := yaml.LoadData(objsFile)
	Expect(err).ShouldNot(HaveOccurred())

	depsFile, err := os.Open("../application-connector-dependencies.yaml")
	Expect(err).ShouldNot(HaveOccurred())

	depsData, err := yaml.LoadData(depsFile)
	Expect(err).ShouldNot(HaveOccurred())

	optionalFile, err := os.Open("../application-connector-optional.yaml")
	Expect(err).ShouldNot(HaveOccurred())

	optionalData, err := yaml.LoadData(optionalFile)
	Expect(err).ShouldNot(HaveOccurred())

	err = (&applicationConnectorReconciler{
		log: mgrLogger.Sugar(),
		K8s: reconciler.K8s{
			Client:        k8sManager.GetClient(),
			EventRecorder: record.NewFakeRecorder(100),
		},
		Cfg: reconciler.Cfg{
			Finalizer:    "application-connector-manager.kyma-project.io/deletion-hook",
			Objs:         objsData,
			Deps:         depsData,
			OptionalObjs: optionalData,
		},
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		defer GinkgoRecover()

		ctx, cancel = context.WithCancel(context.Background())

		err = k8sManager.Start(ctx)
		Expect(err).ToNot(HaveOccurred(), "failed to run manager")
	}()

})

var _ = AfterSuite(func() {
	cancel()

	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
