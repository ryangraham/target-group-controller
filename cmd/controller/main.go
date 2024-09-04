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
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	targetgroupv1 "github.com/ryangraham/target-group-controller/pkg/api/v1"
	"github.com/ryangraham/target-group-controller/pkg/controllers"
)

var (
	// Define the global scheme
	scheme = runtime.NewScheme()
)

func init() {
	// Register Kubernetes built-in types
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	// Register custom types (TargetGroupBinding)
	utilruntime.Must(targetgroupv1.AddToScheme(scheme))
}

func main() {
	// Set up a new logger using Zap (a popular logging library)
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	// Load the AWS configuration
	awsCfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to load AWS SDK config, %v", err)
		os.Exit(1)
	}

	// Create the ELBv2 client using the AWS SDK v2
	elbClient := elasticloadbalancingv2.NewFromConfig(awsCfg)

	// Create a new controller manager, which will manage controllers and start them when ready
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to start manager, %v", err)
		os.Exit(1)
	}

	// Set up the TargetGroupBinding controller with the AWS SDK client and the controller manager
	if err := (&controllers.TargetGroupBindingReconciler{
		Client:      mgr.GetClient(),
		Elbv2Client: elbClient,
	}).SetupWithManager(mgr); err != nil {
		fmt.Fprintf(os.Stderr, "unable to create controller, %v", err)
		os.Exit(1)
	}

	// Start the manager and block until it exits
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		fmt.Fprintf(os.Stderr, "unable to run the manager, %v", err)
		os.Exit(1)
	}
}
