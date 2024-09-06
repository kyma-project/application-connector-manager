package main

import (
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type options struct {
	apiServerURL                string
	applicationSecretsNamespace string
	externalAPIPort             int
	kubeConfig                  string
	logLevel                    *zapcore.Level
	proxyCacheTTL               int
	proxyPort                   int
	proxyPortCompass            int
	proxyTimeout                int
	requestTimeout              int
	secretCacheRetention        time.Duration
	applicationCacheRetention   time.Duration
}

func parseArgs(rootCmd *cobra.Command, opts *options, setupLogger *zap.Logger) {
	rootCmd.Flags().StringVar(&opts.apiServerURL, "apiServerURL", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	rootCmd.Flags().StringVar(&opts.applicationSecretsNamespace, "applicationSecretsNamespace", "kyma-system", "Namespace where Application secrets used by the Application Gateway exist")
	rootCmd.Flags().IntVar(&opts.externalAPIPort, "externalAPIPort", 8081, "Port that exposes the API which allows checking the component status and exposes log configuration")
	rootCmd.Flags().StringVar(&opts.kubeConfig, "kubeConfig", "", "Path to a kubeconfig. Only required if out-of-cluster")
	opts.logLevel = zap.LevelFlag("logLevel", zap.InfoLevel, "Log level: panic | fatal | error | warn | info | debug. Can't be lower than info")
	rootCmd.Flags().IntVar(&opts.proxyCacheTTL, "proxyCacheTTL", 120, "TTL, in seconds, for proxy cache of Remote API information")
	rootCmd.Flags().IntVar(&opts.proxyPort, "proxyPort", 8080, "Port that acts as a proxy for the calls from services and Functions to an external solution in the default standalone mode or Compass bundles with a single API definition")
	rootCmd.Flags().IntVar(&opts.proxyPortCompass, "proxyPortCompass", 8082, "Port that acts as a proxy for the calls from services and Functions to an external solution in the Compass mode")
	rootCmd.Flags().IntVar(&opts.proxyTimeout, "proxyTimeout", 10, "Timeout for requests sent through the proxy, expressed in seconds")
	rootCmd.Flags().IntVar(&opts.requestTimeout, "requestTimeout", 10, "Timeout for requests sent through Central Application Gateway, expressed in seconds")

	rootCmd.Flags().DurationVar(&opts.secretCacheRetention, "secretCacheRetention", time.Minute*5, "Retention time how long a secret is cached by the Central Application Gateway")
	rootCmd.Flags().DurationVar(&opts.applicationCacheRetention, "applicationCacheRetention", time.Minute*5, "Retention time how long an application is cached by the Central Application Gateway")

	opts.logArgs(setupLogger)
}

func (o options) logArgs(log *zap.Logger) {
	log.Info("Parsed flags",
		zap.String("-apiServerURL", o.apiServerURL),
		zap.String("-applicationSecretsNamespace", o.applicationSecretsNamespace),
		zap.Int("-externalAPIPort", o.externalAPIPort),
		zap.String("-kubeConfig", o.kubeConfig),
		zap.String("-logLevel", o.logLevel.String()),
		zap.Int("-proxyCacheTTL", o.proxyCacheTTL),
		zap.Int("-proxyPort", o.proxyPort),
		zap.Int("-proxyPortCompass", o.proxyPortCompass),
		zap.Int("-proxyTimeout", o.proxyTimeout),
		zap.Int("-requestTimeout", o.requestTimeout),
		zap.Duration("-secretCacheRetention", o.secretCacheRetention),
		zap.Duration("-applicationCacheRetention", o.applicationCacheRetention),
	)

}
