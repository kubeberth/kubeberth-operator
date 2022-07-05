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
	"reflect"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	berthv1alpha1 "github.com/kubeberth/kubeberth-operator/api/v1alpha1"
)

const (
	loadbalancerFinalizerName = "finalizers.loadbalancers.berth.kubeberth.io"
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
	log := r.Log.WithValues("LoadBalancer", req.NamespacedName)

	// Get the LoadBalancer.
	loadbalancer := &berthv1alpha1.LoadBalancer{}
	if err := r.Get(ctx, req.NamespacedName, loadbalancer); err != nil {
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if deleted, err := r.handleFinalizer(ctx, loadbalancer); err != nil {
		log.Error(err, "failed to do handleFinalizer")
		return ctrl.Result{Requeue: true}, err
	} else if deleted {
		return ctrl.Result{Requeue: false}, nil
	}

	if err := r.ensureServiceExists(ctx, loadbalancer); err != nil {
		log.Error(err, "failed to do ensureServiceExists")
		return ctrl.Result{Requeue: true}, err
	}

	if err := r.ensureLoadBalancerExists(ctx, loadbalancer); err != nil {
		log.Error(err, "failed to do ensureLoadBalancerExists")
		return ctrl.Result{Requeue: true}, err
	}

	return ctrl.Result{Requeue: false}, nil
}

func (r *LoadBalancerReconciler) ensureLoadBalancerExists(ctx context.Context, loadbalancer *berthv1alpha1.LoadBalancer) error {
	if !reflect.DeepEqual(loadbalancer.Spec.Backends, loadbalancer.Status.Backends) {
		pods := &corev1.PodList{}
		if err := r.List(ctx, pods, &client.ListOptions{
			Namespace: loadbalancer.GetNamespace(),
			LabelSelector: labels.SelectorFromSet(
				map[string]string{
					"berth.kubeberth.io/loadbalancer-" + loadbalancer.GetName(): loadbalancer.GetName(),
				},
			),
		}); err != nil && !k8serrors.IsNotFound(err) {
			return err
		}

		for _, pod := range pods.Items {
			copiedPod := pod.DeepCopy()
			delete(copiedPod.Labels, "berth.kubeberth.io/loadbalancer-"+loadbalancer.GetName())
			patch := client.MergeFrom(&pod)
			if err := r.Patch(ctx, copiedPod, patch); err != nil {
				return err
			}
		}

		destinations := []berthv1alpha1.Destination{}
		for _, destination := range loadbalancer.Spec.Backends {
			server := &berthv1alpha1.Server{}
			serverNsN := types.NamespacedName{
				Namespace: loadbalancer.GetNamespace(),
				Name:      destination.Server,
			}
			if err := r.Get(ctx, serverNsN, server); err != nil && !k8serrors.IsNotFound(err) {
				return err
			}
			destinations = append(destinations, berthv1alpha1.Destination{Server: server.GetName()})
		}
		copiedLB := loadbalancer.DeepCopy()
		copiedLB.Status.State = "Updating"
		copiedLB.Status.Backends = destinations
		copiedLB.Status.Health = "Unhealthy"
		patch := client.MergeFrom(loadbalancer)
		if err := r.Status().Patch(ctx, copiedLB, patch); err != nil {
			return err
		}

		return nil
	}

	if loadbalancer.Status.IP == "None" {
		service := &corev1.Service{}
		nsn := types.NamespacedName{
			Namespace: loadbalancer.GetNamespace() + "-loadbalancer",
			Name:      loadbalancer.GetName(),
		}
		if err := r.Get(ctx, nsn, service); err != nil && !k8serrors.IsNotFound(err) {
			return err
		}
		if len(service.Status.LoadBalancer.Ingress) > 0 {
			copiedLB := loadbalancer.DeepCopy()
			copiedLB.Status.State = "Created"
			copiedLB.Status.IP = service.Status.LoadBalancer.Ingress[0].IP
			patch := client.MergeFrom(loadbalancer)
			if err := r.Status().Patch(ctx, copiedLB, patch); err != nil {
				return err
			}
		} else {
			return errors.New("Failed to get service loadbalancer ingress IP address")
		}
	}

	health := true
	for _, destination := range loadbalancer.Status.Backends {
		pods := &corev1.PodList{}
		namespace := loadbalancer.GetNamespace()
		if err := r.List(ctx, pods, &client.ListOptions{
			Namespace: namespace,
			LabelSelector: labels.SelectorFromSet(
				map[string]string{
					"kubevirt.io":               "virt-launcher",
					"berth.kubeberth.io/server": destination.Server,
				},
			),
		}); err != nil || len(pods.Items) == 0 {
			health = false
		}

		for _, pod := range pods.Items {
			if _, ok := pod.Labels["berth.kubeberth.io/loadbalancer-"+loadbalancer.GetName()]; !ok {
				copiedPod := pod.DeepCopy()
				copiedPod.Labels["berth.kubeberth.io/loadbalancer-"+loadbalancer.GetName()] = loadbalancer.GetName()
				patch := client.MergeFrom(&pod)
				if err := r.Patch(ctx, copiedPod, patch); err != nil {
					return err
				}
			}
		}
	}

	if health {
		copiedLB := loadbalancer.DeepCopy()
		copiedLB.Status.State = "Created"
		copiedLB.Status.Health = "Healthy"
		patch := client.MergeFrom(loadbalancer)
		if err := r.Status().Patch(ctx, copiedLB, patch); err != nil {
			return err
		}
	} else {
		copiedLB := loadbalancer.DeepCopy()
		copiedLB.Status.State = "Created"
		copiedLB.Status.Health = "Unhealthy"
		patch := client.MergeFrom(loadbalancer)
		if err := r.Status().Patch(ctx, copiedLB, patch); err != nil {
			return err
		}
	}

	return nil
}

func (r *LoadBalancerReconciler) checkServiceExists(ctx context.Context, nsn types.NamespacedName) bool {
	service := &corev1.Service{}
	if err := r.Get(ctx, nsn, service); err != nil {
		return false
	}
	return true
}

func (r *LoadBalancerReconciler) ensureServiceExists(ctx context.Context, loadbalancer *berthv1alpha1.LoadBalancer) error {
	nsn := types.NamespacedName{
		Namespace: loadbalancer.GetNamespace(),
		Name:      loadbalancer.GetName(),
	}
	if ok := r.checkServiceExists(ctx, nsn); ok {
		return nil
	}

	kubeberth := &berthv1alpha1.KubeBerth{}
	kubeberthNsN := types.NamespacedName{
		Namespace: "kubeberth",
		Name:      "kubeberth",
	}
	if err := r.Get(ctx, kubeberthNsN, kubeberth); err != nil && !k8serrors.IsNotFound(err) {
		return err
	}
	externalDNSDomainName := kubeberth.Spec.ExternalDNSDomainName
	annotations := map[string]string{
		"external-dns.alpha.kubernetes.io/hostname": loadbalancer.GetName() + ".lb." + externalDNSDomainName,
	}

	service := &corev1.Service{}
	service.SetNamespace(loadbalancer.GetNamespace())
	service.SetName(loadbalancer.GetName() + "-loadbalancer")
	service.SetAnnotations(annotations)
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, service, func() error {
		servicePort := loadbalancer.Spec.Ports
		lbLabels := map[string]string{
			"berth.kubeberth.io/loadbalancer-" + loadbalancer.GetName(): loadbalancer.GetName(),
		}
		service.Spec.Type = corev1.ServiceTypeLoadBalancer
		//service.Spec.ExternalTrafficPolicy = corev1.ServiceExternalTrafficPolicyTypeLocal
		service.Spec.Ports = servicePort
		service.Spec.Selector = lbLabels
		if err := ctrl.SetControllerReference(loadbalancer, service, r.Scheme); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	// TODO: Implementation for Validating Admission Webhook
	/*
		destinations := []berthv1alpha1.Destination{}
		for _, destination := range loadbalancer.Spec.Backends {
			server := &berthv1alpha1.Server{}
			serverNsN := types.NamespacedName{
				Namespace: loadbalancer.GetNamespace(),
				Name:      destination.Server,
			}
			if err := r.Get(ctx, serverNsN, server); err != nil && !k8serrors.IsNotFound(err) {
				return err
			}
			destinations = append(destinations, berthv1alpha1.Destination{Server: server.GetName()})
		}
	*/
	copiedLB := loadbalancer.DeepCopy()
	copiedLB.Status.State = "Creating"
	copiedLB.Status.IP = "None"
	//copiedLB.Status.Backends = destinations
	copiedLB.Status.Backends = loadbalancer.Spec.Backends
	copiedLB.Status.Health = "Unhealthy"
	patch := client.MergeFrom(loadbalancer)
	if err := r.Status().Patch(ctx, copiedLB, patch); err != nil {
		return err
	}

	return nil
}

func (r *LoadBalancerReconciler) handleFinalizer(ctx context.Context, loadbalancer *berthv1alpha1.LoadBalancer) (deleted bool, err error) {
	if loadbalancer.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(loadbalancer, loadbalancerFinalizerName) {
			controllerutil.AddFinalizer(loadbalancer, loadbalancerFinalizerName)
			if err := r.Update(ctx, loadbalancer); err != nil {
				return false, err
			}
		}
	} else {
		if controllerutil.ContainsFinalizer(loadbalancer, loadbalancerFinalizerName) {
			pods := &corev1.PodList{}
			if err := r.List(ctx, pods, &client.ListOptions{
				Namespace:     loadbalancer.GetNamespace(),
				LabelSelector: labels.SelectorFromSet(map[string]string{"berth.kubeberth.io/loadbalancer-" + loadbalancer.GetName(): loadbalancer.GetName()}),
			}); err != nil {
				if k8serrors.IsNotFound(err) {
					return false, nil
				}
				return false, err
			}
			for _, pod := range pods.Items {
				copiedPod := pod.DeepCopy()
				delete(copiedPod.Labels, "berth.kubeberth.io/loadbalancer-"+loadbalancer.GetName())
				patch := client.MergeFrom(&pod)
				if err := r.Patch(ctx, copiedPod, patch); err != nil {
					return false, err
				}
			}

			controllerutil.RemoveFinalizer(loadbalancer, loadbalancerFinalizerName)
			if err := r.Update(ctx, loadbalancer); err != nil {
				return false, err
			}
		}
		return true, nil
	}
	return false, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *LoadBalancerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&berthv1alpha1.LoadBalancer{}).
		Complete(r)
}
