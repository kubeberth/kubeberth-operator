package versioned

import (
	"fmt"
	archives "github.com/kubeberth/kubeberth-operator/pkg/clientset/versioned/typed/archives/v1alpha1"
	cloudinits "github.com/kubeberth/kubeberth-operator/pkg/clientset/versioned/typed/cloudinits/v1alpha1"
	disks "github.com/kubeberth/kubeberth-operator/pkg/clientset/versioned/typed/disks/v1alpha1"
	servers "github.com/kubeberth/kubeberth-operator/pkg/clientset/versioned/typed/servers/v1alpha1"
	loadbalancers "github.com/kubeberth/kubeberth-operator/pkg/clientset/versioned/typed/loadbalancers/v1alpha1"

	discovery "k8s.io/client-go/discovery"
	rest "k8s.io/client-go/rest"
	flowcontrol "k8s.io/client-go/util/flowcontrol"
)

type Interface interface {
	Discovery()     discovery.DiscoveryInterface
	Archives()      archives.ArchivesInterface
	CloudInits()    cloudinits.CloudInitsInterface
	Disks()         disks.DisksInterface
	Servers()       servers.ServersInterface
	LoadBalancers() loadbalancers.LoadBalancersInterface
}

// Clientset contains the clients for groups. Each group has exactly one
// version included in a Clientset.
type Clientset struct {
	*discovery.DiscoveryClient
	archives      *archives.ArchivesClient
	cloudinits    *cloudinits.CloudInitsClient
	disks         *disks.DisksClient
	servers       *servers.ServersClient
	loadbalancers *loadbalancers.LoadBalancerClient
}

// Discovery retrieves the DiscoveryClient
func (c *Clientset) Discovery() discovery.DiscoveryInterface {
	if c == nil {
		return nil
	}
	return c.DiscoveryClient
}

func (c *Clientset) Archives() archives.ArchivesInterface {
	return c.archives
}

func (c *Clientset) CloudInits() cloudinits.CloudInitsInterface {
	return c.cloudinits
}

func (c *Clientset) Disks() disks.DisksInterface {
	return c.disks
}

func (c *Clientset) Servers() servers.ServersInterface {
	return c.servers
}

func (c *Clientset) LoadBalancers() loadbalancers.LoadBalancersInterface {
	return c.loadbalancers
}
// NewForConfig creates a new Clientset for the given config.
// If config's RateLimiter is not set and QPS and Burst are acceptable,
// NewForConfig will generate a rate-limiter in configShallowCopy.
func NewForConfig(c *rest.Config) (*Clientset, error) {
	configShallowCopy := *c
	if configShallowCopy.RateLimiter == nil && configShallowCopy.QPS > 0 {
		if configShallowCopy.Burst <= 0 {
			return nil, fmt.Errorf("burst is required to be greater than 0 when RateLimiter is not set and QPS is set to greater than 0")
		}
		configShallowCopy.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(configShallowCopy.QPS, configShallowCopy.Burst)
	}

	var cs Clientset
	var err error

	cs.DiscoveryClient, err = discovery.NewDiscoveryClientForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}

	cs.archives, err = archives.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}

	cs.cloudinits, err = cloudinits.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}

	cs.disks, err = disks.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}

	cs.servers, err = servers.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}

	cs.loadbalancers, err = loadbalancers.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}

	return &cs, nil
}

// NewForConfigOrDie creates a new Clientset for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *Clientset {
	var cs Clientset
	cs.DiscoveryClient = discovery.NewDiscoveryClientForConfigOrDie(c)
	cs.archives        = archives.NewForConfigOrDie(c)
	cs.cloudinits      = cloudinits.NewForConfigOrDie(c)
	cs.disks           = disks.NewForConfigOrDie(c)
	cs.servers         = servers.NewForConfigOrDie(c)
	cs.loadbalancers   = loadbalancers.NewForConfigOrDie(c)

	return &cs
}

// New creates a new Clientset for the given RESTClient.
func New(c rest.Interface) *Clientset {
	var cs Clientset
	cs.DiscoveryClient = discovery.NewDiscoveryClient(c)
	cs.archives = archives.New(c)
	cs.cloudinits = cloudinits.New(c)
	cs.disks = disks.New(c)
	cs.servers = servers.New(c)
	cs.loadbalancers = loadbalancers.New(c)

	return &cs
}
