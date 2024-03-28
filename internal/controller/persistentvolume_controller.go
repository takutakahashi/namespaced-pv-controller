/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"

	namespacedpvv1 "github.com/homirun/namespaced-pv-controller/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// PersistentVolumeReconciler reconciles a PersistentVolume object
type PersistentVolumeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

var FINALIZER_NAME = "namespacedpv.homi.run/pvFinalizer"

//+kubebuilder:rbac:groups=core,resources=persistentvolumes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=persistentvolumes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core,resources=persistentvolumes/finalizers,verbs=update
//+kubebuilder:rbac:groups=namespaced-pv.homi.run,resources=namespacedpvs,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PersistentVolume object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *PersistentVolumeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	pv := &corev1.PersistentVolume{}
	if err := r.Get(ctx, req.NamespacedName, pv); err != nil {
		logger.Error(err, "unable to fetch PersistentVolume")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if pv.Annotations["pv.kubernetes.io/provisioned-by"] != "namespaced-pv-controller" {
		return ctrl.Result{}, nil
	}

	if !pv.GetDeletionTimestamp().IsZero() {
		return ctrl.Result{}, r.DeletePV(ctx, pv)
	}

	if !controllerutil.ContainsFinalizer(pv, FINALIZER_NAME) {
		pvCopy := pv.DeepCopy()
		pvCopy.Finalizers = append(pvCopy.Finalizers, FINALIZER_NAME)
		patch := client.MergeFrom(pv)
		if err := r.Patch(ctx, pvCopy, patch); err != nil {
			logger.Error(err, "unable to patch PersistentVolume")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *PersistentVolumeReconciler) DeletePV(ctx context.Context, pv *corev1.PersistentVolume) error {
	logger := log.FromContext(ctx)

	namespacedPv := &namespacedpvv1.NamespacedPv{}
	if err := r.Get(ctx, client.ObjectKey{Namespace: pv.Labels["owner-namespace"], Name: pv.Labels["owner"]}, namespacedPv); err != nil {
		if client.IgnoreNotFound(err) == nil {
			// namespacedPv がなかったら pv の finalizer を削除して終了
			logger.Info("NamespacedPv not found", "namespace", pv.Labels["owner-namespace"], "name", pv.Labels["owner"])
			controllerutil.RemoveFinalizer(pv, FINALIZER_NAME)
			if err := r.Update(ctx, pv); err != nil {
				logger.Error(err, "unable to update PersistentVolume")
				return err
			}
			logger.Info("pv deleted", "namespace", pv.Namespace, "name", pv.Name)
			return nil
		}
		logger.Error(err, "unable to fetch NamespacedPv")
		return err
	}

	if pv.Status.Phase == corev1.VolumeReleased && pv.Spec.PersistentVolumeReclaimPolicy == corev1.PersistentVolumeReclaimDelete {
		pvCopy := pv.DeepCopy()
		pvCopy.Spec.ClaimRef = nil
		patch := client.MergeFrom(pv)
		if err := r.Patch(ctx, pvCopy, patch); err != nil {
			logger.Error(err, "unable to patch PersistentVolume")
			return err
		}
	}

	namespacedPvCopy := namespacedPv.DeepCopy()
	namespacedPvCopy.Status.RefPvName = ""
	namespacedPvCopy.Status.RefPvUid = ""
	if err := r.Status().Update(ctx, namespacedPvCopy); err != nil {
		logger.Error(err, "unable to update NamespacedPv")
		return err
	}

	controllerutil.RemoveFinalizer(pv, FINALIZER_NAME)
	if err := r.Update(ctx, pv); err != nil {
		logger.Error(err, "unable to update PersistentVolume")
		return err
	}
	logger.Info("pv deleted", "namespace", pv.Namespace, "name", pv.Name)
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PersistentVolumeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.PersistentVolume{}).
		Complete(r)
}
