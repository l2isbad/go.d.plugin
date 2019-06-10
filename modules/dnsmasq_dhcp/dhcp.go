package dnsmasq_dhcp

import (
	"github.com/netdata/go-orchestrator/module"
)

func init() {
	creator := module.Creator{
		Create: func() module.Module { return New() },
	}

	module.Register("dnsmasq_dhcp", creator)
}

// New creates DnsmasqDHCP with default values.
func New() *DnsmasqDHCP {
	return &DnsmasqDHCP{
		metrics: make(map[string]int64),
	}
}

type userPool struct {
	Name      string `yaml:"name"`
	DHCPRange string `yaml:"range"`
}

// Config is the DnsmasqDHCP module configuration.
type Config struct {
	LeasesPath string     `yaml:"leases_path"`
	Pools      []userPool `yaml:"pools"`
}

// DnsmasqDHCP DnsmasqDHCP module.
type DnsmasqDHCP struct {
	module.Base
	Config `yaml:",inline"`

	metrics map[string]int64
}

// Cleanup makes cleanup.
func (DnsmasqDHCP) Cleanup() {}

// Init makes initialization.
func (DnsmasqDHCP) Init() bool { return true }

// Check makes check.
func (DnsmasqDHCP) Check() bool { return true }

// Charts creates Charts.
func (DnsmasqDHCP) Charts() *Charts { return charts.Copy() }

// Collect collects metrics.
func (d *DnsmasqDHCP) Collect() map[string]int64 {
	mx, err := d.collect()
	if err != nil {
		d.Error(err)
	}

	if len(mx) == 0 {
		return nil
	}

	return mx
}
