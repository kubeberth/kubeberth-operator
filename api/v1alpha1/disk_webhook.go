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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var disklog = logf.Log.WithName("disk-resource")

func (r *Disk) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-berth-kubeberth-io-v1alpha1-disk,mutating=true,failurePolicy=fail,sideEffects=None,groups=berth.kubeberth.io,resources=disks,verbs=create;update,versions=v1alpha1,name=mdisk.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Disk{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Disk) Default() {
	disklog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-berth-kubeberth-io-v1alpha1-disk,mutating=false,failurePolicy=fail,sideEffects=None,groups=berth.kubeberth.io,resources=disks,verbs=create;update;delete,versions=v1alpha1,name=vdisk.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Disk{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Disk) ValidateCreate() error {
	disklog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Disk) ValidateUpdate(old runtime.Object) error {
	disklog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	var errs field.ErrorList
	specSize := resource.MustParse(r.Spec.Size)
	statusSize := resource.MustParse(r.Status.Size)

	if (&statusSize).Cmp(specSize) > 0 {
		errs = append(errs, field.Invalid(field.NewPath("state", "size"), r.Spec.Size, "must be more than "+r.Status.Size))
	}

	if r.Status.State != "Detached" {
		errs = append(errs, field.Invalid(field.NewPath("status", "state"), r.Status.State, "state must be \"Detached\""))
	}

	if len(errs) > 0 {
		err := apierrors.NewInvalid(schema.GroupKind{Group: GroupVersion.Group, Kind: "Disk"}, r.Name, errs)
		return err
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Disk) ValidateDelete() error {
	disklog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	var errs field.ErrorList

	if r.Status.State == "Attached" {
		errs = append(errs, field.Invalid(field.NewPath("status", "state"), r.Status.State, "state must be \"Detached\""))
	}

	if len(errs) > 0 {
		err := apierrors.NewInvalid(schema.GroupKind{Group: GroupVersion.Group, Kind: "Disk"}, r.Name, errs)
		return err
	}

	return nil
}
