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

type LoadBalancersGetter interface {
	LoadBalancers(namespace string) LoadBalancerInterface
}

type LoadBalancerInterface interface {
	Create(ctx context.Context, loadbalancer *v1alpha1.LoadBalancer, opts v1.CreateOptions) (*v1alpha1.LoadBalancer, error)
	Update(ctx context.Context, loadbalancer *v1alpha1.LoadBalancer, opts v1.UpdateOptions) (*v1alpha1.LoadBalancer, error)
	UpdateStatus(ctx context.Context, loadbalancer *v1alpha1.LoadBalancer, opts v1.UpdateOptions) (*v1alpha1.LoadBalancer, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.LoadBalancer, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.LoadBalancerList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.LoadBalancer, err error)
	LoadBalancerExpansion
}

type loadbalancers struct {
	client rest.Interface
	ns     string
}

func newLoadBalancers(c *LoadBalancersClient, namespace string) *loadbalancers {
	return &loadbalancers{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

func (c *loadbalancers) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.LoadBalancer, err error) {
	result = &v1alpha1.LoadBalancer{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("loadbalancers").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

func (c *loadbalancers) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.LoadBalancerList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.LoadBalancerList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("loadbalancers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

func (c *loadbalancers) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("loadbalancers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

func (c *loadbalancers) Create(ctx context.Context, loadbalancer *v1alpha1.LoadBalancer, opts v1.CreateOptions) (result *v1alpha1.LoadBalancer, err error) {
	result = &v1alpha1.LoadBalancer{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("loadbalancers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(loadbalancer).
		Do(ctx).
		Into(result)
	return
}

func (c *loadbalancers) Update(ctx context.Context, loadbalancer *v1alpha1.LoadBalancer, opts v1.UpdateOptions) (result *v1alpha1.LoadBalancer, err error) {
	result = &v1alpha1.LoadBalancer{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("loadbalancers").
		Name(loadbalancer.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(loadbalancer).
		Do(ctx).
		Into(result)
	return
}

func (c *loadbalancers) UpdateStatus(ctx context.Context, loadbalancer *v1alpha1.LoadBalancer, opts v1.UpdateOptions) (result *v1alpha1.LoadBalancer, err error) {
	result = &v1alpha1.LoadBalancer{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("loadbalancers").
		Name(loadbalancer.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(loadbalancer).
		Do(ctx).
		Into(result)
	return
}

func (c *loadbalancers) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("loadbalancers").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

func (c *loadbalancers) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("loadbalancers").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

func (c *loadbalancers) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.LoadBalancer, err error) {
	result = &v1alpha1.LoadBalancer{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("loadbalancers").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
