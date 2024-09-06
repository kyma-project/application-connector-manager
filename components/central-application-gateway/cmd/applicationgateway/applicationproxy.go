package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/kyma-project/kyma/components/central-application-gateway/internal/csrf"
	csrfClient "github.com/kyma-project/kyma/components/central-application-gateway/internal/csrf/client"
	csrfStrategy "github.com/kyma-project/kyma/components/central-application-gateway/internal/csrf/strategy"
	"github.com/kyma-project/kyma/components/central-application-gateway/internal/externalapi"
	"github.com/kyma-project/kyma/components/central-application-gateway/internal/metadata"
	"github.com/kyma-project/kyma/components/central-application-gateway/internal/metadata/applications"
	"github.com/kyma-project/kyma/components/central-application-gateway/internal/metadata/secrets"
	"github.com/kyma-project/kyma/components/central-application-gateway/internal/metadata/serviceapi"
	"github.com/kyma-project/kyma/components/central-application-gateway/internal/proxy"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/apperrors"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/httptools"
	"github.com/oklog/run"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	shutdownTimeout = 2 * time.Second
)

func main() {
	setupLogger := zap.Must(zap.NewProduction())
	defer func(setupLogger *zap.Logger) {
		err := setupLogger.Sync()
		if err != nil && !errors.Is(err, syscall.ENOTTY) {
			panic(fmt.Sprintf("Failed to synchronize logger: %v", err))
		}
	}(setupLogger)

	initViper(setupLogger)

	//init command
	opts := &options{}
	var rootCmd = &cobra.Command{
		Use:   "acm",
		Short: "Root short description",
		Long:  "Root long description",
		Run: func(cmd *cobra.Command, args []string) {
			runCmd(setupLogger, opts)
		},
	}

	//parse args
	parseArgs(rootCmd, opts, setupLogger)

	if err := rootCmd.Execute(); err != nil {
		setupLogger.Error("Failed to execute command: %v", zap.Error(err))
	}
}

func initViper(setupLogger *zap.Logger) {
	viper.SetEnvPrefix("ACM_GATEWAY")
	viper.AutomaticEnv()

	configFile := viper.GetString("config")
	if len(configFile) > 0 {
		err := viper.ReadInConfig()
		if err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				setupLogger.Error("Config file '%s' not found. Check the path provided in env-var '%s'",
					zap.String("configFile", configFile),
					zap.String("envVar", strings.ToUpper(fmt.Sprintf("%s_%s", viper.GetEnvPrefix(), "config"))))
			} else {
				setupLogger.Panic("Failed to parse config file: %v", zap.Error(err))
			}
		}
	}
}

func runCmd(setupLogger *zap.Logger, options *options) {

	setupLogger.Info("Starting Application Gateway")

	logCfg := zap.NewProductionConfig()
	logCfg.Level.SetLevel(*options.logLevel)

	log, err := logCfg.Build()
	zap.ReplaceGlobals(log)
	defer func(log *zap.Logger) {
		err := log.Sync()
		if err != nil {
			panic(err)
		}
	}(log)

	if err != nil {
		setupLogger.Fatal("Couldn't initiate logger", zap.Error(err))
	}

	k8sConfig, err := clientcmd.BuildConfigFromFlags(options.apiServerURL, options.kubeConfig)
	if err != nil {
		log.Fatal("Error reading in cluster config", zap.Error(err))
	}

	coreClientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		log.Fatal("Error creating core clientset", zap.Error(err))
	}

	serviceDefinitionService, err := newServiceDefinitionService(
		k8sConfig,
		coreClientset,
		options,
	)
	if err != nil {
		log.Fatal("Unable to create ServiceDefinitionService:'", zap.Error(err))
	}

	internalHandler := newInternalHandler(serviceDefinitionService, *options)
	internalHandlerForCompass := newInternalHandlerForCompass(serviceDefinitionService, *options)
	externalHandler := externalapi.NewHandler(logCfg.Level)

	internalHandler = httptools.RequestLogger("Internal handler: ", internalHandler)
	internalHandlerForCompass = httptools.RequestLogger("Internal handler: ", internalHandlerForCompass)
	externalHandler = httptools.RequestLogger("External handler: ", externalHandler)

	externalSrv := &http.Server{
		Addr:         ":" + strconv.Itoa(options.externalAPIPort),
		Handler:      externalHandler,
		ReadTimeout:  time.Duration(options.requestTimeout) * time.Second,
		WriteTimeout: time.Duration(options.requestTimeout) * time.Second,
	}

	internalSrv := &http.Server{
		Addr:        ":" + strconv.Itoa(options.proxyPort),
		Handler:     internalHandler,
		ReadTimeout: time.Duration(options.requestTimeout) * time.Second,
	}

	internalSrvCompass := &http.Server{
		Addr:        ":" + strconv.Itoa(options.proxyPortCompass),
		Handler:     internalHandlerForCompass,
		ReadTimeout: time.Duration(options.requestTimeout) * time.Second,
	}

	var g run.Group

	addHttpServerToRunGroup("external-api", &g, externalSrv)
	addHttpServerToRunGroup("proxy-kyma-os", &g, internalSrv)
	addHttpServerToRunGroup("proxy-kyma-mps", &g, internalSrvCompass)
	addInterruptSignalToRunGroup(&g)

	err = g.Run()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal("Servers encountered error", zap.Error(err))
	}
}

func addHttpServerToRunGroup(name string, g *run.Group, srv *http.Server) {
	log := zap.L().Sugar()

	log.Infof("Starting %s HTTP server on %s", name, srv.Addr)
	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		log.Fatalf("Unable to start %s HTTP server: '%s'", name, err)
	}
	g.Add(func() error {
		defer log.Infof("Server %s finished", name)
		return srv.Serve(ln)
	}, func(error) {
		log.Infof("Shutting down %s HTTP server on %s", name, srv.Addr)

		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		err = srv.Shutdown(ctx)
		if err != nil && err != http.ErrServerClosed {
			log.Warnf("HTTP server shutdown %s failed: %s", name, err)
		}
	})
}

func addInterruptSignalToRunGroup(g *run.Group) {
	cancelInterrupt := make(chan struct{})
	g.Add(func() error {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		select {
		case <-cancelInterrupt:
		case sig := <-c:
			zap.L().Sugar().Infof("received signal %s", sig)
		}
		return nil
	}, func(error) {
		close(cancelInterrupt)
	})
}

func newInternalHandler(serviceDefinitionService metadata.ServiceDefinitionService, options options) http.Handler {
	authStrategyFactory := newAuthenticationStrategyFactory(options.proxyTimeout)
	csrfCl := newCSRFClient(options.proxyTimeout)
	csrfTokenStrategyFactory := csrfStrategy.NewTokenStrategyFactory(csrfCl)

	return proxy.New(serviceDefinitionService, authStrategyFactory, csrfTokenStrategyFactory, getProxyConfig(options))
}

func newInternalHandlerForCompass(serviceDefinitionService metadata.ServiceDefinitionService, options options) http.Handler {
	authStrategyFactory := newAuthenticationStrategyFactory(options.proxyTimeout)
	csrfCl := newCSRFClient(options.proxyTimeout)
	csrfTokenStrategyFactory := csrfStrategy.NewTokenStrategyFactory(csrfCl)

	return proxy.NewForCompass(serviceDefinitionService, authStrategyFactory, csrfTokenStrategyFactory, getProxyConfig(options))
}

func getProxyConfig(options options) proxy.Config {
	return proxy.Config{
		ProxyTimeout:  options.proxyTimeout,
		ProxyCacheTTL: options.proxyCacheTTL,
	}
}

func newAuthenticationStrategyFactory(oauthClientTimeout int) authorization.StrategyFactory {
	return authorization.NewStrategyFactory(authorization.FactoryConfiguration{
		OAuthClientTimeout: oauthClientTimeout,
	})
}

func newServiceDefinitionService(k8sConfig *restclient.Config, coreClientset kubernetes.Interface, options *options) (metadata.ServiceDefinitionService, error) {
	applicationServiceRepository, apperror := newApplicationRepository(k8sConfig, options.applicationCacheRetention)
	if apperror != nil {
		return nil, apperror
	}

	secretsRepository := newSecretsRepository(coreClientset, options.applicationSecretsNamespace, options.secretCacheRetention)

	serviceAPIService := serviceapi.NewService(secretsRepository)

	return metadata.NewServiceDefinitionService(serviceAPIService, applicationServiceRepository), nil
}

func newApplicationRepository(config *restclient.Config, cacheRetention time.Duration) (applications.ServiceRepository, apperrors.AppError) {
	applicationClientset, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, apperrors.Internal("failed to create k8s application client, %s", err)
	}

	rei := applicationClientset.ApplicationconnectorV1alpha1().Applications()

	return applications.NewServiceRepository(rei, cacheRetention), nil
}

func newSecretsRepository(coreClientset kubernetes.Interface, namespace string, cacheRetention time.Duration) secrets.Repository {
	sei := coreClientset.CoreV1().Secrets(namespace)

	return secrets.NewRepository(sei, cacheRetention)
}

func newCSRFClient(timeout int) csrf.Client {
	cache := csrfClient.NewTokenCache()
	return csrfClient.New(timeout, cache)
}
