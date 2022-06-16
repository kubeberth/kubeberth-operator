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

type ServersGetter interface {
	Servers(namespace string) ServerInterface
}

type ServerInterface interface {
	Create(ctx context.Context, server *v1alpha1.Server, opts v1.CreateOptions) (*v1alpha1.Server, error)
	Update(ctx context.Context, server *v1alpha1.Server, opts v1.UpdateOptions) (*v1alpha1.Server, error)
	UpdateStatus(ctx context.Context, server *v1alpha1.Server, opts v1.UpdateOptions) (*v1alpha1.Server, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.Server, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.ServerList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Server, err error)
	ServerExpansion
}

type servers struct {
	client rest.Interface
	ns     string
}

func newServers(c *ServersClient, namespace string) *servers {
	return &servers{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

func (c *servers) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.Server, err error) {
	result = &v1alpha1.Server{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("servers").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

func (c *servers) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.ServerList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.ServerList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("servers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

func (c *servers) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("servers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

func (c *servers) Create(ctx context.Context, server *v1alpha1.Server, opts v1.CreateOptions) (result *v1alpha1.Server, err error) {
	result = &v1alpha1.Server{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("servers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(server).
		Do(ctx).
		Into(result)
	return
}

func (c *servers) Update(ctx context.Context, server *v1alpha1.Server, opts v1.UpdateOptions) (result *v1alpha1.Server, err error) {
	result = &v1alpha1.Server{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("servers").
		Name(server.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(server).
		Do(ctx).
		Into(result)
	return
}

func (c *servers) UpdateStatus(ctx context.Context, server *v1alpha1.Server, opts v1.UpdateOptions) (result *v1alpha1.Server, err error) {
	result = &v1alpha1.Server{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("servers").
		Name(server.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(server).
		Do(ctx).
		Into(result)
	return
}

func (c *servers) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("servers").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

func (c *servers) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("servers").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

func (c *servers) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Server, err error) {
	result = &v1alpha1.Server{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("servers").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
