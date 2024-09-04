package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"k8s.io/apimachinery/pkg/runtime"

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

	// Create a new Manager to manage the controllers
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme, // Attach the created scheme
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create manager: %v\n", err)
		os.Exit(1)
	}

	// Register the TargetGroupBinding API types with the manager's scheme
	if err := targetgroupv1.AddToScheme(mgr.GetScheme()); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to register TargetGroupBinding scheme: %v\n", err)
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
