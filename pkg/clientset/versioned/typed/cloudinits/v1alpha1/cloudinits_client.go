package v1alpha1

import (
	berth "github.com/kubeberth/berth-operator/api/v1alpha1"
	"github.com/kubeberth/berth-operator/pkg/clientset/versioned/scheme"

	rest "k8s.io/client-go/rest"
)

type CloudInitsInterface interface {
	RESTClient() rest.Interface
	CloudInitsGetter
}

type CloudInitsClient struct {
	restClient rest.Interface
}

func (c *CloudInitsClient) CloudInits(namespace string) CloudInitInterface {
	return newCloudInits(c, namespace)
}

func NewForConfig(c *rest.Config) (*CloudInitsClient, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &CloudInitsClient{client}, nil
}

func NewForConfigOrDie(c *rest.Config) *CloudInitsClient {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

func New(c rest.Interface) *CloudInitsClient {
	return &CloudInitsClient{c}
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
func (c *CloudInitsClient) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
