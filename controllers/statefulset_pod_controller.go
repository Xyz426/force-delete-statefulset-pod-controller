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
		logger.Info("Pod not found")
		return reconcile.Result{}, nil
	}

	if err != nil {
		logger.Error(err, "Failed to get Pod")
		return reconcile.Result{}, err
	}

	if !isPodManagedByStatefulSet(pod) {
		return reconcile.Result{}, nil
	}

	if pod.ObjectMeta.DeletionTimestamp != nil {
		nodeName := pod.Spec.NodeName
		if nodeName == "" {
			logger.Info("Pod has not been scheduled to a node, skipping force delete", "pod", pod.Name)
			return reconcile.Result{}, nil
		}
		var node corev1.Node
		nodeErr := r.Get(ctx, client.ObjectKey{Name: nodeName}, &node)
		if apierrors.IsNotFound(nodeErr) {
			logger.Info("Pod's node unhealthy, force deleting pod", "pod", pod.Name, "node", nodeName)
			err = r.Delete(ctx, pod, client.GracePeriodSeconds(0))
			if err != nil {
				logger.Error(err, "Failed to force delete Statefulset pod", "pod", pod.Name, "node", nodeName)
				return reconcile.Result{}, err
			}
		}

		if nodeErr != nil {
			logger.Error(nodeErr, "Failed to get node", "node", nodeName)
			return reconcile.Result{}, nodeErr
		}

		logger.Info("Pod still terminating, but node exists; will try again", "pod", pod.Name)
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
		if owner.Kind == "StatefulSet" {
			return true
		}
	}
	return false
}
