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

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var archivelog = logf.Log.WithName("archive-resource")

func (r *Archive) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-berth-kubeberth-io-v1alpha1-archive,mutating=true,failurePolicy=fail,sideEffects=None,groups=berth.kubeberth.io,resources=archives,verbs=create;update,versions=v1alpha1,name=marchive.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Archive{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Archive) Default() {
	archivelog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-berth-kubeberth-io-v1alpha1-archive,mutating=false,failurePolicy=fail,sideEffects=None,groups=berth.kubeberth.io,resources=archives,verbs=create;update;delete,versions=v1alpha1,name=varchive.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Archive{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Archive) ValidateCreate() error {
	archivelog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	var errs field.ErrorList
	ctx := context.Background()

	if r.Spec.Repository == "" && r.Spec.Source == nil {
		errs = append(errs, field.Invalid(field.NewPath("archive", "Spec"), r.Spec, "must be defined, Repository or Source"))
	}

	if r.Spec.Repository != "" && r.Spec.Source != nil {
		errs = append(errs, field.Invalid(field.NewPath("archive", "Spec"), r.Spec, "conflict between Repository and Source"))
	}

	if r.Spec.Repository != "" && r.Spec.Source == nil {
		return nil
	}

	if r.Spec.Source.Archive != nil && r.Spec.Source.Disk != nil {
		errs = append(errs, field.Invalid(field.NewPath("spec", "source"), r.Spec.Source, "conflict between Archive and Disk"))
	} else if r.Spec.Source.Archive != nil {
		sourceArchive := &Archive{}
		nsn := types.NamespacedName{
			Namespace: r.Namespace,
			Name:      r.Spec.Source.Archive.Name,
		}
		if err := diskClient.Get(ctx, nsn, sourceArchive); err != nil {
			errs = append(errs, field.Invalid(field.NewPath("spec.source", "archive"), r.Spec.Source.Archive.Name, "is not found"))
		}
	} else if r.Spec.Source.Disk != nil {
		sourceDisk := &Disk{}
		nsn := types.NamespacedName{
			Namespace: r.Namespace,
			Name:      r.Spec.Source.Disk.Name,
		}
		if err := diskClient.Get(ctx, nsn, sourceDisk); err != nil {
			errs = append(errs, field.Invalid(field.NewPath("spec.source", "disk"), r.Spec.Source.Disk.Name, "is not found"))
		}

		if sourceDisk.Status.State != "Detached" {
			errs = append(errs, field.Invalid(field.NewPath("spec.source", "disk"), sourceDisk.Status.State, "state must be \"Detached\""))
		}
	}

	if len(errs) > 0 {
		err := apierrors.NewInvalid(schema.GroupKind{Group: GroupVersion.Group, Kind: "Archive"}, r.Name, errs)
		return err
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Archive) ValidateUpdate(old runtime.Object) error {
	archivelog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Archive) ValidateDelete() error {
	archivelog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	var errs field.ErrorList

	if r.Status.State != "Created" {
		errs = append(errs, field.Invalid(field.NewPath("status", "state"), r.Status.State, "state must be \"Created\""))
	}

	if len(errs) > 0 {
		err := apierrors.NewInvalid(schema.GroupKind{Group: GroupVersion.Group, Kind: "Archive"}, r.Name, errs)
		return err
	}

	return nil
}
