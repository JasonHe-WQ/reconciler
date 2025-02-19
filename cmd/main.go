package main

import (
	"flag"
	"time"

	"github.com/JasonHe-WQ/reconciler/pkg/controllers"

	fluentbitv1alpha2 "github.com/fluent/fluent-operator/apis/fluentbit/v1alpha2"
	fluentdv1alpha1 "github.com/fluent/fluent-operator/apis/fluentd/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

func main() {
	leaderElection := flag.Bool("enable-leader-election", true,
		"Enable leader election for controller manager")
	leaseDuration := flag.Duration("lease-duration", 15*time.Second,
		"Lease duration for leader election")
	renewDeadline := flag.Duration("renew-deadline", 10*time.Second,
		"Renew deadline for leader election")
	retryPeriod := flag.Duration("retry-period", 2*time.Second,
		"Retry period for leader election")

	flag.Parse()
	ctrl.SetLogger(klog.NewKlogr())
	localScheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(localScheme)
	_ = fluentbitv1alpha2.AddToScheme(localScheme)
	_ = fluentdv1alpha1.AddToScheme(localScheme)
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: localScheme,
		Metrics: server.Options{
			BindAddress: ":8080",
		},
		HealthProbeBindAddress:  "0.0.0.0:8081",
		LeaderElection:          *leaderElection,
		LeaderElectionID:        "rate-limiter-controller-leader-election",
		LeaderElectionNamespace: "default",
		LeaseDuration:           leaseDuration,
		RenewDeadline:           renewDeadline,
		RetryPeriod:             retryPeriod,
	})
	if err != nil {
		klog.Fatal("Rate limiter initialization error: ", err)
	}

	if err = (&controllers.RateLimiterReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("RateLimiter"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		klog.Fatal("unable to setup RateLimiter controller: ", err)
	}
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		klog.Fatal("unable to set up health check: ", err)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		klog.Fatal("unable to set up ready check: ", err)
	}
	klog.Info("Starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		klog.Fatal("Manager start error: ", err)
	}
}
