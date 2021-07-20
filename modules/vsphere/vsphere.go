package vsphere

import (
	"errors"
	"sync"
	"time"

	"github.com/netdata/go.d.plugin/modules/vsphere/client"
	"github.com/netdata/go.d.plugin/modules/vsphere/discover"
	"github.com/netdata/go.d.plugin/modules/vsphere/match"
	rs "github.com/netdata/go.d.plugin/modules/vsphere/resources"
	"github.com/netdata/go.d.plugin/modules/vsphere/scrape"
	"github.com/netdata/go.d.plugin/pkg/web"

	"github.com/netdata/go.d.plugin/agent/module"
	"github.com/vmware/govmomi/performance"
)

func init() {
	creator := module.Creator{
		Defaults: module.Defaults{
			UpdateEvery: 20,
		},
		Create: func() module.Module { return New() },
	}

	module.Register("vsphere", creator)
}

func New() *VSphere {
	config := Config{
		HTTP: web.HTTP{
			Client: web.Client{
				Timeout: web.Duration{Duration: time.Second * 20},
			},
		},
		DiscoveryInterval: web.Duration{Duration: time.Minute * 5},
		HostsInclude:      []string{"/*"},
		VMsInclude:        []string{"/*"},
		DatastoresInclude: []string{"/*"},
	}

	return &VSphere{
		collectionLock:  new(sync.RWMutex),
		Config:          config,
		charts:          &Charts{},
		discoveredHosts: make(map[string]int),
		discoveredVMs:   make(map[string]int),
		discoveredDatastores: make(map[string]int),
		charted:         make(map[string]bool),
	}
}

type discoverer interface {
	Discover() (*rs.Resources, error)
}

type scraper interface {
	ScrapeHosts(rs.Hosts) []performance.EntityMetric
	ScrapeVMs(rs.VMs) []performance.EntityMetric
	ScrapeDatastores(rs.Datastores) []performance.EntityMetric
}

type Config struct {
	web.HTTP          `yaml:",inline"`
	DiscoveryInterval web.Duration                                   `yaml:"discovery_interval"`
	HostsInclude      match.HostIncludes                             `yaml:"host_include"`
	VMsInclude        match.VMIncludes                               `yaml:"vm_include"`
	DatastoresInclude match.DatastoreIncludes                        `yaml:"datastore_include"`
	HostMetrics       struct{ Name, Cluster, DataCenter bool }       `yaml:"host_metrics"`
	VMMetrics         struct{ Name, Host, Cluster, DataCenter bool } `yaml:"vm_metrics"`
	DatastoreMetrics  struct{ Name, Cluster, DataCenter bool}        `yaml:"datastore_metrics"`
}

type VSphere struct {
	module.Base
	UpdateEvery int `yaml:"update_every"`
	Config      `yaml:",inline"`

	discoverer
	scraper

	collectionLock  *sync.RWMutex
	resources       *rs.Resources
	discoveryTask   *task
	discoveredHosts map[string]int
	discoveredVMs   map[string]int
	discoveredDatastores map[string]int
	charted         map[string]bool
	charts          *Charts
}

func (vs *VSphere) Cleanup() {
	if vs.discoveryTask == nil {
		return
	}
	vs.discoveryTask.stop()
}

func (vs VSphere) createClient() (*client.Client, error) {
	config := client.Config{
		URL:       vs.URL,
		User:      vs.Username,
		Password:  vs.Password,
		Timeout:   vs.Timeout.Duration,
		TLSConfig: vs.Client.TLSConfig,
	}
	return client.New(config)
}

func (vs *VSphere) createDiscoverer(c *client.Client) error {
	d := discover.New(c)
	d.Logger = vs.Logger

	hm, err := vs.HostsInclude.Parse()
	if err != nil {
		return err
	}
	if hm != nil {
		d.HostMatcher = hm
	}
	vmm, err := vs.VMsInclude.Parse()
	if err != nil {
		return err
	}
	if vmm != nil {
		d.VMMatcher = vmm
	}
	dm, err := vs.DatastoresInclude.Parse()
	if err!= nil {
		return err
	}
	if dm != nil {
		d.DatastoreMatcher = dm
	}

	vs.discoverer = d
	return nil
}

func (vs *VSphere) createScraper(c *client.Client) {
	ms := scrape.New(c)
	ms.Logger = vs.Logger
	vs.scraper = ms
}

const (
	minRecommendedUpdateEvery = 20
)

func (vs VSphere) validateConfig() error {
	if vs.URL == "" {
		return errors.New("URL is not set")
	}
	if vs.Username == "" || vs.Password == "" {
		return errors.New("username or password not set")
	}
	if vs.UpdateEvery < minRecommendedUpdateEvery {
		vs.Warningf("update_every is to low, minimum recommended is %d", minRecommendedUpdateEvery)
	}
	return nil
}

func (vs *VSphere) Init() bool {
	if err := vs.validateConfig(); err != nil {
		vs.Errorf("error on validating config: %v", err)
		return false
	}

	c, err := vs.createClient()
	if err != nil {
		vs.Errorf("error on creating vsphere client: %v", err)
		return false
	}

	err = vs.createDiscoverer(c)
	if err != nil {
		vs.Errorf("error on creating vsphere discoverer: %v", err)
		return false
	}

	vs.createScraper(c)

	err = vs.discoverOnce()
	if err != nil {
		vs.Errorf("error on discovering: %v", err)
		return false
	}
	vs.goDiscovery()

	return true
}

func (vs VSphere) Check() bool {
	return true
}

func (vs VSphere) Charts() *Charts {
	return vs.charts
}

func (vs *VSphere) Collect() map[string]int64 {
	mx, err := vs.collect()
	if err != nil {
		vs.Error(err)
	}

	if len(mx) == 0 {
		return nil
	}
	return mx
}
