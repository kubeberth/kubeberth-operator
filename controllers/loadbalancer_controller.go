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

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	berthv1alpha1 "github.com/kubeberth/kubeberth-operator/api/v1alpha1"
)

// LoadBalancerReconciler reconciles a LoadBalancer object
type LoadBalancerReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=berth.kubeberth.io,resources=loadbalancers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=berth.kubeberth.io,resources=loadbalancers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=berth.kubeberth.io,resources=loadbalancers/finalizers,verbs=update
//+kubebuilder:rbac:groups=berth.kubeberth.io,resources=servers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the LoadBalancer object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *LoadBalancerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("Server", req.NamespacedName)

	// TODO(user): your logic here

	// Get the LoadBalancer.
	loadbalancer := &berthv1alpha1.LoadBalancer{}
	if err := r.Get(ctx, req.NamespacedName, loadbalancer); err != nil {
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if loadbalancer.Status.State != "Created" {
		for i := 0; i < len(loadbalancer.Spec.Servers); i++ {
			toServer := &berthv1alpha1.Server{}
			toServerNsN := types.NamespacedName{
				Namespace: "kubeberth",
				Name:      loadbalancer.Spec.Servers[i].Server,
			}
			if err := r.Get(ctx, toServerNsN, toServer); err != nil {
				if k8serrors.IsNotFound(err) {
					return ctrl.Result{}, nil
				}
				return ctrl.Result{}, err
			}
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
		externalDNSDomainName := kubeberth.Spec.ExternalDNSDomainName
		annotations := map[string]string{
			"external-dns.alpha.kubernetes.io/hostname": loadbalancer.Name + ".lb." + externalDNSDomainName,
		}
		service := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:        loadbalancer.Name,
				Namespace:   loadbalancer.Namespace,
				Annotations: annotations,
			},
		}
		servicePort := loadbalancer.Spec.Ports
		lbLabels := map[string]string{
			"berth.kubeberth.io/loadbalancer": loadbalancer.Name,
		}
		service.Spec.Type = corev1.ServiceTypeLoadBalancer
		service.Spec.ExternalTrafficPolicy = corev1.ServiceExternalTrafficPolicyTypeLocal
		service.Spec.Ports = servicePort
		service.Spec.Selector = lbLabels

		if _, err := ctrl.CreateOrUpdate(ctx, r.Client, service, func() error {
			if err := ctrl.SetControllerReference(loadbalancer, service, r.Scheme); err != nil {
				log.Error(err, "unable to set controllerReference from LoadBalancer to Service")
				return err
			}
			return nil
		}); err != nil {
			// error handling of ctrl.CreateOrUpdate
			log.Error(err, "unable to ensure Service is correct")
			return ctrl.Result{}, err
		}

		loadbalancer.Status.State = "Creating"
		if err := r.Status().Update(ctx, loadbalancer); err != nil {
			log.Error(err, "unable to update LoadBalancer status")
			return ctrl.Result{}, err
		}
	} else if loadbalancer.Status.State != "Creating" {
		pods := &corev1.PodList{}
		namespace := "kubeberth"
		if err := r.List(ctx, pods, &client.ListOptions{
			Namespace:     namespace,
			LabelSelector: labels.SelectorFromSet(map[string]string{"kubevirt.io": "virt-launcher"}),
		}); err != nil {
			if k8serrors.IsNotFound(err) {
				return ctrl.Result{}, nil
			}
			return ctrl.Result{}, err
		}

		var statusPods []berthv1alpha1.ToPod
		for _, pod := range pods.Items {
			statusPods = append(statusPods, berthv1alpha1.ToPod{Pod: pod.GetName()})
			/*
				createdVMIPod := &corev1.Service{}
				createdVMIPodNsN := types.NamespacedName{
					Namespace: loadbalancer.Namespace,
				}
			*/
		}
		loadbalancer.Status.Pods = statusPods
		if err := r.Status().Update(ctx, loadbalancer); err != nil {
			log.Error(err, "unable to update LoadBalancer status")
			return ctrl.Result{}, err
		}

		creatingService := &corev1.Service{}
		creatingServiceNsN := types.NamespacedName{
			Namespace: loadbalancer.Namespace,
			Name:      loadbalancer.Name,
		}
		if err := r.Get(ctx, creatingServiceNsN, creatingService); err != nil {
			if k8serrors.IsNotFound(err) {
				return ctrl.Result{}, nil
			}
			return ctrl.Result{}, err
		}

		if len(creatingService.Status.LoadBalancer.Ingress) > 0 {
			loadbalancer.Status.IP = creatingService.Status.LoadBalancer.Ingress[0].IP
			loadbalancer.Status.State = "Created"
			if err := r.Status().Update(ctx, loadbalancer); err != nil {
				log.Error(err, "unable to update LoadBalancer status")
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *LoadBalancerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&berthv1alpha1.LoadBalancer{}).
		Complete(r)
}
