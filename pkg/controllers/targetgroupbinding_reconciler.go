package controllers

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	elbv2types "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	v1 "github.com/ryangraham/target-group-controller/pkg/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// TargetGroupBindingReconciler reconciles a TargetGroupBinding object
type TargetGroupBindingReconciler struct {
	client.Client
	Elbv2Client *elasticloadbalancingv2.Client
	Scheme      *runtime.Scheme
}

// Reconcile is the main logic for the controller
func (r *TargetGroupBindingReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var tgb v1.TargetGroupBinding

	// Fetch the TargetGroupBinding resource
	if err := r.Get(ctx, req.NamespacedName, &tgb); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Step 1: Find the Target Group by name
	targetGroupARN, err := r.findTargetGroupByName(ctx, tgb.Spec.TargetGroupName)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to find target group: %w", err)
	}

	// Step 2: Get the service endpoints (IPs) to register
	serviceIPs, err := r.getServiceEndpoints(ctx, tgb.Spec.ServiceRef.Name, tgb.Spec.ServiceRef.Namespace)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get service endpoints: %w", err)
	}

	// Step 3: Register IPs with the target group
	if err := r.registerIPsWithTargetGroup(ctx, targetGroupARN, serviceIPs); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to register IPs: %w", err)
	}

	// Update the status with the Target Group ARN and registered IPs
	tgb.Status.TargetGroupARN = targetGroupARN
	tgb.Status.RegisteredIPs = serviceIPs
	tgb.Status.LastSyncTime = metav1.Now()

	if err := r.Status().Update(ctx, &tgb); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// findTargetGroupByName finds a target group by the specified name
func (r *TargetGroupBindingReconciler) findTargetGroupByName(ctx context.Context, targetGroupName string) (string, error) {
	input := &elasticloadbalancingv2.DescribeTargetGroupsInput{
		Names: []string{targetGroupName},
	}

	// Describe the target group by name
	result, err := r.Elbv2Client.DescribeTargetGroups(ctx, input)
	if err != nil {
		return "", err
	}

	if len(result.TargetGroups) == 0 {
		return "", fmt.Errorf("no target groups found with name %s", targetGroupName)
	}

	// Return the first matched target group ARN
	return aws.ToString(result.TargetGroups[0].TargetGroupArn), nil
}

// getServiceEndpoints fetches the IPs of the service endpoints
func (r *TargetGroupBindingReconciler) getServiceEndpoints(ctx context.Context, serviceName string, namespace string) ([]string, error) {
	var svc corev1.Service
	if err := r.Get(ctx, client.ObjectKey{Name: serviceName, Namespace: namespace}, &svc); err != nil {
		return nil, err
	}

	var endpoints corev1.Endpoints
	if err := r.Get(ctx, client.ObjectKey{Name: serviceName, Namespace: namespace}, &endpoints); err != nil {
		return nil, err
	}

	var ips []string
	for _, subset := range endpoints.Subsets {
		for _, address := range subset.Addresses {
			ips = append(ips, address.IP)
		}
	}

	return ips, nil
}

// registerIPsWithTargetGroup registers IP addresses as targets in the specified target group
func (r *TargetGroupBindingReconciler) registerIPsWithTargetGroup(ctx context.Context, targetGroupARN string, ips []string) error {
	if len(ips) == 0 {
		return nil
	}

	var targets []elbv2types.TargetDescription
	for _, ip := range ips {
		targets = append(targets, elbv2types.TargetDescription{
			Id: aws.String(ip),
		})
	}

	_, err := r.Elbv2Client.RegisterTargets(ctx, &elasticloadbalancingv2.RegisterTargetsInput{
		TargetGroupArn: aws.String(targetGroupARN),
		Targets:        targets,
	})
	return err
}

// SetupWithManager sets up the controller with the Manager
func (r *TargetGroupBindingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.TargetGroupBinding{}).
		Complete(r)
}
