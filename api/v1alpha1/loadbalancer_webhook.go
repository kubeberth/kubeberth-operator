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
var loadbalancerlog = logf.Log.WithName("loadbalancer-resource")
var loadbalancerClient client.Client

func (r *LoadBalancer) SetupWebhookWithManager(mgr ctrl.Manager) error {
	loadbalancerClient = mgr.GetClient()
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-berth-kubeberth-io-v1alpha1-loadbalancer,mutating=true,failurePolicy=fail,sideEffects=None,groups=berth.kubeberth.io,resources=loadbalancers,verbs=create;update,versions=v1alpha1,name=mloadbalancer.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &LoadBalancer{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *LoadBalancer) Default() {
	loadbalancerlog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-berth-kubeberth-io-v1alpha1-loadbalancer,mutating=false,failurePolicy=fail,sideEffects=None,groups=berth.kubeberth.io,resources=loadbalancers,verbs=create;update;delete,versions=v1alpha1,name=vloadbalancer.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &LoadBalancer{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *LoadBalancer) ValidateCreate() error {
	loadbalancerlog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	var errs field.ErrorList
	ctx := context.Background()

	if r.Spec.Backends == nil {
		return nil
	}

	destinations := []Destination{}
	for _, destination := range r.Spec.Backends {
		server := &Server{}
		nsn := types.NamespacedName{
			Namespace: r.GetNamespace(),
			Name:      destination.Server,
		}
		if err := loadbalancerClient.Get(ctx, nsn, server); err != nil {
			errs = append(errs, field.Invalid(field.NewPath("spec", "backends"), destination.Server, "is not found"))
		}
		destinations = append(destinations, Destination{Server: server.GetName()})
	}

	if len(errs) > 0 {
		err := apierrors.NewInvalid(schema.GroupKind{Group: GroupVersion.Group, Kind: "LoadBalancer"}, r.GetName(), errs)
		return err
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *LoadBalancer) ValidateUpdate(old runtime.Object) error {
	loadbalancerlog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	var errs field.ErrorList
	ctx := context.Background()

	if r.Spec.Backends == nil {
		return nil
	}

	destinations := []Destination{}
	for _, destination := range r.Spec.Backends {
		server := &Server{}
		nsn := types.NamespacedName{
			Namespace: r.GetNamespace(),
			Name:      destination.Server,
		}
		if err := loadbalancerClient.Get(ctx, nsn, server); err != nil {
			errs = append(errs, field.Invalid(field.NewPath("spec", "backends"), destination.Server, "is not found"))
		}
		destinations = append(destinations, Destination{Server: server.GetName()})
	}

	if len(errs) > 0 {
		err := apierrors.NewInvalid(schema.GroupKind{Group: GroupVersion.Group, Kind: "LoadBalancer"}, r.GetName(), errs)
		return err
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *LoadBalancer) ValidateDelete() error {
	loadbalancerlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
