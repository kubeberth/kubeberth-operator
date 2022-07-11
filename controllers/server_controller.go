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
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	berthv1alpha1 "github.com/kubeberth/kubeberth-operator/api/v1alpha1"
	kubevirtv1 "kubevirt.io/api/core/v1"
)

const (
	serverRequeueAfter  = time.Second * 3
	serverFinalizerName = "finalizers.servers.berth.kubeberth.io"
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
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubevirt.io,resources=virtualmachines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubevirt.io,resources=virtualmachineinstances,verbs=get;list;watch;create;update;patch;delete

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

	server := &berthv1alpha1.Server{}
	if err := r.Get(ctx, req.NamespacedName, server); err != nil {
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{Requeue: false}, nil
		}
		return ctrl.Result{Requeue: true}, err
	}

	if deleted, err := r.handleFinalizer(ctx, server); err != nil {
		log.Error(err, "failed to do handleFinalizer")
		return ctrl.Result{Requeue: true}, err
	} else if deleted {
		return ctrl.Result{Requeue: false}, nil
	}

	if err := r.ensureVirtualMachineExists(ctx, server); err != nil {
		log.Error(err, "failed to do ensureVirtualMachineExists")
		return ctrl.Result{Requeue: true}, err
	}

	if ensuring, err := r.ensureServerExists(ctx, server); err != nil {
		log.Error(err, "failed to do ensureServerExists")
		return ctrl.Result{Requeue: true}, err
	} else if ensuring {
		return ctrl.Result{Requeue: true, RequeueAfter: serverRequeueAfter}, nil
	}

	if err := r.ensureServiceExists(ctx, server); err != nil {
		log.Error(err, "failed to do ensureServiceExists")
		return ctrl.Result{Requeue: true}, err
	}

	return ctrl.Result{Requeue: false}, nil
}

func (r *ServerReconciler) createNodeSelector(server *berthv1alpha1.Server) map[string]string {
	var nodeSelector map[string]string
	if server.Spec.Hosting != "" {
		nodeSelector = map[string]string{"kubernetes.io/hostname": server.Spec.Hosting}
	}

	return nodeSelector

}

func (r *ServerReconciler) createDomainDevicesDisks(ctx context.Context, server *berthv1alpha1.Server) []kubevirtv1.Disk {
	var domainDevicesDisks []kubevirtv1.Disk

	if has, cloudinit := r.hasCloudInit(ctx, server); has {
		readOnly := true
		domainDevicesDisks = []kubevirtv1.Disk{
			kubevirtv1.Disk{
				Name: server.Spec.Disk.Name + "-disk",
				DiskDevice: kubevirtv1.DiskDevice{
					Disk: &kubevirtv1.DiskTarget{
						Bus: "virtio",
					},
				},
			},
			kubevirtv1.Disk{
				Name: cloudinit.Name + "-cloudinit",
				DiskDevice: kubevirtv1.DiskDevice{
					CDRom: &kubevirtv1.CDRomTarget{
						Bus:      "scsi",
						ReadOnly: &readOnly,
					},
				},
			},
		}
	} else {
		domainDevicesDisks = []kubevirtv1.Disk{
			kubevirtv1.Disk{
				Name: server.Spec.Disk.Name + "-disk",
				DiskDevice: kubevirtv1.DiskDevice{
					Disk: &kubevirtv1.DiskTarget{
						Bus: "virtio",
					},
				},
			},
		}
	}

	return domainDevicesDisks
}

func (r *ServerReconciler) createInterfaces(server *berthv1alpha1.Server) []kubevirtv1.Interface {
	var interfaces []kubevirtv1.Interface
	if server.Spec.MACAddress != "" {
		interfaces = []kubevirtv1.Interface{
			kubevirtv1.Interface{
				Name:       "default",
				MacAddress: server.Spec.MACAddress,
				InterfaceBindingMethod: kubevirtv1.InterfaceBindingMethod{
					Bridge: &kubevirtv1.InterfaceBridge{},
				},
			},
		}
	} else {
		interfaces = []kubevirtv1.Interface{
			kubevirtv1.Interface{
				Name: "default",
				InterfaceBindingMethod: kubevirtv1.InterfaceBindingMethod{
					Bridge: &kubevirtv1.InterfaceBridge{},
				},
			},
		}
	}

	return interfaces
}

func (r *ServerReconciler) createVolumes(ctx context.Context, server *berthv1alpha1.Server) []kubevirtv1.Volume {
	var volumes []kubevirtv1.Volume
	if has, cloudinit := r.hasCloudInit(ctx, server); has {
		volumes = []kubevirtv1.Volume{
			kubevirtv1.Volume{
				Name: server.Spec.Disk.Name + "-disk",
				VolumeSource: kubevirtv1.VolumeSource{
					DataVolume: &kubevirtv1.DataVolumeSource{
						Name: server.Spec.Disk.Name,
					},
				},
			},
			kubevirtv1.Volume{
				Name: cloudinit.Name + "-cloudinit",
				VolumeSource: kubevirtv1.VolumeSource{
					CloudInitNoCloud: &kubevirtv1.CloudInitNoCloudSource{
						UserData: cloudinit.Spec.UserData,
					},
				},
			},
		}
	} else {
		volumes = []kubevirtv1.Volume{
			kubevirtv1.Volume{
				Name: server.Spec.Disk.Name + "-disk",
				VolumeSource: kubevirtv1.VolumeSource{
					DataVolume: &kubevirtv1.DataVolumeSource{
						Name: server.Spec.Disk.Name,
					},
				},
			},
		}
	}

	return volumes
}

func (r *ServerReconciler) hasCloudInit(ctx context.Context, server *berthv1alpha1.Server) (bool, *berthv1alpha1.CloudInit) {
	if server.Spec.CloudInit != nil {
		nsn := types.NamespacedName{
			Namespace: server.GetNamespace(),
			Name:      server.Spec.CloudInit.Name,
		}
		cloudinit := &berthv1alpha1.CloudInit{}
		if err := r.Get(ctx, nsn, cloudinit); err != nil {
			return false, nil
		}

		return true, cloudinit
	}

	return false, nil
}

func (r *ServerReconciler) createVirtualMachineSpec(ctx context.Context, server *berthv1alpha1.Server) *kubevirtv1.VirtualMachineSpec {
	nodeSelector := r.createNodeSelector(server)
	resourceRequest := corev1.ResourceList{corev1.ResourceMemory: resource.MustParse(server.Spec.Memory.String())}
	domainDevicesDisks := r.createDomainDevicesDisks(ctx, server)
	interfaces := r.createInterfaces(server)
	volumes := r.createVolumes(ctx, server)

	return &kubevirtv1.VirtualMachineSpec{
		Running: server.Spec.Running,
		Template: &kubevirtv1.VirtualMachineInstanceTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"berth.kubeberth.io/server": server.GetName(),
				},
			},
			Spec: kubevirtv1.VirtualMachineInstanceSpec{
				NodeSelector: nodeSelector,
				Hostname:     server.Spec.Hostname,
				Domain: kubevirtv1.DomainSpec{
					CPU: &kubevirtv1.CPU{
						Cores: uint32(server.Spec.CPU.Value()),
					},
					Resources: kubevirtv1.ResourceRequirements{
						Requests: resourceRequest,
					},
					Devices: kubevirtv1.Devices{
						Disks:      domainDevicesDisks,
						Interfaces: interfaces,
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
				Volumes: volumes,
			},
		},
	}

}

func (r *ServerReconciler) ensureVirtualMachineExists(ctx context.Context, server *berthv1alpha1.Server) error {
	vm := &kubevirtv1.VirtualMachine{}
	vm.SetNamespace(server.GetNamespace())
	vm.SetName(server.GetName())
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, vm, func() error {
		spec := r.createVirtualMachineSpec(ctx, server)
		vm.Spec = *spec
		if err := ctrl.SetControllerReference(server, vm, r.Scheme); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (r *ServerReconciler) ensureServerExists(ctx context.Context, server *berthv1alpha1.Server) (ensuring bool, err error) {
	log := r.Log.WithValues("ensureServerExists", server.GetName())

	if (server.Status.AttachedDisk != "") && (server.Status.AttachedDisk != server.Spec.Disk.Name) {
		attachedDisk := &berthv1alpha1.Disk{}
		nsn := types.NamespacedName{
			Namespace: server.GetNamespace(),
			Name:      server.Status.AttachedDisk,
		}
		if err := r.Get(ctx, nsn, attachedDisk); err != nil {
			return true, err
		}

		attachedDisk.Status.State = "Detached"
		attachedDisk.Status.AttachedTo = ""
		if err := r.Status().Update(ctx, attachedDisk); err != nil {
			return true, err
		}

		server.Status.AttachedDisk = ""
		if err := r.Status().Update(ctx, server); err != nil {
			return true, err
		}
	}

	disk := &berthv1alpha1.Disk{}
	diskNsN := types.NamespacedName{
		Namespace: server.GetNamespace(),
		Name:      server.Spec.Disk.Name,
	}
	if err := r.Get(ctx, diskNsN, disk); err != nil {
		return true, err
	} else {
		disk.Status.State = "Attached"
		disk.Status.AttachedTo = server.GetName()
		if err := r.Status().Update(ctx, disk); err != nil {
			log.Error(err, "unable to update a status of the Disk")
			return true, err
		}
	}

	nsn := types.NamespacedName{
		Namespace: server.GetNamespace(),
		Name:      server.GetName(),
	}

	vm := &kubevirtv1.VirtualMachine{}
	if err := r.Get(ctx, nsn, vm); err != nil {
		return true, err
	} else {
		server.Status.State = (string)(vm.Status.PrintableStatus)
	}

	vmi := &kubevirtv1.VirtualMachineInstance{}
	if err := r.Get(ctx, nsn, vmi); err != nil {
		server.Status.IP = ""
		server.Status.Hosting = ""
	} else {
		server.Status.Hosting = vmi.Status.NodeName
		if len(vmi.Status.Interfaces) > 0 {
			server.Status.IP = vmi.Status.Interfaces[0].IP
		}
	}

	server.Status.CPU = server.Spec.CPU.String()
	server.Status.Memory = server.Spec.Memory.String()
	server.Status.Hostname = server.Spec.Hostname
	server.Status.AttachedDisk = disk.GetName()
	if err := r.Status().Update(ctx, server); err != nil {
		log.Error(err, "unable to update a status of the Server")
		return true, err
	}

	return true, nil
}

func (r *ServerReconciler) checkServiceExists(ctx context.Context, nsn types.NamespacedName) bool {
	service := &corev1.Service{}
	if err := r.Get(ctx, nsn, service); err != nil {
		return false
	}
	return true
}

func (r *ServerReconciler) ensureServiceExists(ctx context.Context, server *berthv1alpha1.Server) error {
	nsn := types.NamespacedName{
		Namespace: server.GetNamespace(),
		Name:      server.GetName() + "-server",
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

	service := &corev1.Service{}
	service.SetNamespace(server.GetNamespace())
	service.SetName(server.GetName() + "-server")

	if kubeberth.Status.ExternalDNSDomain != "" {
		annotations := map[string]string{
			"external-dns.alpha.kubernetes.io/hostname": server.Status.Hostname + "." + kubeberth.Status.ExternalDNSDomain,
		}
		service.SetAnnotations(annotations)
	}

	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, service, func() error {
		service.Spec.Ports = []corev1.ServicePort{
			corev1.ServicePort{
				Name:       "ssh",
				Protocol:   corev1.ProtocolTCP,
				Port:       22,
				TargetPort: intstr.FromInt(22),
			},
		}
		service.Spec.Selector = map[string]string{
			"berth.kubeberth.io/server": server.GetName(),
		}
		service.Spec.Type = corev1.ServiceTypeLoadBalancer
		service.Spec.ExternalTrafficPolicy = corev1.ServiceExternalTrafficPolicyTypeLocal
		if err := ctrl.SetControllerReference(server, service, r.Scheme); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (r *ServerReconciler) handleFinalizer(ctx context.Context, server *berthv1alpha1.Server) (deleted bool, err error) {
	if server.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(server, serverFinalizerName) {
			controllerutil.AddFinalizer(server, serverFinalizerName)
			if err := r.Update(ctx, server); err != nil {
				return false, err
			}
		}

		return false, nil
	} else {
		if controllerutil.ContainsFinalizer(server, serverFinalizerName) {
			disk := &berthv1alpha1.Disk{}
			diskNsN := types.NamespacedName{
				Namespace: server.GetNamespace(),
				Name:      server.Spec.Disk.Name,
			}
			if err := r.Get(ctx, diskNsN, disk); err == nil {
				disk.Status.State = "Detached"
				disk.Status.AttachedTo = ""
				if err := r.Status().Update(ctx, disk); err != nil {
					return false, err
				}
			}

			controllerutil.RemoveFinalizer(server, serverFinalizerName)
			if err := r.Update(ctx, server); err != nil {
				return false, err
			}
		}

		return true, nil
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&berthv1alpha1.Server{}).
		Complete(r)
}
