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
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var serverlog = logf.Log.WithName("server-resource")
var c client.Client

func (r *Server) SetupWebhookWithManager(mgr ctrl.Manager) error {
	c = mgr.GetClient()
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
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-berth-kubeberth-io-v1alpha1-server,mutating=false,failurePolicy=fail,sideEffects=None,groups=berth.kubeberth.io,resources=servers,verbs=create;update,versions=v1alpha1,name=vserver.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Server{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Server) ValidateCreate() error {
	serverlog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	var errs field.ErrorList
	ctx := context.Background()

	if r.Spec.Disk != nil {
		diskNsN := types.NamespacedName{
			Namespace: r.Namespace,
			Name:      r.Spec.Disk.Name,
		}

		// Get the Disk.
		disk := &Disk{}
		if err := c.Get(ctx, diskNsN, disk); err != nil {
			serverlog.Info("could not get disk", "name", r.Name)
			errs = append(errs, field.Invalid(field.NewPath("spec", "disk"), r.Spec.Disk.Name, "is not found"))
		}
	}

	if r.Spec.CloudInit != nil {
		cloudinitNsN := types.NamespacedName{
			Namespace: r.Namespace,
			Name:      r.Spec.CloudInit.Name,
		}

		// Get the CloudInit.
		cloudinit := &CloudInit{}
		if err := c.Get(ctx, cloudinitNsN, cloudinit); err != nil {
			serverlog.Info("could not get cloudinit", "name", r.Name)
			errs = append(errs, field.Invalid(field.NewPath("spec", "cloudinit"), r.Spec.CloudInit.Name, "is not found"))
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
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Server) ValidateDelete() error {
	serverlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
