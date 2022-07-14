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

const (
	isoimageRequeueAfter = time.Second * 1
)

// ISOImageReconciler reconciles a ISOImage object
type ISOImageReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=berth.kubeberth.io,resources=isoimages,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=berth.kubeberth.io,resources=isoimages/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=berth.kubeberth.io,resources=isoimages/finalizers,verbs=update
//+kubebuilder:rbac:groups=cdi.kubevirt.io,resources=datavolumes,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ISOImage object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *ISOImageReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("ISOImage", req.NamespacedName)

	isoimage := &berthv1alpha1.ISOImage{}
	if err := r.Get(ctx, req.NamespacedName, isoimage); err != nil {
		log.Error(err, "could not get the ISOImage resource")
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{Requeue: false}, nil
		}
		return ctrl.Result{Requeue: true}, err
	}

	if err := r.ensureDataVolumeExists(ctx, isoimage); err != nil {
		log.Error(err, "failed to do ensureDataVolumeExists")
		return ctrl.Result{Requeue: true}, err
	}

	if ensuring, err := r.ensureISOImageExists(ctx, isoimage); err != nil {
		log.Error(err, "failed to do ensureISOImageExists")
		return ctrl.Result{Requeue: true}, err
	} else if ensuring {
		return ctrl.Result{Requeue: true, RequeueAfter: isoimageRequeueAfter}, nil
	}

	return ctrl.Result{Requeue: false}, nil
}

func (r *ISOImageReconciler) ensureISOImageExists(ctx context.Context, isoimage *berthv1alpha1.ISOImage) (ensuring bool, err error) {
	log := r.Log.WithValues("ensureISOImageExists", isoimage.GetName())

	if isoimage.Status.State == "Created" {
		return false, nil
	}

	datavolume := &cdiv1.DataVolume{}
	nsn := types.NamespacedName{
		Namespace: isoimage.GetNamespace(),
		Name:      isoimage.GetName(),
	}
	if err := r.Get(ctx, nsn, datavolume); err != nil {
		return true, err
	}
	isoimage.Status.Progress = string(datavolume.Status.Progress)

	isoimage.Status.State = "Provisioning"
	switch datavolume.Status.Phase {
	case cdiv1.Pending:
		isoimage.Status.State = "Pending"
	case cdiv1.PVCBound:
		isoimage.Status.State = "PVCBound"
	case cdiv1.ImportScheduled:
		isoimage.Status.State = "Scheduled"
	case cdiv1.ImportInProgress, cdiv1.CloneInProgress:
		isoimage.Status.State = "Creating"
	case cdiv1.ExpansionInProgress:
		isoimage.Status.State = "Expanding"
	case cdiv1.Succeeded:
		isoimage.Status.State = "Created"
	case cdiv1.Failed:
		isoimage.Status.State = "Failed"
	case cdiv1.Unknown:
		isoimage.Status.State = "Unknown"
	case cdiv1.Paused:
		isoimage.Status.State = "Paused"
	}

	isoimage.Status.Size = isoimage.Spec.Size
	isoimage.Status.Repository = isoimage.Spec.Repository
	if err := r.Status().Update(ctx, isoimage); err != nil {
		log.Error(err, "unable to update a status of the Disk")
		return true, err
	}

	return true, nil
}

func (r *ISOImageReconciler) ensureDataVolumeExists(ctx context.Context, isoimage *berthv1alpha1.ISOImage) error {
	log := r.Log.WithValues("ensureDataVolumeExists", isoimage.GetName())

	datavolume := &cdiv1.DataVolume{}
	nsn := types.NamespacedName{
		Namespace: isoimage.GetNamespace(),
		Name:      isoimage.GetName(),
	}
	if err := r.Get(ctx, nsn, datavolume); err == nil {
		return nil
	}

	if err := r.createDataVolume(ctx, isoimage); err != nil {
		log.Error(err, "unable to create a DataVolume")
		return err
	}

	return nil
}

func (r *ISOImageReconciler) createDataVolume(ctx context.Context, isoimage *berthv1alpha1.ISOImage) error {
	log := r.Log.WithValues("createDataVolume", isoimage.GetName())

	datavolumeSource := &cdiv1.DataVolumeSource{}
	datavolumeSource.HTTP = &cdiv1.DataVolumeSourceHTTP{
		URL: isoimage.Spec.Repository,
	}

	kubeberth := &berthv1alpha1.KubeBerth{}
	kubeberthNsN := types.NamespacedName{
		Namespace: "kubeberth",
		Name:      "kubeberth",
	}
	if err := r.Get(ctx, kubeberthNsN, kubeberth); err != nil {
		return err
	}

	accessModes := []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
	resourceRequest := corev1.ResourceList{corev1.ResourceStorage: resource.MustParse(isoimage.Spec.Size)}
	storageClassName := kubeberth.Spec.StorageClassName
	volumeMode := kubeberth.Spec.VolumeMode
	datavolume := &cdiv1.DataVolume{}
	datavolume.SetNamespace(isoimage.GetNamespace())
	datavolume.SetName(isoimage.GetName())
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, datavolume, func() error {
		spec := cdiv1.DataVolumeSpec{
			Source: datavolumeSource,
			PVC: &corev1.PersistentVolumeClaimSpec{
				AccessModes: accessModes,
				Resources: corev1.ResourceRequirements{
					Requests: resourceRequest,
				},
				StorageClassName: &storageClassName,
				VolumeMode:       volumeMode,
			},
		}
		datavolume.Spec = spec
		if err := ctrl.SetControllerReference(isoimage, datavolume, r.Scheme); err != nil {
			log.Error(err, "unable to set controlerReference from the ISOImage to the DataVolume")
			return err
		}

		return nil
	}); err != nil {
		log.Error(err, "unable to ensure the DataVolume is correct")
		return err
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ISOImageReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&berthv1alpha1.ISOImage{}).
		Complete(r)
}
