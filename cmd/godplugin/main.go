package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/netdata/go-orchestrator"
	"github.com/netdata/go-orchestrator/cli"
	"github.com/netdata/go-orchestrator/logger"
	"github.com/netdata/go-orchestrator/pkg/multipath"

	_ "github.com/netdata/go.d.plugin/modules/activemq"
	_ "github.com/netdata/go.d.plugin/modules/apache"
	_ "github.com/netdata/go.d.plugin/modules/bind"
	_ "github.com/netdata/go.d.plugin/modules/consul"
	_ "github.com/netdata/go.d.plugin/modules/dnsquery"
	_ "github.com/netdata/go.d.plugin/modules/example"
	_ "github.com/netdata/go.d.plugin/modules/fluentd"
	_ "github.com/netdata/go.d.plugin/modules/freeradius"
	_ "github.com/netdata/go.d.plugin/modules/httpcheck"
	_ "github.com/netdata/go.d.plugin/modules/lighttpd"
	_ "github.com/netdata/go.d.plugin/modules/lighttpd2"
	_ "github.com/netdata/go.d.plugin/modules/logstash"
	_ "github.com/netdata/go.d.plugin/modules/mysql"
	_ "github.com/netdata/go.d.plugin/modules/nginx"
	_ "github.com/netdata/go.d.plugin/modules/portcheck"
	_ "github.com/netdata/go.d.plugin/modules/rabbitmq"
	_ "github.com/netdata/go.d.plugin/modules/solr"
	_ "github.com/netdata/go.d.plugin/modules/springboot2"
	_ "github.com/netdata/go.d.plugin/modules/weblog"
	_ "github.com/netdata/go.d.plugin/modules/x509check"
)

var (
	cd, _       = os.Getwd()
	configPaths = multipath.New(
		os.Getenv("NETDATA_USER_CONFIG_DIR"),
		os.Getenv("NETDATA_STOCK_CONFIG_DIR"),
		path.Join(cd, "/../../../../etc/netdata"),
		path.Join(cd, "/../../../../usr/lib/netdata/conf.d"),
	)
)

var version = "unknown"

func main() {
	opt := parseCLI()

	if opt.Version {
		fmt.Println(fmt.Sprintf("go.d.plugin, version: %s", version))
		return
	}
	if opt.Debug {
		logger.SetSeverity(logger.DEBUG)
	}

	plugin := newPlugin(opt)

	if !plugin.Setup() {
		return
	}

	plugin.Serve()
}

func newPlugin(opt *cli.Option) *orchestrator.Orchestrator {
	plugin := orchestrator.New()
	plugin.Name = "go.d"
	plugin.Option = opt
	plugin.ConfigPath = configPaths
	return plugin
}

func parseCLI() *cli.Option {
	opt, err := cli.Parse(os.Args)
	if err != nil {
		if err != flag.ErrHelp {
			os.Exit(1)
		}
		os.Exit(0)
	}
	return opt
}
