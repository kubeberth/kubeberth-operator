package v1alpha1

import (
	berth "github.com/kubeberth/kubeberth-operator/api/v1alpha1"
	"github.com/kubeberth/kubeberth-operator/pkg/clientset/versioned/scheme"

	rest "k8s.io/client-go/rest"
)

type LoadBalancersInterface interface {
	RESTClient() rest.Interface
	LoadBalancersGetter
}

type LoadBalancersClient struct {
	restClient rest.Interface
}

func (c *LoadBalancersClient) LoadBalancers(namespace string) LoadBalancerInterface {
	return newLoadBalancers(c, namespace)
}

func NewForConfig(c *rest.Config) (*LoadBalancersClient, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &LoadBalancersClient{client}, nil
}

func NewForConfigOrDie(c *rest.Config) *LoadBalancersClient {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

func New(c rest.Interface) *LoadBalancersClient {
	return &LoadBalancersClient{c}
}

func setConfigDefaults(config *rest.Config) error {
	gv := berth.GroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *LoadBalancersClient) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
