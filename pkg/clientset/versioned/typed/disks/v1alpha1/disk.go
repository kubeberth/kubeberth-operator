package v1alpha1

import (
	"context"
	v1alpha1 "github.com/kubeberth/kubeberth-operator/api/v1alpha1"
	scheme "github.com/kubeberth/kubeberth-operator/pkg/clientset/versioned/scheme"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

type DisksGetter interface {
	Disks(namespace string) DiskInterface
}

type DiskInterface interface {
	Create(ctx context.Context, disk *v1alpha1.Disk, opts v1.CreateOptions) (*v1alpha1.Disk, error)
	Update(ctx context.Context, disk *v1alpha1.Disk, opts v1.UpdateOptions) (*v1alpha1.Disk, error)
	UpdateStatus(ctx context.Context, disk *v1alpha1.Disk, opts v1.UpdateOptions) (*v1alpha1.Disk, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.Disk, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.DiskList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Disk, err error)
	DiskExpansion
}

type disks struct {
	client rest.Interface
	ns     string
}

func newDisks(c *DisksClient, namespace string) *disks {
	return &disks{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

func (c *disks) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.Disk, err error) {
	result = &v1alpha1.Disk{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("disks").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

func (c *disks) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.DiskList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.DiskList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("disks").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

func (c *disks) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("disks").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

func (c *disks) Create(ctx context.Context, disk *v1alpha1.Disk, opts v1.CreateOptions) (result *v1alpha1.Disk, err error) {
	result = &v1alpha1.Disk{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("disks").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(disk).
		Do(ctx).
		Into(result)
	return
}

func (c *disks) Update(ctx context.Context, disk *v1alpha1.Disk, opts v1.UpdateOptions) (result *v1alpha1.Disk, err error) {
	result = &v1alpha1.Disk{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("disks").
		Name(disk.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(disk).
		Do(ctx).
		Into(result)
	return
}

func (c *disks) UpdateStatus(ctx context.Context, disk *v1alpha1.Disk, opts v1.UpdateOptions) (result *v1alpha1.Disk, err error) {
	result = &v1alpha1.Disk{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("disks").
		Name(disk.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(disk).
		Do(ctx).
		Into(result)
	return
}

func (c *disks) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("disks").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

func (c *disks) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("disks").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

func (c *disks) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Disk, err error) {
	result = &v1alpha1.Disk{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("disks").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
