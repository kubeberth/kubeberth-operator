package v1alpha1

import (
	berth "github.com/kubeberth/berth-operator/api/v1alpha1"
	"github.com/kubeberth/berth-operator/pkg/clientset/versioned/scheme"

	rest "k8s.io/client-go/rest"
)

type ServersInterface interface {
	RESTClient() rest.Interface
	ServersGetter
}

type ServersClient struct {
	restClient rest.Interface
}

func (c *ServersClient) Servers(namespace string) ServerInterface {
	return newServers(c, namespace)
}

func NewForConfig(c *rest.Config) (*ServersClient, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &ServersClient{client}, nil
}

func NewForConfigOrDie(c *rest.Config) *ServersClient {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

func New(c rest.Interface) *ServersClient {
	return &ServersClient{c}
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
func (c *ServersClient) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
