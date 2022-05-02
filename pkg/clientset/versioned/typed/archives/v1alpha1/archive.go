package v1alpha1

import (
	"context"
	v1alpha1 "github.com/kubeberth/berth-operator/api/v1alpha1"
	scheme "github.com/kubeberth/berth-operator/pkg/clientset/versioned/scheme"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

type ArchivesGetter interface {
	Archives(namespace string) ArchiveInterface
}

type ArchiveInterface interface {
	Create(ctx context.Context, archive *v1alpha1.Archive, opts v1.CreateOptions) (*v1alpha1.Archive, error)
	Update(ctx context.Context, archive *v1alpha1.Archive, opts v1.UpdateOptions) (*v1alpha1.Archive, error)
	UpdateStatus(ctx context.Context, archive *v1alpha1.Archive, opts v1.UpdateOptions) (*v1alpha1.Archive, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.Archive, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.ArchiveList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Archive, err error)
	ArchiveExpansion
}

type archives struct {
	client rest.Interface
	ns     string
}

func newArchives(c *ArchivesClient, namespace string) *archives {
	return &archives{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

func (c *archives) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.Archive, err error) {
	result = &v1alpha1.Archive{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("archives").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		//Do(ctx).
		Do().
		Into(result)
	return
}

func (c *archives) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.ArchiveList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.ArchiveList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("archives").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		//Do(ctx).
		Do().
		Into(result)
	return
}

func (c *archives) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("archives").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		//Watch(ctx)
		Watch()
}

func (c *archives) Create(ctx context.Context, archive *v1alpha1.Archive, opts v1.CreateOptions) (result *v1alpha1.Archive, err error) {
	result = &v1alpha1.Archive{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("archives").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(archive).
		//Do(ctx).
		Do().
		Into(result)
	return
}

func (c *archives) Update(ctx context.Context, archive *v1alpha1.Archive, opts v1.UpdateOptions) (result *v1alpha1.Archive, err error) {
	result = &v1alpha1.Archive{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("archives").
		Name(archive.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(archive).
		//Do(ctx).
		Do().
		Into(result)
	return
}

func (c *archives) UpdateStatus(ctx context.Context, archive *v1alpha1.Archive, opts v1.UpdateOptions) (result *v1alpha1.Archive, err error) {
	result = &v1alpha1.Archive{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("archives").
		Name(archives.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(archive).
		//Do(ctx).
		Do().
		Into(result)
	return
}

func (c *archives) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("archives").
		Name(name).
		Body(&opts).
		//Do(ctx).
		Do().
		Error()
}

func (c *archives) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("archives").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		//Do(ctx).
		Do().
		Error()
}

func (c *archives) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Archive, err error) {
	result = &v1alpha1.Archive{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("archives").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		//Do(ctx).
		Do().
		Into(result)
	return
}
