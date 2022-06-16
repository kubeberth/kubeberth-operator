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

type CloudInitsGetter interface {
	CloudInits(namespace string) CloudInitInterface
}

type CloudInitInterface interface {
	Create(ctx context.Context, cloudInit *v1alpha1.CloudInit, opts v1.CreateOptions) (*v1alpha1.CloudInit, error)
	Update(ctx context.Context, cloudInit *v1alpha1.CloudInit, opts v1.UpdateOptions) (*v1alpha1.CloudInit, error)
	UpdateStatus(ctx context.Context, cloudInit *v1alpha1.CloudInit, opts v1.UpdateOptions) (*v1alpha1.CloudInit, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.CloudInit, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.CloudInitList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.CloudInit, err error)
	CloudInitExpansion
}

type cloudInits struct {
	client rest.Interface
	ns     string
}

func newCloudInits(c *CloudInitsClient, namespace string) *cloudInits {
	return &cloudInits{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

func (c *cloudInits) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.CloudInit, err error) {
	result = &v1alpha1.CloudInit{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("cloudinits").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

func (c *cloudInits) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.CloudInitList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.CloudInitList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("cloudinits").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

func (c *cloudInits) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("cloudinits").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

func (c *cloudInits) Create(ctx context.Context, cloudInit *v1alpha1.CloudInit, opts v1.CreateOptions) (result *v1alpha1.CloudInit, err error) {
	result = &v1alpha1.CloudInit{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("cloudinits").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(cloudInit).
		Do(ctx).
		Into(result)
	return
}

func (c *cloudInits) Update(ctx context.Context, cloudInit *v1alpha1.CloudInit, opts v1.UpdateOptions) (result *v1alpha1.CloudInit, err error) {
	result = &v1alpha1.CloudInit{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("cloudinits").
		Name(cloudInit.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(cloudInit).
		Do(ctx).
		Into(result)
	return
}

func (c *cloudInits) UpdateStatus(ctx context.Context, cloudInit *v1alpha1.CloudInit, opts v1.UpdateOptions) (result *v1alpha1.CloudInit, err error) {
	result = &v1alpha1.CloudInit{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("cloudinits").
		Name(cloudInit.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(cloudInit).
		Do(ctx).
		Into(result)
	return
}

func (c *cloudInits) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("cloudinits").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

func (c *cloudInits) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("cloudinits").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

func (c *cloudInits) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.CloudInit, err error) {
	result = &v1alpha1.CloudInit{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("cloudinits").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
