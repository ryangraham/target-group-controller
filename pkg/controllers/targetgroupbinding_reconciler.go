package controllers

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	elbv2types "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	v1 "github.com/ryangraham/target-group-controller/pkg/api/v1"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// TargetGroupBindingReconciler reconciles a TargetGroupBinding object
type TargetGroupBindingReconciler struct {
	client.Client
	Elbv2Client *elasticloadbalancingv2.Client
}

// Reconcile function manages adding nodes to the AWS target group
func (r *TargetGroupBindingReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var tgb v1.TargetGroupBinding

	// Fetch the TargetGroupBinding resource
	if err := r.Get(ctx, req.NamespacedName, &tgb); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// List all nodes matching the nodeSelector
	nodeList := &corev1.NodeList{}
	nodeSelector := client.MatchingLabels(tgb.Spec.NodeSelector)
	if err := r.List(ctx, nodeList, nodeSelector); err != nil {
		return ctrl.Result{}, err
	}

	// Get the current targets in the AWS target group
	currentTargets, err := r.getTargetGroupTargets(tgb.Spec.TargetGroupARN)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Determine which nodes need to be added to the target group
	nodesToAdd := r.getNodesToAdd(nodeList.Items, currentTargets)

	// Add nodes to the target group
	if err := r.addNodesToTargetGroup(tgb.Spec.TargetGroupARN, nodesToAdd); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// getTargetGroupTargets fetches the current targets in the AWS target group
func (r *TargetGroupBindingReconciler) getTargetGroupTargets(targetGroupARN string) ([]elbv2types.TargetHealthDescription, error) {
	input := &elasticloadbalancingv2.DescribeTargetHealthInput{
		TargetGroupArn: aws.String(targetGroupARN),
	}
	result, err := r.Elbv2Client.DescribeTargetHealth(context.TODO(), input)
	if err != nil {
		return nil, err
	}
	return result.TargetHealthDescriptions, nil
}

// getNodesToAdd compares the node list with the current targets to determine nodes to add
func (r *TargetGroupBindingReconciler) getNodesToAdd(nodes []corev1.Node, targets []elbv2types.TargetHealthDescription) []corev1.Node {
	var nodesToAdd []corev1.Node
	for _, node := range nodes {
		instanceID := getInstanceIDFromNode(node)
		if instanceID == "" {
			continue
		}
		if !r.isNodeInTargetGroup(instanceID, targets) {
			nodesToAdd = append(nodesToAdd, node)
		}
	}
	return nodesToAdd
}

// isNodeInTargetGroup checks if a node is already a target in the target group
func (r *TargetGroupBindingReconciler) isNodeInTargetGroup(instanceID string, targets []elbv2types.TargetHealthDescription) bool {
	for _, target := range targets {
		if *target.Target.Id == instanceID {
			return true
		}
	}
	return false
}

// getInstanceIDFromNode extracts the EC2 instance ID from the node object
func getInstanceIDFromNode(node corev1.Node) string {
	providerID := node.Spec.ProviderID
	// Check if the providerID contains 'aws://'
	if strings.HasPrefix(providerID, "aws://") {
		// Split by '/' and return the last part, which is the instance ID
		parts := strings.Split(providerID, "/")
		if len(parts) > 0 {
			return parts[len(parts)-1]
		}
	}
	return ""
}

// addNodesToTargetGroup adds nodes (by instance ID) to the target group
func (r *TargetGroupBindingReconciler) addNodesToTargetGroup(targetGroupARN string, nodes []corev1.Node) error {
	if len(nodes) == 0 {
		return nil
	}
	targets := []elbv2types.TargetDescription{}
	for _, node := range nodes {
		instanceID := getInstanceIDFromNode(node)
		if instanceID == "" {
			continue
		}
		targets = append(targets, elbv2types.TargetDescription{Id: aws.String(instanceID)})
	}
	input := &elasticloadbalancingv2.RegisterTargetsInput{
		TargetGroupArn: aws.String(targetGroupARN),
		Targets:        targets,
	}
	_, err := r.Elbv2Client.RegisterTargets(context.TODO(), input)
	if err != nil {
		return err
	}
	fmt.Printf("Added instance IDs to target group %s\n", targetGroupARN)
	return nil
}

// SetupWithManager sets up the controller with the Manager
func (r *TargetGroupBindingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.TargetGroupBinding{}).
		Complete(r)
}
