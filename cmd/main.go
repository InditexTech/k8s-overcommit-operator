// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"crypto/tls"
	"flag"
	"io"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/filters"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	overcommit "github.com/InditexTech/k8s-overcommit-operator/api/v1alphav1"
	occontroller "github.com/InditexTech/k8s-overcommit-operator/internal/controller/overcommitclass"
	"github.com/InditexTech/k8s-overcommit-operator/internal/metrics"
	"github.com/InditexTech/k8s-overcommit-operator/internal/utils"

	overcommitcontroller "github.com/InditexTech/k8s-overcommit-operator/internal/controller/overcommit"
	webhookcorev1mutating "github.com/InditexTech/k8s-overcommit-operator/internal/webhook/v1alphav1/mutating"
	webhookcorev1validating "github.com/InditexTech/k8s-overcommit-operator/internal/webhook/v1alphav1/validating"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

// tlsKeyLogger implements io.Writer to log TLS master keys at info level
// WARNING: This logs sensitive cryptographic material and should NEVER be used in production
type tlsKeyLogger struct {
	logger logr.Logger
}

func (w *tlsKeyLogger) Write(p []byte) (n int, err error) {
	// Log TLS master keys - WARNING: Contains sensitive data
	w.logger.Info("TLS master key", "keylog", string(p))
	return len(p), nil
}

// createTLSKeyLogger creates a writer that logs TLS master keys for debugging
// WARNING: This should NEVER be enabled in production as it logs sensitive cryptographic material
func createTLSKeyLogger() io.Writer {
	return &tlsKeyLogger{logger: setupLog}
}

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(overcommit.AddToScheme(scheme))
	utilruntime.Must(certmanagerv1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var secureMetrics bool
	var enableHTTP2 bool
	var tlsOpts []func(*tls.Config)
	flag.StringVar(&metricsAddr, "metrics-bind-address", "0", "The address the metrics endpoint binds to. "+
		"Use :8443 for HTTPS or :8080 for HTTP, or leave as 0 to disable the metrics service.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&secureMetrics, "metrics-secure", false,
		"If set, the metrics endpoint is served securely via HTTPS. Use --metrics-secure=false to use HTTP instead.")
	flag.BoolVar(&enableHTTP2, "enable-http2", false,
		"If set, HTTP/2 will be enabled for the metrics and webhook servers")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	// Check if TLS key logging is enabled (for debugging only)
	enableTLSKeyLogging := os.Getenv("ENABLE_TLS_KEY_LOGGING") == "true"
	if enableTLSKeyLogging {
		setupLog.Info("WARNING: TLS master key logging is ENABLED - DO NOT use in production!")
	}

	// if the enable-http2 flag is false (the default), http/2 should be disabled
	// due to its vulnerabilities. More specifically, disabling http/2 will
	// prevent from being vulnerable to the HTTP/2 Stream Cancellation and
	// Rapid Reset CVEs. For more information see:
	// - https://github.com/advisories/GHSA-qppj-fm5r-hxr3
	// - https://github.com/advisories/GHSA-4374-p667-p6c8
	disableHTTP2 := func(c *tls.Config) {
		setupLog.Info("disabling http/2")
		c.NextProtos = []string{"http/1.1"}
	}

	// Force TLS version 1.2 instead of allowing TLS 1.3
	setTLSVersion := func(c *tls.Config) {
		setupLog.Info("setting TLS version to 1.2")
		c.MinVersion = tls.VersionTLS12
		c.MaxVersion = tls.VersionTLS12

		// Add TLS key logging if enabled (WARNING: Only for debugging)
		if enableTLSKeyLogging {
			c.KeyLogWriter = createTLSKeyLogger()
			setupLog.Info("TLS master key logging enabled for debugging")
		}
	}

	if !enableHTTP2 {
		tlsOpts = append(tlsOpts, disableHTTP2)
	}

	// Apply TLS 1.2 configuration
	tlsOpts = append(tlsOpts, setTLSVersion)

	// Metrics endpoint is enabled in 'config/default/kustomization.yaml'. The Metrics options configure the server.
	// More info:
	// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/metrics/server
	// - https://book.kubebuilder.io/reference/metrics.html
	metricsServerOptions := metricsserver.Options{
		BindAddress:   metricsAddr,
		SecureServing: secureMetrics,
		// TODO(user): TLSOpts is used to allow configuring the TLS config used for the server. If certificates are
		// not provided, self-signed certificates will be generated by default. This option is not recommended for
		// production environments as self-signed certificates do not offer the same level of trust and security
		// as certificates issued by a trusted Certificate Authority (CA). The primary risk is potentially allowing
		// unauthorized access to sensitive metrics data. Consider replacing with CertDir, CertName, and KeyName
		// to provide certificates, ensuring the server communicates using trusted and secure certificates.
		TLSOpts: tlsOpts,
	}

	if secureMetrics {
		// FilterProvider is used to protect the metrics endpoint with authn/authz.
		// These configurations ensure that only authorized users and service accounts
		// can access the metrics endpoint. The RBAC are configured in 'config/rbac/kustomization.yaml'. More info:
		// https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/metrics/filters#WithAuthenticationAndAuthorization
		metricsServerOptions.FilterProvider = filters.WithAuthenticationAndAuthorization
	}
	deploymentName, err := utils.GetPodDeploymentName()
	if err != nil {
		setupLog.Error(err, "unable to get pod deployment name")
		os.Exit(1)
	}

	mgrOptions := ctrl.Options{
		Scheme:                  scheme,
		Metrics:                 metricsServerOptions,
		HealthProbeBindAddress:  probeAddr,
		LeaderElection:          enableLeaderElection,
		LeaderElectionID:        deploymentName + ".inditex.dev",
		LeaderElectionNamespace: os.Getenv("POD_NAMESPACE"),
	}
	// nolint:goconst
	if os.Getenv("ENABLE_POD_MUTATING_WEBHOOK") == "true" || os.Getenv("ENABLE_POD_VALIDATING_WEBHOOK") == "true" || os.Getenv("ENABLE_OC_VALIDATING_WEBHOOK") == "true" {
		webhookServer := webhook.NewServer(webhook.Options{
			TLSOpts: tlsOpts,
			CertDir: os.Getenv("WEBHOOK_CERT_DIR"),
		})

		mgrOptions.WebhookServer = webhookServer
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), mgrOptions)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if os.Getenv("ENABLE_OVERCOMMIT_CONTROLLER") == "true" {
		// Get the image of the pod
		ctx := context.Background()
		registry, image, tag, err := utils.GetPodImageDetails(ctx, mgr.GetAPIReader())
		if err != nil {
			setupLog.Error(err, "unable to get pod image details")
			os.Exit(1)
		}

		serviceAccountName, err := utils.GetPodServiceAccount(mgr.GetAPIReader())
		if err != nil {
			setupLog.Error(err, "unable to get pod service account")
			os.Exit(1)
		}
		err = os.Setenv("SERVICE_ACCOUNT_NAME", serviceAccountName)
		if err != nil {
			setupLog.Error(err, "unable to set service account name")
			os.Exit(1)
		}
		err = os.Setenv("IMAGE_REGISTRY", registry)
		if err != nil {
			setupLog.Error(err, "unable to set image registry")
			os.Exit(1)
		}
		err = os.Setenv("IMAGE_REPOSITORY", image)
		if err != nil {
			setupLog.Error(err, "unable to set image repository")
			os.Exit(1)
		}
		err = os.Setenv("APP_VERSION", tag)
		if err != nil {
			setupLog.Error(err, "unable to set app version")
			os.Exit(1)
		}

		setupLog.Info("Enabling bootstrap controller")
		if err = (&overcommitcontroller.OvercommitReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "Overcommit")
			os.Exit(1)
		}
	}

	// nolint:goconst
	if os.Getenv("ENABLE_OVERCOMMIT_CLASS_CONTROLLER") == "true" {
		setupLog.Info("Enabling overcommit class controller")
		// Register overcommitClass controller
		if err = (&occontroller.OvercommitClassReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "OvercommitClass")
			os.Exit(1)
		}
	}

	// nolint:goconst
	if os.Getenv("ENABLE_POD_MUTATING_WEBHOOK") == "true" {
		setupLog.Info("Enabling pod mutating webhook")
		// Register pod mutating webhook
		if err = webhookcorev1mutating.SetupPodWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create mutating webhook", "webhook", "Pod")
			os.Exit(1)
		}
	}
	// nolint:goconst
	if os.Getenv("ENABLE_POD_VALIDATING_WEBHOOK") == "true" {
		setupLog.Info("Enabling pod validating webhook")
		// Register pod validating webhook
		if err = webhookcorev1validating.SetupPodWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create validating webhook", "webhook", "Pod")
			os.Exit(1)
		}
	}

	// nolint:goconst
	if os.Getenv("ENABLE_OC_VALIDATING_WEBHOOK") == "true" {
		setupLog.Info("Enabling overcommitClass validating webhook")
		// Register overcommitClass validation webhook
		if err = (&overcommit.OvercommitClass{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "OvercommitClass")
			os.Exit(1)
		}
	}

	// +kubebuilder:scaffold:builder
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}
	metrics.K8sOvercommitOperatorVersion.WithLabelValues(os.Getenv("APP_VERSION")).Set(1)
	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
