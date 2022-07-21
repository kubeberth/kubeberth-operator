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
	"strings"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/go-logr/logr"

	berthv1alpha1 "github.com/kubeberth/kubeberth-operator/api/v1alpha1"
)

const (
	archiveRequeueAfter          = time.Second * 30
	archiveFinalizerRequeueAfter = time.Second * 3
	archiveFinalizerName         = "finalizers.archives.berth.kubeberth.io"
)

// ArchiveReconciler reconciles a Archive object
type ArchiveReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=berth.kubeberth.io,resources=archives,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=berth.kubeberth.io,resources=archives/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=berth.kubeberth.io,resources=archives/finalizers,verbs=update
//+kubebuilder:rbac:groups="batch",resources=jobs,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Archive object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *ArchiveReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("Archive", req.NamespacedName)

	// Get the archive.
	archive := &berthv1alpha1.Archive{}
	if err := r.Get(ctx, req.NamespacedName, archive); err != nil && !k8serrors.IsNotFound(err) {
		log.Error(err, "could not get the Archive resource")
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{Requeue: false}, nil
		}
		return ctrl.Result{Requeue: true}, err
	}

	if deleted, deleting, err := r.handleFinalizer(ctx, archive); err != nil {
		log.Error(err, "failed to do handleFinalizer")
		return ctrl.Result{Requeue: true}, err
	} else if deleted {
		return ctrl.Result{Requeue: false}, nil
	} else if deleting {
		return ctrl.Result{Requeue: true, RequeueAfter: archiveFinalizerRequeueAfter}, nil
	}

	if ensuring, err := r.ensureArchiveExists(ctx, archive); err != nil {
		log.Error(err, "failed to do ensureArchiveExists")
		return ctrl.Result{Requeue: true}, err
	} else if ensuring {
		log.Error(err, "ensureing Archive Exists")
		return ctrl.Result{Requeue: true, RequeueAfter: archiveRequeueAfter}, nil
	}

	return ctrl.Result{Requeue: false}, nil
}

func (r *ArchiveReconciler) checkArchiving(ctx context.Context, nsn types.NamespacedName) (archiving bool) {
	createdJob := &batchv1.Job{}
	if err := r.Get(ctx, nsn, createdJob); err != nil {
		return false
	}
	return true
}

func (r *ArchiveReconciler) checkFinishArchiving(ctx context.Context, nsn types.NamespacedName) (finished bool, err error) {
	createdJob := &batchv1.Job{}
	if err := r.Get(ctx, nsn, createdJob); err == nil {
		if len(createdJob.Status.Conditions) > 0 {
			if createdJob.Status.Conditions[0].Type == batchv1.JobComplete {
				return true, nil
			} else {
				return false, nil
			}
		} else {
			return false, nil
		}
	} else {
		return false, err
	}
}

func (r *ArchiveReconciler) createJobSpecForSourceFilesystemDisk(ctx context.Context, kubeberth *berthv1alpha1.KubeBerth, archive *berthv1alpha1.Archive) *batchv1.JobSpec {
	return &batchv1.JobSpec{
		Template: corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					corev1.Container{
						Name:            "berth-archiver-creating-" + archive.GetName(),
						Image:           "kubeberth/berth-archiver:v1alpha1",
						ImagePullPolicy: corev1.PullAlways,
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("1"),
								corev1.ResourceMemory: resource.MustParse("2Gi"),
							},
						},
						Env: []corev1.EnvVar{
							corev1.EnvVar{
								Name:  "MODE",
								Value: "Create",
							},
							corev1.EnvVar{
								Name:  "URL",
								Value: kubeberth.Spec.ArchiveRepositoryURL,
							},
							corev1.EnvVar{
								Name:  "ACCESS_KEY",
								Value: kubeberth.Spec.AccessKey,
							},
							corev1.EnvVar{
								Name:  "SECRET_KEY",
								Value: kubeberth.Spec.SecretKey,
							},
							corev1.EnvVar{
								Name:  "TARGET",
								Value: kubeberth.Spec.ArchiveRepositoryTarget,
							},
							corev1.EnvVar{
								Name:  "SOURCE",
								Value: "Disk",
							},
							corev1.EnvVar{
								Name:  "SOURCE_NAME",
								Value: archive.Spec.Source.Disk.Name,
							},
							corev1.EnvVar{
								Name:  "ARCHIVE_NAME",
								Value: archive.GetName(),
							},
							corev1.EnvVar{
								Name:  "VOLUME_MODE",
								Value: "Filesystem",
							},
						},
						VolumeMounts: []corev1.VolumeMount{
							corev1.VolumeMount{
								Name:      "disk",
								MountPath: "/tmp/" + archive.Spec.Source.Disk.Name,
							},
						},
					},
				},
				RestartPolicy: corev1.RestartPolicyNever,
				Volumes: []corev1.Volume{
					corev1.Volume{
						Name: "disk",
						VolumeSource: corev1.VolumeSource{
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
								ClaimName: archive.Spec.Source.Disk.Name + "-disk",
							},
						},
					},
				},
			},
		},
	}
}

func (r *ArchiveReconciler) createJobSpecForSourceBlockDisk(ctx context.Context, kubeberth *berthv1alpha1.KubeBerth, archive *berthv1alpha1.Archive) *batchv1.JobSpec {
	return &batchv1.JobSpec{
		Template: corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					corev1.Container{
						Name:            "berth-archiver-creating-" + archive.GetName(),
						Image:           "kubeberth/berth-archiver:v1alpha1",
						ImagePullPolicy: corev1.PullAlways,
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("1"),
								corev1.ResourceMemory: resource.MustParse("2Gi"),
							},
						},
						Env: []corev1.EnvVar{
							corev1.EnvVar{
								Name:  "MODE",
								Value: "Create",
							},
							corev1.EnvVar{
								Name:  "URL",
								Value: kubeberth.Spec.ArchiveRepositoryURL,
							},
							corev1.EnvVar{
								Name:  "ACCESS_KEY",
								Value: kubeberth.Spec.AccessKey,
							},
							corev1.EnvVar{
								Name:  "SECRET_KEY",
								Value: kubeberth.Spec.SecretKey,
							},
							corev1.EnvVar{
								Name:  "TARGET",
								Value: kubeberth.Spec.ArchiveRepositoryTarget,
							},
							corev1.EnvVar{
								Name:  "SOURCE",
								Value: "Disk",
							},
							corev1.EnvVar{
								Name:  "SOURCE_NAME",
								Value: archive.Spec.Source.Disk.Name,
							},
							corev1.EnvVar{
								Name:  "ARCHIVE_NAME",
								Value: archive.GetName(),
							},
							corev1.EnvVar{
								Name:  "VOLUME_MODE",
								Value: "Block",
							},
						},
						VolumeDevices: []corev1.VolumeDevice{
							corev1.VolumeDevice{
								Name:       "disk",
								DevicePath: "/dev/" + archive.Spec.Source.Disk.Name,
							},
						},
					},
				},
				RestartPolicy: corev1.RestartPolicyNever,
				Volumes: []corev1.Volume{
					corev1.Volume{
						Name: "disk",
						VolumeSource: corev1.VolumeSource{
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
								ClaimName: archive.Spec.Source.Disk.Name + "-disk",
							},
						},
					},
				},
			},
		},
	}
}

func (r *ArchiveReconciler) createJobSpecForSourceArchive(ctx context.Context, kubeberth *berthv1alpha1.KubeBerth, archive *berthv1alpha1.Archive) *batchv1.JobSpec {
	return &batchv1.JobSpec{
		Template: corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					corev1.Container{
						Name:            "berth-archiver-creating-" + archive.GetName(),
						Image:           "kubeberth/berth-archiver:v1alpha1",
						ImagePullPolicy: corev1.PullAlways,
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("1"),
								corev1.ResourceMemory: resource.MustParse("2Gi"),
							},
						},
						Env: []corev1.EnvVar{
							corev1.EnvVar{
								Name:  "MODE",
								Value: "Create",
							},
							corev1.EnvVar{
								Name:  "URL",
								Value: kubeberth.Spec.ArchiveRepositoryURL,
							},
							corev1.EnvVar{
								Name:  "ACCESS_KEY",
								Value: kubeberth.Spec.AccessKey,
							},
							corev1.EnvVar{
								Name:  "SECRET_KEY",
								Value: kubeberth.Spec.SecretKey,
							},
							corev1.EnvVar{
								Name:  "TARGET",
								Value: kubeberth.Spec.ArchiveRepositoryTarget,
							},
							corev1.EnvVar{
								Name:  "SOURCE",
								Value: "Archive",
							},
							corev1.EnvVar{
								Name:  "SOURCE_NAME",
								Value: archive.Spec.Source.Archive.Name,
							},
							corev1.EnvVar{
								Name:  "ARCHIVE_NAME",
								Value: archive.GetName(),
							},
						},
					},
				},
				RestartPolicy: corev1.RestartPolicyNever,
			},
		},
	}
}

func (r *ArchiveReconciler) ensureArchiveExists(ctx context.Context, archive *berthv1alpha1.Archive) (ensuring bool, err error) {
	kubeberth := &berthv1alpha1.KubeBerth{}
	kubeberthNsN := types.NamespacedName{
		Namespace: "kubeberth",
		Name:      "kubeberth",
	}
	if err := r.Get(ctx, kubeberthNsN, kubeberth); err != nil && !k8serrors.IsNotFound(err) {
		return true, err
	}

	if archive.Spec.Repository != "" {
		archive.Status.State = "Created"
		archive.Status.Repository = archive.Spec.Repository
		if err := r.Status().Update(ctx, archive); err != nil {
			return true, err
		}
	} else if archive.Spec.Source.Archive != nil {
		nsn := types.NamespacedName{
			Namespace: archive.GetNamespace(),
			Name:      "berth-archiver-creating-" + archive.GetName(),
		}
		if archiving := r.checkArchiving(ctx, nsn); archiving {
			if finished, err := r.checkFinishArchiving(ctx, nsn); err != nil {
				return true, err
			} else if !finished {
				return true, nil
			} else {
				createdJob := &batchv1.Job{}
				nsn := types.NamespacedName{
					Namespace: archive.GetNamespace(),
					Name:      "berth-archiver-creating-" + archive.GetName(),
				}
				if err := r.Get(ctx, nsn, createdJob); err != nil {
					return true, err
				}
				deletionPropagation := metav1.DeletePropagationForeground
				if err := r.Delete(ctx, createdJob, &client.DeleteOptions{
					PropagationPolicy: &deletionPropagation,
				}); err != nil {
					return true, err
				}

				archive.Status.State = "Created"
				archive.Status.Repository = kubeberth.Spec.ArchiveRepositoryURL + "/" + strings.Join(strings.Split(kubeberth.Spec.ArchiveRepositoryTarget, "/")[1:], "/") + "/" + archive.GetName() + ".qcow2"
				if err := r.Status().Update(ctx, archive); err != nil {
					return true, err
				}
				return false, nil
			}
		}

		job := &batchv1.Job{}
		job.SetNamespace(archive.GetNamespace())
		job.SetName("berth-archiver-creating-" + archive.GetName())
		spec := r.createJobSpecForSourceArchive(ctx, kubeberth, archive)
		if _, err := ctrl.CreateOrUpdate(ctx, r.Client, job, func() error {
			job.Spec = *spec
			if err := ctrl.SetControllerReference(archive, job, r.Scheme); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return true, err
		}

		archive.Status.State = "Creating"
		if err := r.Status().Update(ctx, archive); err != nil {
			return true, err
		}

		return true, nil
	} else if archive.Spec.Source.Disk != nil {
		nsn := types.NamespacedName{
			Namespace: archive.GetNamespace(),
			Name:      "berth-archiver-creating-" + archive.GetName(),
		}
		if archiving := r.checkArchiving(ctx, nsn); archiving {
			if finished, err := r.checkFinishArchiving(ctx, nsn); err != nil {
				return true, err
			} else if !finished {
				return true, nil
			} else {
				createdJob := &batchv1.Job{}
				nsn := types.NamespacedName{
					Namespace: archive.GetNamespace(),
					Name:      "berth-archiver-creating-" + archive.GetName(),
				}
				if err := r.Get(ctx, nsn, createdJob); err != nil {
					return true, err
				}
				deletionPropagation := metav1.DeletePropagationForeground
				if err := r.Delete(ctx, createdJob, &client.DeleteOptions{
					PropagationPolicy: &deletionPropagation,
				}); err != nil {
					return true, err
				}

				archive.Status.State = "Created"
				archive.Status.Repository = kubeberth.Spec.ArchiveRepositoryURL + "/" + strings.Join(strings.Split(kubeberth.Spec.ArchiveRepositoryTarget, "/")[1:], "/") + "/" + archive.GetName() + ".qcow2"
				if err := r.Status().Update(ctx, archive); err != nil {
					return true, err
				}

				attachedDisk := &berthv1alpha1.Disk{}
				attachedDiskNsN := types.NamespacedName{
					Namespace: archive.GetNamespace(),
					Name:      archive.Spec.Source.Disk.Name,
				}
				if err := r.Get(ctx, attachedDiskNsN, attachedDisk); err != nil {
					return true, err
				}
				attachedDisk.Status.State = "Detached"
				attachedDisk.Status.AttachedTo = ""
				if err := r.Status().Update(ctx, attachedDisk); err != nil {
					return true, err
				}

				return false, nil
			}
		}

		kubeberth := &berthv1alpha1.KubeBerth{}
		kubeberthNsN := types.NamespacedName{
			Namespace: "kubeberth",
			Name:      "kubeberth",
		}
		if err := r.Get(ctx, kubeberthNsN, kubeberth); err != nil && !k8serrors.IsNotFound(err) {
			return true, err
		}

		attachedDisk := &berthv1alpha1.Disk{}
		attachedDiskNsN := types.NamespacedName{
			Namespace: archive.GetNamespace(),
			Name:      archive.Spec.Source.Disk.Name,
		}
		if err := r.Get(ctx, attachedDiskNsN, attachedDisk); err != nil {
			return true, err
		}
		attachedDisk.Status.State = "Attached"
		attachedDisk.Status.AttachedTo = "archive.berth.kubeberth.io/" + archive.GetName()
		if err := r.Status().Update(ctx, attachedDisk); err != nil {
			return true, err
		}

		var spec *batchv1.JobSpec
		job := &batchv1.Job{}
		job.SetNamespace(archive.GetNamespace())
		job.SetName("berth-archiver-creating-" + archive.GetName())
		if *kubeberth.Status.VolumeMode == corev1.PersistentVolumeBlock {
			spec = r.createJobSpecForSourceBlockDisk(ctx, kubeberth, archive)
		} else if *kubeberth.Status.VolumeMode == corev1.PersistentVolumeFilesystem {
			spec = r.createJobSpecForSourceFilesystemDisk(ctx, kubeberth, archive)
		}

		if _, err := ctrl.CreateOrUpdate(ctx, r.Client, job, func() error {
			job.Spec = *spec
			if err := ctrl.SetControllerReference(archive, job, r.Scheme); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return true, err
		}

		archive.Status.State = "Creating"
		if err := r.Status().Update(ctx, archive); err != nil {
			return true, err
		}

		return true, nil
	}

	return false, nil
}

func (r *ArchiveReconciler) handleFinalizer(ctx context.Context, archive *berthv1alpha1.Archive) (deleted, deleting bool, err error) {
	if archive.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(archive, archiveFinalizerName) {
			controllerutil.AddFinalizer(archive, archiveFinalizerName)
			if err := r.Update(ctx, archive); err != nil {
				return false, false, err
			}
		}
	} else {
		if controllerutil.ContainsFinalizer(archive, archiveFinalizerName) {
			createdJob := &batchv1.Job{}
			nsn := types.NamespacedName{
				Namespace: archive.GetNamespace(),
				Name:      "berth-archiver-deleting-" + archive.GetName(),
			}
			if err := r.Get(ctx, nsn, createdJob); err == nil {
				if len(createdJob.Status.Conditions) > 0 {
					if createdJob.Status.Conditions[0].Type == batchv1.JobComplete {
						archive.Status.State = "Deleted"
						archive.Status.Repository = ""
						if err := r.Status().Update(ctx, archive); err != nil {
							return false, true, err
						}
						controllerutil.RemoveFinalizer(archive, archiveFinalizerName)
						if err := r.Update(ctx, archive); err != nil {
							return false, true, err
						}
						return true, false, nil
					} else {
						return false, true, nil
					}
				} else {
					return false, true, nil
				}
			}

			kubeberth := &berthv1alpha1.KubeBerth{}
			kubeberthNsN := types.NamespacedName{
				Namespace: "kubeberth",
				Name:      "kubeberth",
			}
			if err := r.Get(ctx, kubeberthNsN, kubeberth); err != nil && !k8serrors.IsNotFound(err) {
				return false, true, err
			}

			job := &batchv1.Job{}
			job.SetNamespace(archive.GetNamespace())
			job.SetName("berth-archiver-deleting-" + archive.GetName())
			if _, err := ctrl.CreateOrUpdate(ctx, r.Client, job, func() error {
				job.Spec = batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								corev1.Container{
									Name:            "berth-archiver-deleting-" + archive.GetName(),
									Image:           "kubeberth/berth-archiver:v1alpha1",
									ImagePullPolicy: corev1.PullAlways,
									Resources: corev1.ResourceRequirements{
										Limits: corev1.ResourceList{
											corev1.ResourceCPU:    resource.MustParse("1"),
											corev1.ResourceMemory: resource.MustParse("2Gi"),
										},
									},
									Env: []corev1.EnvVar{
										corev1.EnvVar{
											Name:  "MODE",
											Value: "Delete",
										},
										corev1.EnvVar{
											Name:  "URL",
											Value: kubeberth.Spec.ArchiveRepositoryURL,
										},
										corev1.EnvVar{
											Name:  "ACCESS_KEY",
											Value: kubeberth.Spec.AccessKey,
										},
										corev1.EnvVar{
											Name:  "SECRET_KEY",
											Value: kubeberth.Spec.SecretKey,
										},
										corev1.EnvVar{
											Name:  "TARGET",
											Value: kubeberth.Spec.ArchiveRepositoryTarget,
										},
										corev1.EnvVar{
											Name:  "ARCHIVE_NAME",
											Value: archive.GetName(),
										},
									},
								},
							},
							RestartPolicy: corev1.RestartPolicyNever,
						},
					},
				}
				if err := ctrl.SetControllerReference(archive, job, r.Scheme); err != nil {
					return err
				}
				return nil
			}); err != nil {
				return false, true, err
			}

			archive.Status.State = "Deleting"
			if err := r.Status().Update(ctx, archive); err != nil {
				return false, true, nil
			}

			return false, true, nil
		}

		return true, false, nil
	}

	return false, false, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ArchiveReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&berthv1alpha1.Archive{}).
		Complete(r)
}
