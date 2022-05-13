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

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	berthv1alpha1 "github.com/kubeberth/berth-operator/api/v1alpha1"
	kubevirtv1 "kubevirt.io/api/core/v1"
)

// ServerReconciler reconciles a Server object
type ServerReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=berth.kubeberth.io,resources=servers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=berth.kubeberth.io,resources=servers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=berth.kubeberth.io,resources=servers/finalizers,verbs=update
//+kubebuilder:rbac:groups=berth.kubeberth.io,resources=disks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cdi.kubevirt.io,resources=dvs,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Server object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *ServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("Server", req.NamespacedName)

	// Get the Server.
	server := &berthv1alpha1.Server{}
	if err := r.Get(ctx, req.NamespacedName, server); err != nil {
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	diskNsN := types.NamespacedName{
		Namespace: server.Spec.Disk.Namespace,
		Name:      server.Spec.Disk.Name,
	}

	// Get the Disk.
	disk := &berthv1alpha1.Disk{}
	if err := r.Get(ctx, diskNsN, disk); err != nil {
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	cloudInitNsN := types.NamespacedName{
		Namespace: server.Spec.CloudInit.Namespace,
		Name:      server.Spec.CloudInit.Name,
	}

	// Get the CloudInit.
	cloudInit := &berthv1alpha1.CloudInit{}
	if err := r.Get(ctx, cloudInitNsN, cloudInit); err != nil {
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	vm := &kubevirtv1.VirtualMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      server.Name,
			Namespace: server.Namespace,
		},
	}

	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, vm, func() error {
		resourceRequest := corev1.ResourceList{}
		resourceRequest[corev1.ResourceMemory] = resource.MustParse(server.Spec.Memory.String())
		readOnly := true
		userData := cloudInit.Spec.UserData

		vm.Spec = kubevirtv1.VirtualMachineSpec{
			Running: server.Spec.Running,
			Template: &kubevirtv1.VirtualMachineInstanceTemplateSpec{
				Spec: kubevirtv1.VirtualMachineInstanceSpec{
					Hostname: server.Spec.HostName,
					Domain: kubevirtv1.DomainSpec{
						CPU: &kubevirtv1.CPU{
							Cores: uint32(server.Spec.CPU.Value()),
						},
						Resources: kubevirtv1.ResourceRequirements{
							Requests: resourceRequest,
						},
						Devices: kubevirtv1.Devices{
							Disks: []kubevirtv1.Disk{
								kubevirtv1.Disk{
									Name: server.Spec.Disk.Name,
									DiskDevice: kubevirtv1.DiskDevice{
										Disk: &kubevirtv1.DiskTarget{
											Bus: "virtio",
										},
									},
								},
								kubevirtv1.Disk{
									Name: cloudInit.Name,
									DiskDevice: kubevirtv1.DiskDevice{
										CDRom: &kubevirtv1.CDRomTarget{
											Bus:      "scsi",
											ReadOnly: &readOnly,
										},
									},
								},
							},
							Interfaces: []kubevirtv1.Interface{
								kubevirtv1.Interface{
									Name:       "default",
									MacAddress: server.Spec.MACAddress,
									InterfaceBindingMethod: kubevirtv1.InterfaceBindingMethod{
										Masquerade: &kubevirtv1.InterfaceMasquerade{},
									},
								},
							},
						},
					},
					Networks: []kubevirtv1.Network{
						kubevirtv1.Network{
							Name: "default",
							NetworkSource: kubevirtv1.NetworkSource{
								Pod: &kubevirtv1.PodNetwork{},
							},
						},
					},
					Volumes: []kubevirtv1.Volume{
						kubevirtv1.Volume{
							Name: server.Spec.Disk.Name,
							VolumeSource: kubevirtv1.VolumeSource{
								DataVolume: &kubevirtv1.DataVolumeSource{
									Name: server.Spec.Disk.Name,
								},
							},
						},
						kubevirtv1.Volume{
							Name: cloudInit.Name,
							VolumeSource: kubevirtv1.VolumeSource{
								CloudInitNoCloud: &kubevirtv1.CloudInitNoCloudSource{
									UserData: userData,
								},
							},
						},
					},
				},
			},
		}

		if err := ctrl.SetControllerReference(server, vm, r.Scheme); err != nil {
			log.Error(err, "unable to set contrrollerReference from Server to VirtualMachine")
			return err
		}
		return nil

	}); err != nil {
		// error handling of ctrl.CreateOrUpdate
		log.Error(err, "unable to ensure VirtualMachine is correct")
		return ctrl.Result{}, err
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
		"external-dns.alpha.kubernetes.io/hostname": server.Name + "." + externalDNSDomainName,
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        server.Name,
			Namespace:   server.Namespace,
			Annotations: annotations,
		},
	}

	servicePort := []corev1.ServicePort{
		corev1.ServicePort{
			Name:       "ssh",
			Protocol:   corev1.ProtocolTCP,
			Port:       22,
			TargetPort: intstr.FromInt(22),
		},
	}

	labels := map[string]string{
		"vm.kubevirt.io/name": server.Name,
	}

	service.Spec.Type = corev1.ServiceTypeLoadBalancer
	service.Spec.ExternalTrafficPolicy = corev1.ServiceExternalTrafficPolicyTypeLocal
	service.Spec.Ports = servicePort
	service.Spec.Selector = labels

	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, service, func() error {
		if err := ctrl.SetControllerReference(server, service, r.Scheme); err != nil {
			log.Error(err, "unable to set controllerReference from Server to Service")
			return err
		}
		return nil
	}); err != nil {
		// error handling of ctrl.CreateOrUpdate
		log.Error(err, "unable to ensure Service is correct")
		return ctrl.Result{}, err
	}

	if err := r.Get(ctx, req.NamespacedName, server); err != nil {
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	createdService := &corev1.Service{}
	serviceNsN := types.NamespacedName{
		Namespace: server.Namespace,
		Name:      server.Name,
	}

	if err := r.Get(ctx, serviceNsN, createdService); err != nil {
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	/*
		disk.Status.State = "Attached"
		if err := r.Status().Update(ctx, disk); err != nil {
			log.Error(err, "unable to update Disk status")
			return ctrl.Result{}, err
		}
	*/

	if *server.Spec.Running {
		server.Status.State = "Running"
	} else {
		server.Status.State = "Stopped"
	}

	server.Status.CPU = server.Spec.CPU.String()
	server.Status.Memory = server.Spec.Memory.String()
	server.Status.HostName = server.Spec.HostName

	if len(createdService.Status.LoadBalancer.Ingress) > 0 {
		server.Status.IP = createdService.Status.LoadBalancer.Ingress[0].IP
	}

	if err := r.Status().Update(ctx, server); err != nil {
		log.Error(err, "unable to update Server status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&berthv1alpha1.Server{}).
		Complete(r)
}
