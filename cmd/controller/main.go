package main

import (
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	v1 "github.com/ryangraham/target-group-controller/pkg/api/v1"
	"github.com/ryangraham/target-group-controller/pkg/controllers"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	// Register the TargetGroupBinding API with the scheme
	utilruntime.Must(v1.AddToScheme(scheme))
}

func main() {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	awsConfig := aws.Config{} // Initialize AWS config here

	elbv2Client := elasticloadbalancingv2.NewFromConfig(awsConfig)
	taggingClient := resourcegroupstaggingapi.NewFromConfig(awsConfig)

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		HealthProbeBindAddress: ":8081", // Health check endpoint
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Add health checks
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}

	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up readiness check")
		os.Exit(1)
	}

	// Setup the reconciler for TargetGroupBinding
	if err := (&controllers.TargetGroupBindingReconciler{
		Client:                mgr.GetClient(),
		Scheme:                mgr.GetScheme(),
		Elbv2Client:           elbv2Client,
		ResourceTaggingClient: taggingClient,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "TargetGroupBinding")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
