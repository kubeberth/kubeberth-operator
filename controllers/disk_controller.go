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
	"time"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"

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

const (
	diskRequeueAfter = time.Second * 3
)

//+kubebuilder:rbac:groups=berth.kubeberth.io,resources=disks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=berth.kubeberth.io,resources=disks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=berth.kubeberth.io,resources=disks/finalizers,verbs=update
//+kubebuilder:rbac:groups=berth.kubeberth.io,resources=archives,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cdi.kubevirt.io,resources=datavolumes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete

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
		log.Error(err, "cloud not get the Disk resource")
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{Requeue: false}, nil
		}
		return ctrl.Result{Requeue: true}, err
	}

	if err := r.ensureDataVolumeExists(ctx, disk); err != nil {
		log.Error(err, "failed to do ensureDataVolumeExists")
		return ctrl.Result{Requeue: true}, err
	}

	if ensuring, err := r.ensureDiskExists(ctx, disk); err != nil {
		log.Error(err, "failed to do ensureDiskExists")
		return ctrl.Result{Requeue: true}, err
	} else if ensuring {
		return ctrl.Result{Requeue: true, RequeueAfter: diskRequeueAfter}, nil
	}

	if expanding, err := r.isDiskExpanding(ctx, disk); err != nil {
		log.Error(err, "failed to do ensureDiskExpanding")
		return ctrl.Result{Requeue: true}, err
	} else if expanding {
		return ctrl.Result{Requeue: true, RequeueAfter: diskRequeueAfter}, nil
	}

	return ctrl.Result{Requeue: false}, nil
}

func (r *DiskReconciler) isDiskExpanding(ctx context.Context, disk *berthv1alpha1.Disk) (bool, error) {
	log := r.Log.WithValues("ensureDiskExpanding", disk.GetName())

	specSize := resource.MustParse(disk.Spec.Size)
	statusSize := resource.MustParse(disk.Status.Size)

	if (&specSize).Cmp(statusSize) > 0 {
		pvc := &corev1.PersistentVolumeClaim{}
		pvc.SetNamespace(disk.GetNamespace())
		pvc.SetName(disk.GetName())
		if _, err := ctrl.CreateOrUpdate(ctx, r.Client, pvc, func() error {
			pvc.Spec.Resources = corev1.ResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceStorage: specSize},
			}
			return nil
		}); err != nil {
			log.Error(err, "unable to update the PVC")
			return true, err
		}

		disk.Status.Size = disk.Spec.Size
		disk.Status.State = "Resizing"
		disk.Status.Phase = "Expanding"
		if err := r.Status().Update(ctx, disk); err != nil {
			log.Error(err, "unable to update a status of the Disk")
			return true, err
		}

		return true, nil
	} else if disk.Status.State == "Resizing" {
		pvc := &corev1.PersistentVolumeClaim{}
		pvcNsN := types.NamespacedName{
			Namespace: disk.GetNamespace(),
			Name:      disk.GetName(),
		}
		if err := r.Get(ctx, pvcNsN, pvc); err != nil {
			return true, err
		}

		pvcSize := pvc.Status.Capacity[corev1.ResourceStorage]
		diskSize := resource.MustParse(disk.Status.Size)

		if (&diskSize).Cmp(pvcSize) != 0 {
			return true, nil
		}

		disk.Status.State = "Detached"
		disk.Status.Phase = "Created"
		if err := r.Status().Update(ctx, disk); err != nil {
			log.Error(err, "unable to update a status of the Disk")
			return true, err
		}
		return false, nil
	} else {
		return false, nil
	}
}

func (r *DiskReconciler) ensureDiskExists(ctx context.Context, disk *berthv1alpha1.Disk) (ensuring bool, err error) {
	log := r.Log.WithValues("ensureDiskExists", disk.GetName())

	if disk.Status.State == "Resizing" || disk.Status.Phase == "Created" {
		return false, nil
	}

	datavolume := &cdiv1.DataVolume{}
	nsn := types.NamespacedName{
		Namespace: disk.GetNamespace(),
		Name:      disk.GetName(),
	}
	if err := r.Get(ctx, nsn, datavolume); err != nil {
		return true, err
	}

	disk.Status.State = "Provisioning"
	switch datavolume.Status.Phase {
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
		disk.Status.State = "Detached"
	case cdiv1.Failed:
		disk.Status.Phase = "Failed"
	case cdiv1.Unknown:
		disk.Status.Phase = "Unknown"
	case cdiv1.Paused:
		disk.Status.Phase = "Paused"
	}

	disk.Status.Progress = string(datavolume.Status.Progress)
	if err := r.Status().Update(ctx, disk); err != nil {
		log.Error(err, "unable to update a status of the Disk")
		return true, err
	}

	return true, nil
}

func (r *DiskReconciler) ensureDataVolumeExists(ctx context.Context, disk *berthv1alpha1.Disk) error {
	log := r.Log.WithValues("ensureDataVolumeExists", disk.GetName())

	datavolume := &cdiv1.DataVolume{}
	nsn := types.NamespacedName{
		Namespace: disk.GetNamespace(),
		Name:      disk.GetName(),
	}
	if err := r.Get(ctx, nsn, datavolume); err == nil {
		return nil
	}

	if err := r.createDataVolume(ctx, disk); err != nil {
		return err
	}

	disk.Status.Size = disk.Spec.Size
	disk.Status.State = "Provisioning"
	if err := r.Status().Update(ctx, disk); err != nil {
		log.Error(err, "unable to update a status of the Disk")
		return err
	}

	return nil
}

func (r *DiskReconciler) createDataVolume(ctx context.Context, disk *berthv1alpha1.Disk) error {
	log := r.Log.WithValues("createDataVolume", disk.GetName())

	datavolumeSource, err := r.createDataVolumeSource(ctx, disk)
	if err != nil {
		return err
	}

	kubeberth := &berthv1alpha1.KubeBerth{}
	kubeberthNsN := types.NamespacedName{
		Namespace: "kubeberth",
		Name:      "kubeberth",
	}
	if err := r.Get(ctx, kubeberthNsN, kubeberth); err != nil {
		return err
	}

	storageClassName := kubeberth.Spec.StorageClassName
	volumeMode := corev1.PersistentVolumeBlock
	resourceRequest := corev1.ResourceList{}
	resourceRequest[corev1.ResourceStorage] = resource.MustParse(disk.Spec.Size)
	datavolume := &cdiv1.DataVolume{}
	datavolume.SetNamespace(disk.GetNamespace())
	datavolume.SetName(disk.GetName())
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, datavolume, func() error {
		spec := cdiv1.DataVolumeSpec{
			Source: datavolumeSource,
			PVC: &corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse(disk.Spec.Size)},
				},
				StorageClassName: &storageClassName,
				VolumeMode:       &volumeMode,
			},
		}
		datavolume.Spec = spec
		if err := ctrl.SetControllerReference(disk, datavolume, r.Scheme); err != nil {
			log.Error(err, "unable to set controlerReference from the Disk to the DataVolume")
			return err
		}

		return nil
	}); err != nil {
		log.Error(err, "unable to ensure the DataVolume is correct")
		return err
	}

	return nil
}

func (r *DiskReconciler) createDataVolumeSource(ctx context.Context, disk *berthv1alpha1.Disk) (*cdiv1.DataVolumeSource, error) {
	datavolumeSource := &cdiv1.DataVolumeSource{}

	if disk.Spec.Source == nil {
		datavolumeSource.Blank = &cdiv1.DataVolumeBlankImage{}
	} else if disk.Spec.Source.Archive != nil {
		sourceArchive := &berthv1alpha1.Archive{}
		nsn := types.NamespacedName{
			Namespace: disk.GetNamespace(),
			Name:      disk.Spec.Source.Archive.Name,
		}
		if err := r.Get(ctx, nsn, sourceArchive); err != nil {
			return nil, err
		}

		datavolumeSource.HTTP = &cdiv1.DataVolumeSourceHTTP{
			URL: sourceArchive.Spec.Repository,
		}

	} else if disk.Spec.Source.Disk != nil {
		sourceDisk := &berthv1alpha1.Disk{}
		nsn := types.NamespacedName{
			Namespace: disk.GetNamespace(),
			Name:      disk.Spec.Source.Disk.Name,
		}
		if err := r.Get(ctx, nsn, sourceDisk); err != nil {
			return nil, err
		}

		datavolumeSource.PVC = &cdiv1.DataVolumeSourcePVC{
			Namespace: disk.GetNamespace(),
			Name:      disk.Spec.Source.Disk.Name,
		}
	} else {
		datavolumeSource.Blank = &cdiv1.DataVolumeBlankImage{}
	}

	return datavolumeSource, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DiskReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&berthv1alpha1.Disk{}).
		Complete(r)
}
