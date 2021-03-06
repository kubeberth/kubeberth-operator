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

package v1alpha1

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var kubeberthlog = logf.Log.WithName("kubeberth-resource")
var kubeberthClient client.Client

func (r *KubeBerth) SetupWebhookWithManager(mgr ctrl.Manager) error {
	kubeberthClient = mgr.GetClient()
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-berth-kubeberth-io-v1alpha1-kubeberth,mutating=true,failurePolicy=fail,sideEffects=None,groups=berth.kubeberth.io,resources=kubeberths,verbs=create;update,versions=v1alpha1,name=mkubeberth.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &KubeBerth{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *KubeBerth) Default() {
	kubeberthlog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
	if r.Spec.VolumeMode == nil {
		volumeMode := corev1.PersistentVolumeFilesystem
		r.Spec.VolumeMode = &volumeMode
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-berth-kubeberth-io-v1alpha1-kubeberth,mutating=false,failurePolicy=fail,sideEffects=None,groups=berth.kubeberth.io,resources=kubeberths,verbs=create;update,versions=v1alpha1,name=vkubeberth.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &KubeBerth{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *KubeBerth) ValidateCreate() error {
	kubeberthlog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	var errs field.ErrorList
	ctx := context.Background()

	if !(*r.Spec.VolumeMode == corev1.PersistentVolumeBlock || *r.Spec.VolumeMode == corev1.PersistentVolumeFilesystem) {
		errs = append(errs, field.Invalid(field.NewPath("Spec", "VolumeMode"), *r.Spec.VolumeMode, "must be either Block or Filesystem"))
	}

	storageClasses := &storagev1.StorageClassList{}
	if err := kubeberthClient.List(ctx, storageClasses, &client.ListOptions{}); err != nil {
		errs = append(errs, field.Invalid(field.NewPath("Spec", "StorageClassName"), r.Spec.StorageClassName, "is not found"))
	} else {
		notFound := true
		for _, storageClass := range storageClasses.Items {
			if r.Spec.StorageClassName == storageClass.GetName() {
				notFound = false
			}
		}
		if notFound {
			errs = append(errs, field.Invalid(field.NewPath("Spec", "StorageClassName"), r.Spec.StorageClassName, "is not found"))
		}
	}

	if len(errs) > 0 {
		err := apierrors.NewInvalid(schema.GroupKind{Group: GroupVersion.Group, Kind: "KubeBerth"}, r.Name, errs)
		return err
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *KubeBerth) ValidateUpdate(old runtime.Object) error {
	kubeberthlog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	var errs field.ErrorList
	ctx := context.Background()

	if !(*r.Spec.VolumeMode == corev1.PersistentVolumeBlock || *r.Spec.VolumeMode == corev1.PersistentVolumeFilesystem) {
		errs = append(errs, field.Invalid(field.NewPath("Spec", "VolumeMode"), r.Spec.StorageClassName, "must be either Block or Filesystem"))
	}

	storageClasses := &storagev1.StorageClassList{}
	if err := kubeberthClient.List(ctx, storageClasses, &client.ListOptions{}); err != nil {
		errs = append(errs, field.Invalid(field.NewPath("Spec", "StorageClassName"), r.Spec.StorageClassName, "is not found"))
	} else {
		notFound := true
		for _, storageClass := range storageClasses.Items {
			if r.Spec.StorageClassName == storageClass.GetName() {
				notFound = false
			}
		}
		if notFound {
			errs = append(errs, field.Invalid(field.NewPath("Spec", "StorageClassName"), r.Spec.StorageClassName, "is not found"))
		}
	}

	if len(errs) > 0 {
		err := apierrors.NewInvalid(schema.GroupKind{Group: GroupVersion.Group, Kind: "KubeBerth"}, r.Name, errs)
		return err
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *KubeBerth) ValidateDelete() error {
	kubeberthlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
