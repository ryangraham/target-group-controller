package controllers

import (
	"context"
	"fmt"

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

// Reconcile function manages adding/removing nodes to/from the AWS target group
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

	// Determine which nodes need to be added or removed from the target group
	nodesToAdd, nodesToRemove := r.getTargetChanges(nodeList.Items, currentTargets)

	// Add nodes to the target group
	if err := r.addNodesToTargetGroup(tgb.Spec.TargetGroupARN, nodesToAdd); err != nil {
		return ctrl.Result{}, err
	}

	// Remove nodes from the target group
	if err := r.removeNodesFromTargetGroup(tgb.Spec.TargetGroupARN, nodesToRemove); err != nil {
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

// getTargetChanges compares the node list with the current targets to determine nodes to add and remove
func (r *TargetGroupBindingReconciler) getTargetChanges(nodes []corev1.Node, targets []elbv2types.TargetHealthDescription) (nodesToAdd []corev1.Node, nodesToRemove []elbv2types.TargetDescription) {
	nodeIPMap := make(map[string]bool)
	for _, node := range nodes {
		nodeIP := node.Status.Addresses[0].Address
		nodeIPMap[nodeIP] = true
		if !r.isNodeInTargetGroup(nodeIP, targets) {
			nodesToAdd = append(nodesToAdd, node)
		}
	}

	for _, target := range targets {
		if _, exists := nodeIPMap[*target.Target.Id]; !exists {
			nodesToRemove = append(nodesToRemove, *target.Target)
		}
	}

	return nodesToAdd, nodesToRemove
}

// isNodeInTargetGroup checks if a node is already a target in the target group
func (r *TargetGroupBindingReconciler) isNodeInTargetGroup(nodeIP string, targets []elbv2types.TargetHealthDescription) bool {
	for _, target := range targets {
		if *target.Target.Id == nodeIP {
			return true
		}
	}
	return false
}

// addNodesToTargetGroup adds nodes to the target group
func (r *TargetGroupBindingReconciler) addNodesToTargetGroup(targetGroupARN string, nodes []corev1.Node) error {
	if len(nodes) == 0 {
		return nil
	}
	targets := []elbv2types.TargetDescription{}
	for _, node := range nodes {
		nodeIP := node.Status.Addresses[0].Address
		targets = append(targets, elbv2types.TargetDescription{Id: aws.String(nodeIP)})
	}
	input := &elasticloadbalancingv2.RegisterTargetsInput{
		TargetGroupArn: aws.String(targetGroupARN),
		Targets:        targets,
	}
	_, err := r.Elbv2Client.RegisterTargets(context.TODO(), input)
	if err != nil {
		return err
	}
	fmt.Printf("Added nodes to target group %s\n", targetGroupARN)
	return nil
}

// removeNodesFromTargetGroup removes nodes from the target group
func (r *TargetGroupBindingReconciler) removeNodesFromTargetGroup(targetGroupARN string, targets []elbv2types.TargetDescription) error {
	if len(targets) == 0 {
		return nil
	}
	input := &elasticloadbalancingv2.DeregisterTargetsInput{
		TargetGroupArn: aws.String(targetGroupARN),
		Targets:        targets,
	}
	_, err := r.Elbv2Client.DeregisterTargets(context.TODO(), input)
	if err != nil {
		return err
	}
	fmt.Printf("Removed nodes from target group %s\n", targetGroupARN)
	return nil
}

// SetupWithManager sets up the controller with the Manager
func (r *TargetGroupBindingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.TargetGroupBinding{}).
		Complete(r)
}
