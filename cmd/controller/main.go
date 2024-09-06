package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"k8s.io/apimachinery/pkg/runtime"

	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	// Import your custom API package
	targetgroupv1 "github.com/ryangraham/target-group-controller/pkg/api/v1"

	// Import your controller logic
	"github.com/ryangraham/target-group-controller/pkg/controllers"
)

func main() {
	// Set up logging with Zap (a popular logging library)
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	// Load AWS configuration (region, credentials, etc.)
	awsCfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to load AWS config: %v\n", err)
		os.Exit(1)
	}

	// Create the ELBv2 client using the AWS SDK v2
	elbClient := elasticloadbalancingv2.NewFromConfig(awsCfg)

	// Create a new runtime Scheme
	scheme := runtime.NewScheme()

	// Register core Kubernetes types into the scheme
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to add core types to scheme: %v\n", err)
		os.Exit(1)
	}

	// Register your custom TargetGroupBinding types into the scheme
	if err := targetgroupv1.AddToScheme(scheme); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to add custom types to scheme: %v\n", err)
		os.Exit(1)
	}

	// Create a new Manager to manage the controllers
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme, // Attach the created scheme
		HealthProbeBindAddress: ":8081",
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create manager: %v\n", err)
		os.Exit(1)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to set up health check: %v\n", err)
		os.Exit(1)
	}

	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to set up readiness check: %v\n", err)
		os.Exit(1)
	}

	// Set up the TargetGroupBinding controller with the manager and AWS ELBv2 client
	if err := (&controllers.TargetGroupBindingReconciler{
		Client:      mgr.GetClient(),
		Elbv2Client: elbClient, // Your AWS client
	}).SetupWithManager(mgr); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create TargetGroupBinding controller: %v\n", err)
		os.Exit(1)
	}

	// Start the manager (this blocks until the manager exits)
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to run manager: %v\n", err)
		os.Exit(1)
	}
}
