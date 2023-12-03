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
	"errors"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	namespacedpvv1 "github.com/homirun/namespaced-pv-controller/api/v1"
)

// NamespacedPvReconciler reconciles a NamespacedPv object
type NamespacedPvReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=namespaced-pv.homi.run,resources=namespacedpvs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=namespaced-pv.homi.run,resources=namespacedpvs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=namespaced-pv.homi.run,resources=namespacedpvs/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=persistentvolumes,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NamespacedPv object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *NamespacedPvReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	var namespacedPv namespacedpvv1.NamespacedPv
	if err := r.Get(ctx, req.NamespacedName, &namespacedPv); err != nil {
		logger.Error(err, "unable to fetch NamespacedPv")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	err := r.CreateOrUpdatePv(ctx, &namespacedPv)
	if err != nil {
		logger.Error(err, "unable to create PersistentVolume")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NamespacedPvReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&namespacedpvv1.NamespacedPv{}).
		Complete(r)
}

// create PersietentVolume
func (r *NamespacedPvReconciler) CreateOrUpdatePv(ctx context.Context, namespacedPv *namespacedpvv1.NamespacedPv) error {
	// only nfs supported
	logger := log.FromContext(ctx)
	if &namespacedPv.Spec.Nfs == nil {
		nilNfsError := errors.New("only nfs supported")
		return nilNfsError
	}

	// create PersistentVolume
	pv := &corev1.PersistentVolume{}

	pv.SetName(namespacedPv.Spec.VolumeName + "-" + namespacedPv.Namespace)
	op, err := ctrl.CreateOrUpdate(ctx, r.Client, pv, func() error {
		pv.Spec = corev1.PersistentVolumeSpec{
			AccessModes: namespacedPv.Spec.AccessModes,
			Capacity:    namespacedPv.Spec.Capacity,
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				NFS: &corev1.NFSVolumeSource{
					Server:   namespacedPv.Spec.Nfs.Server,
					Path:     namespacedPv.Spec.Nfs.Path,
					ReadOnly: namespacedPv.Spec.Nfs.ReadOnly,
				},
			},
			PersistentVolumeReclaimPolicy: namespacedPv.Spec.ReclaimPolicy,
			StorageClassName:              namespacedPv.Spec.StorageClassName,
			VolumeMode:                    &namespacedPv.Spec.VolumeMode,
		}
		return nil
	})

	if err != nil {
		return err
	}

	if op != controllerutil.OperationResultNone {
		logger.Info("PersistentVolume created or updated", "operation", op)
	}

	return nil
}
