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

type ISOImagesGetter interface {
	ISOImages(namespace string) ISOImageInterface
}

type ISOImageInterface interface {
	Create(ctx context.Context, isoimage *v1alpha1.ISOImage, opts v1.CreateOptions) (*v1alpha1.ISOImage, error)
	Update(ctx context.Context, isoimage *v1alpha1.ISOImage, opts v1.UpdateOptions) (*v1alpha1.ISOImage, error)
	UpdateStatus(ctx context.Context, isoimage *v1alpha1.ISOImage, opts v1.UpdateOptions) (*v1alpha1.ISOImage, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.ISOImage, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.ISOImageList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.ISOImage, err error)
	ISOImageExpansion
}

type isoimages struct {
	client rest.Interface
	ns     string
}

func newISOImages(c *ISOImagesClient, namespace string) *isoimages {
	return &isoimages{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

func (c *isoimages) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.ISOImage, err error) {
	result = &v1alpha1.ISOImage{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("isoimages").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

func (c *isoimages) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.ISOImageList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.ISOImageList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("isoimages").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

func (c *isoimages) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("isoimages").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

func (c *isoimages) Create(ctx context.Context, isoimage *v1alpha1.ISOImage, opts v1.CreateOptions) (result *v1alpha1.ISOImage, err error) {
	result = &v1alpha1.ISOImage{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("isoimages").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(isoimage).
		Do(ctx).
		Into(result)
	return
}

func (c *isoimages) Update(ctx context.Context, isoimage *v1alpha1.ISOImage, opts v1.UpdateOptions) (result *v1alpha1.ISOImage, err error) {
	result = &v1alpha1.ISOImage{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("isoimages").
		Name(isoimage.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(isoimage).
		Do(ctx).
		Into(result)
	return
}

func (c *isoimages) UpdateStatus(ctx context.Context, isoimage *v1alpha1.ISOImage, opts v1.UpdateOptions) (result *v1alpha1.ISOImage, err error) {
	result = &v1alpha1.ISOImage{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("isoimages").
		Name(isoimage.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(isoimage).
		Do(ctx).
		Into(result)
	return
}

func (c *isoimages) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("isoimages").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

func (c *isoimages) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("isoimages").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

func (c *isoimages) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.ISOImage, err error) {
	result = &v1alpha1.ISOImage{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("isoimages").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
