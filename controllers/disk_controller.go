/*
Copyright 2022.

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

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cdiv1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"

	berthv1alpha1 "github.com/kubeberth/kubeberth-operator/api/v1alpha1"
)

// DiskReconciler reconciles a Disk object
type DiskReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=berth.kubeberth.io,resources=disks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=berth.kubeberth.io,resources=disks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=berth.kubeberth.io,resources=disks/finalizers,verbs=update
//+kubebuilder:rbac:groups=berth.kubeberth.io,resources=archives,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cdi.kubevirt.io,resources=datavolumes,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Disk object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *DiskReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("Disk", req.NamespacedName)

	// Get the Disk.
	disk := &berthv1alpha1.Disk{}
	if err := r.Get(ctx, req.NamespacedName, disk); err != nil {
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if disk.Status.Phase == "Created" {
		return ctrl.Result{}, nil
	}

	if disk.Status.Phase != "" {
		// Get the DataVolume
		createdDV := &cdiv1.DataVolume{}
		if err := r.Get(ctx, req.NamespacedName, createdDV); err != nil {
			if k8serrors.IsNotFound(err) {
				return ctrl.Result{}, nil
			}
			return ctrl.Result{}, err
		}

		disk.Status.Progress = string(createdDV.Status.Progress)

		switch createdDV.Status.Phase {
		case cdiv1.Pending:
			disk.Status.Phase = "Pending"
		case cdiv1.PVCBound:
			disk.Status.Phase = "PVCBound"
		case cdiv1.ImportScheduled:
			disk.Status.Phase = "Scheduled"
		case cdiv1.ImportInProgress, cdiv1.CloneInProgress:
			disk.Status.Phase = "Creating"
		case cdiv1.ExpansionInProgress:
			disk.Status.Phase = "Expanding"
		case cdiv1.Succeeded:
			disk.Status.Phase = "Created"
			//disk.Status.State = "Unattached"
		case cdiv1.Failed:
			disk.Status.Phase = "Failed"
		case cdiv1.Unknown:
			disk.Status.Phase = "Unknown"
		case cdiv1.Paused:
			disk.Status.Phase = "Paused"
		}

		if err := r.Status().Update(ctx, disk); err != nil {
			log.Error(err, "unable to update Disk status")
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	datavolumeSource := &cdiv1.DataVolumeSource{}

	if disk.Spec.Source.Archive != nil {
		nsn := types.NamespacedName{
			Namespace: disk.Spec.Source.Archive.Namespace,
			Name:      disk.Spec.Source.Archive.Name,
		}
		// Get the Archive.
		createdArchive := &berthv1alpha1.Archive{}
		if err := r.Get(ctx, nsn, createdArchive); err != nil {
			disk.Status.Size = disk.Spec.Size
			disk.Status.Phase = "Failed"

			if err := r.Status().Update(ctx, disk); err != nil {
				log.Error(err, "unable to update Disk status")
				return ctrl.Result{}, err
			}

			if k8serrors.IsNotFound(err) {
				return ctrl.Result{}, nil
			}
			return ctrl.Result{}, err
		}

		datavolumeSource.HTTP = &cdiv1.DataVolumeSourceHTTP{
			URL: createdArchive.Spec.URL,
		}

	} else if disk.Spec.Source.Disk != nil {
		nsn := types.NamespacedName{
			Namespace: disk.Spec.Source.Disk.Namespace,
			Name:      disk.Spec.Source.Disk.Name,
		}
		// Get the source Disk
		sourceDisk := &berthv1alpha1.Disk{}
		if err := r.Get(ctx, nsn, sourceDisk); err != nil {
			if k8serrors.IsNotFound(err) {
				return ctrl.Result{}, nil
			}
			return ctrl.Result{}, err
		}

		datavolumeSource.PVC = &cdiv1.DataVolumeSourcePVC{
			Namespace: disk.Spec.Source.Disk.Namespace,
			Name:      disk.Spec.Source.Disk.Name,
		}

	} else {
		log.Info("no SourceArchive")
		datavolumeSource.Blank = &cdiv1.DataVolumeBlankImage{}
	}

	kubeberthNsN := types.NamespacedName{
		Namespace: "kubeberth",
		Name:      "kubeberth",
	}

	// Get the KubeBerth.
	kubeberth := &berthv1alpha1.KubeBerth{}
	if err := r.Get(ctx, kubeberthNsN, kubeberth); err != nil {
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	storageClassName := kubeberth.Spec.StorageClassName
	volumeMode := corev1.PersistentVolumeBlock

	resourceRequest := corev1.ResourceList{}
	resourceRequest[corev1.ResourceStorage] = resource.MustParse(disk.Spec.Size)

	// Create the DataVolume.
	datavolume := &cdiv1.DataVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:      disk.Name,
			Namespace: disk.Namespace,
		},
		Spec: cdiv1.DataVolumeSpec{
			Source: datavolumeSource,
			PVC: &corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse(disk.Spec.Size)},
				},
				StorageClassName: &storageClassName,
				VolumeMode:       &volumeMode,
			},
		},
	}

	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, datavolume, func() error {
		if err := ctrl.SetControllerReference(disk, datavolume, r.Scheme); err != nil {
			log.Error(err, "unable to set controlerReference from Disk to DataVolume")
			return err
		}
		return nil
	}); err != nil {
		// error handling of ctrl.CreateOrUpdate
		log.Error(err, "unable to ensure DataVolume is correct")
		return ctrl.Result{}, err
	}

	disk.Status.Size = disk.Spec.Size
	//disk.Status.State = "Preparing"
	disk.Status.Phase = "Creating"

	if err := r.Status().Update(ctx, disk); err != nil {
		log.Error(err, "unable to update Disk status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DiskReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&berthv1alpha1.Disk{}).
		Complete(r)
}
