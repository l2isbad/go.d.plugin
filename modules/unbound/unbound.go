package unbound

import (
	"github.com/netdata/go-orchestrator/module"
	"github.com/netdata/go.d.plugin/pkg/web"
)

func init() {
	creator := module.Creator{
		Create: func() module.Module { return New() },
	}

	module.Register("unbound", creator)
}

func New() *Unbound {
	config := Config{}

	return &Unbound{
		Config: config,
		charts: nil,
	}
}

type (
	Config struct {
		Address             string       `yaml:"address"`
		ConfPath            string       `yaml:"conf_path"`
		Timeout             web.Duration `yaml:"timeout"`
		web.ClientTLSConfig `yaml:",inline"`
	}
	Unbound struct {
		module.Base
		Config `yaml:",inline"`

		charts *module.Charts
	}
)

func (Unbound) Cleanup() {}

func (u *Unbound) Init() bool { return true }

func (u Unbound) Check() bool { return false }

func (u Unbound) Charts() *module.Charts { return u.charts }

func (u *Unbound) Collect() map[string]int64 {
	mx, err := u.collect()

	if err != nil {
		u.Error(err)
	}

	if len(mx) == 0 {
		return nil
	}

	return mx
}
