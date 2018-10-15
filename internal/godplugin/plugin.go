package godplugin

import (
	"fmt"
	"io/ioutil"
	"path"
	"runtime"
	"sync"

	"gopkg.in/yaml.v2"

	_ "github.com/l2isbad/go.d.plugin/internal/modules/all"
	"github.com/l2isbad/go.d.plugin/internal/pkg/cli"
	"github.com/l2isbad/go.d.plugin/internal/pkg/logger"
)

var (
	pluginConf = "go.d.conf"
	modConfDir = "go.d/"
)

type P interface {
	Start()
}

type dir struct {
	pluginConf  string
	modulesConf string
}

func New(p string) P {
	return &goDPlugin{
		dir:  dir{p, path.Join(p, modConfDir)},
		conf: newConfig(),
		cmd:  cli.Parse(),
	}

}

type goDPlugin struct {
	dir  dir
	conf config
	cmd  cli.ParsedCMD
	wg   sync.WaitGroup
}

func (gd *goDPlugin) Start() {
	err := gd.loadConfig()

	if err != nil {
		log.Critical(err)
	}

	if !gd.conf.Enabled {
		log.Info("disabled in configuration file")
		return
	}

	if gd.cmd.Debug {
		logger.SetLevel(logger.DEBUG)
	}

	if gd.conf.MaxProcs != 0 {
		log.Warningf("setting GOMAXPROCS to %d", gd.conf.MaxProcs)
		runtime.GOMAXPROCS(gd.conf.MaxProcs)
	}

	gd.jobsStart(gd.jobsCheck(gd.jobsInit(gd.jobsCreate())))

	gd.wg.Wait()
	fmt.Println("DISABLE")
}

func (gd *goDPlugin) loadConfig() error {
	f, err := ioutil.ReadFile(path.Join(gd.dir.pluginConf, pluginConf))

	if err != nil {
		log.Error(err)
		return nil
	}

	return yaml.Unmarshal(f, &gd.conf)
}
