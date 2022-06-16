package v1alpha1

import (
	berth "github.com/kubeberth/kubeberth-operator/api/v1alpha1"
	"github.com/kubeberth/kubeberth-operator/pkg/clientset/versioned/scheme"

	rest "k8s.io/client-go/rest"
)

type ArchivesInterface interface {
	RESTClient() rest.Interface
	ArchivesGetter
}

type ArchivesClient struct {
	restClient rest.Interface
}

func (c *ArchivesClient) Archives(namespace string) ArchiveInterface {
	return newArchives(c, namespace)
}

func NewForConfig(c *rest.Config) (*ArchivesClient, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &ArchivesClient{client}, nil
}

func NewForConfigOrDie(c *rest.Config) *ArchivesClient {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

func New(c rest.Interface) *ArchivesClient {
	return &ArchivesClient{c}
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
func (c *ArchivesClient) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
