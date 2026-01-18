package controllers

import (
	"context"
	"time"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/log"
	ctrl "sigs.k8s.io/controller-runtime"
)

type PodReconciler struct {
    client.Client
}

var _ reconcile.Reconciler = &PodReconciler{}

func (r *PodReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := log.FromContext(ctx)
	pod := &corev1.Pod{}
	err := r.Get(ctx, req.NamespacedName, pod)
	if apierrors.IsNotFound(err) {
		return reconcile.Result{}, nil
	}

	if err != nil {
		logger.Error(err, "Failed to get Pod")
		return reconcile.Result{}, err
	}

	if !isPodManagedByStatefulSet(pod) {
		return reconcile.Result{}, nil
	}

	// Ensure the Pod is in the "price-engine" namespace
	if pod.Namespace != "app" {
		logger.Info("Pod is not in application namespaces, skipping", "namespace", pod.Namespace)
		return reconcile.Result{}, nil
	}

	if pod.ObjectMeta.DeletionTimestamp != nil {
		nodeName := pod.Spec.NodeName
		var node corev1.Node
		nodeErr := r.Get(ctx, client.ObjectKey{Name: nodeName}, &node)
		if apierrors.IsNotFound(nodeErr) || !isNodeReady(&node) {
			logger.Info("Pod's node unhealthy or not ready, force deleting pod", "pod", pod.Name, "node", nodeName)
			if err := r.Delete(ctx, pod, client.GracePeriodSeconds(0)); err != nil {
				return reconcile.Result{}, err
			}
			return reconcile.Result{}, nil
		}

		if nodeErr != nil {
			logger.Error(nodeErr, "Failed to get node", "node", nodeName)
			return reconcile.Result{}, nodeErr
		}

		logger.Info("Pod status is terminating, but node is healthy", "pod", pod.Name)
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}

	return reconcile.Result{}, nil
}

func (r *PodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		Complete(r)
}

func isPodManagedByStatefulSet(pod *corev1.Pod) bool {
    for _, owner := range pod.OwnerReferences {
        if owner.Kind == "StatefulSet" && owner.Controller != nil && *owner.Controller {
            return true
        }
    }
    return false
}

// isNodeReady checks if the node is in Ready state
func isNodeReady(node *corev1.Node) bool {
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}
