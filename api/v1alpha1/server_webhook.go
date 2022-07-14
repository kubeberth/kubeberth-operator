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
	"reflect"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var serverlog = logf.Log.WithName("server-resource")
var serverClient client.Client

func (r *Server) SetupWebhookWithManager(mgr ctrl.Manager) error {
	serverClient = mgr.GetClient()
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-berth-kubeberth-io-v1alpha1-server,mutating=true,failurePolicy=fail,sideEffects=None,groups=berth.kubeberth.io,resources=servers,verbs=create;update,versions=v1alpha1,name=mserver.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Server{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Server) Default() {
	serverlog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
	if r.Spec.Running == nil {
		running := false
		r.Spec.Running = &running
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-berth-kubeberth-io-v1alpha1-server,mutating=false,failurePolicy=fail,sideEffects=None,groups=berth.kubeberth.io,resources=servers,verbs=create;update;delete,versions=v1alpha1,name=vserver.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Server{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Server) ValidateCreate() error {
	serverlog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	var errs field.ErrorList
	ctx := context.Background()

	if len(r.Spec.Disks) != 0 {
		for _, disk := range r.Spec.Disks {
			validatingDisk := &Disk{}
			nsn := types.NamespacedName{
				Namespace: r.Namespace,
				Name:      disk.Name,
			}
			if err := serverClient.Get(ctx, nsn, validatingDisk); err != nil {
				serverlog.Info("could not get disk", "name", disk.Name)
				errs = append(errs, field.Invalid(field.NewPath("Spec", "Disks"), disk.Name, "is not found"))
			} else {
				if validatingDisk.Status.State != "Detached" {
					errs = append(errs, field.Invalid(field.NewPath("Spec", "Disks"), validatingDisk.Status.State, "state must be \"Detached\""))
				}
			}
		}
	}

	if r.Spec.ISOImage != nil {
		isoimage := &ISOImage{}
		nsn := types.NamespacedName{
			Namespace: r.Namespace,
			Name:      r.Spec.ISOImage.Name,
		}
		if err := serverClient.Get(ctx, nsn, isoimage); err != nil {
			serverlog.Info("could not get isoimage", "name", r.Spec.ISOImage.Name)
			errs = append(errs, field.Invalid(field.NewPath("Spec", "ISOImage"), r.Spec.ISOImage.Name, "is not found"))
		}
	}

	if r.Spec.CloudInit != nil {
		cloudinit := &CloudInit{}
		nsn := types.NamespacedName{
			Namespace: r.Namespace,
			Name:      r.Spec.CloudInit.Name,
		}
		if err := serverClient.Get(ctx, nsn, cloudinit); err != nil {
			serverlog.Info("could not get cloudinit", "name", r.Spec.CloudInit.Name)
			errs = append(errs, field.Invalid(field.NewPath("Spec", "CloudInit"), r.Spec.CloudInit.Name, "is not found"))
		}
	}

	if len(errs) > 0 {
		err := apierrors.NewInvalid(schema.GroupKind{Group: GroupVersion.Group, Kind: "Server"}, r.Name, errs)
		return err
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Server) ValidateUpdate(old runtime.Object) error {
	serverlog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	var errs field.ErrorList
	ctx := context.Background()

	if r.Status.State == "Running" && *r.Spec.Running {
		errs = append(errs, field.Invalid(field.NewPath("spec", "running"), r.Status.State, "cloud not update a status of the state when the state is \"Running\" state"))
	}

	if (len(r.Spec.Disks) != 0) && !reflect.DeepEqual(r.Status.AttachedDisks, r.Spec.Disks) {
		for _, disk := range r.Spec.Disks {
			validatingDisk := &Disk{}
			nsn := types.NamespacedName{
				Namespace: r.Namespace,
				Name:      disk.Name,
			}
			if err := serverClient.Get(ctx, nsn, validatingDisk); err != nil {
				serverlog.Info("could not get disk", "name", disk.Name)
				errs = append(errs, field.Invalid(field.NewPath("Spec", "Disks"), disk.Name, "is not found"))
			} else {
				if validatingDisk.Status.State != "Detached" && validatingDisk.Status.AttachedTo != r.Name {
					errs = append(errs, field.Invalid(field.NewPath("Spec", "Disks"), validatingDisk.Status.State, "state must be \"Detached\""))
				}
			}
		}
	}

	if r.Spec.ISOImage != nil {
		isoimage := &ISOImage{}
		nsn := types.NamespacedName{
			Namespace: r.Namespace,
			Name:      r.Spec.ISOImage.Name,
		}
		if err := serverClient.Get(ctx, nsn, isoimage); err != nil {
			serverlog.Info("could not get isoimage", "name", r.Spec.ISOImage.Name)
			errs = append(errs, field.Invalid(field.NewPath("Spec", "ISOImage"), r.Spec.ISOImage.Name, "is not found"))
		}
	}

	if r.Spec.CloudInit != nil {
		cloudinit := &CloudInit{}
		nsn := types.NamespacedName{
			Namespace: r.Namespace,
			Name:      r.Spec.CloudInit.Name,
		}
		if err := serverClient.Get(ctx, nsn, cloudinit); err != nil {
			serverlog.Info("could not get cloudinit", "name", r.Spec.CloudInit.Name)
			errs = append(errs, field.Invalid(field.NewPath("Spec", "CloudInit"), r.Spec.CloudInit.Name, "is not found"))
		}
	}

	if len(errs) > 0 {
		err := apierrors.NewInvalid(schema.GroupKind{Group: GroupVersion.Group, Kind: "Server"}, r.Name, errs)
		return err
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Server) ValidateDelete() error {
	serverlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	var errs field.ErrorList

	if r.Status.State != "Stopped" {
		errs = append(errs, field.Invalid(field.NewPath("spec", "running"), r.Status.State, "cloud not delete the Server when a status of the state is \"Stopped\""))
	}

	if len(errs) > 0 {
		err := apierrors.NewInvalid(schema.GroupKind{Group: GroupVersion.Group, Kind: "Server"}, r.Name, errs)
		return err
	}

	return nil
}
