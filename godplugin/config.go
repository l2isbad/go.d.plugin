package godplugin

import (
	"io"
	"os"

	"gopkg.in/yaml.v2"
)

func newConfig() *config {
	return &config{
		Enabled:    true,
		DefaultRun: true,
	}
}

type config struct {
	Enabled    bool            `yaml:"enabled"`
	DefaultRun bool            `yaml:"default_run"`
	MaxProcs   int             `yaml:"max_procs"`
	Modules    map[string]bool `yaml:"modules"`
}

func (c config) isModuleEnabled(module string, explicit bool) bool {
	if run, ok := c.Modules[module]; ok {
		return run
	}
	if explicit {
		return false
	}
	return c.DefaultRun
}

func newModuleConfig() *moduleConfig {
	return &moduleConfig{
		UpdateEvery:        1,
		AutoDetectionRetry: 0,
	}
}

type moduleConfig struct {
	UpdateEvery        int                      `yaml:"update_every"`
	AutoDetectionRetry int                      `yaml:"autodetection_retry"`
	Jobs               []map[string]interface{} `yaml:"jobs"`

	name string
}

func (m *moduleConfig) updateJobs(moduleUpdateEvery, pluginUpdateEvery int) {
	if moduleUpdateEvery > 0 {
		m.UpdateEvery = moduleUpdateEvery
	}

	for _, job := range m.Jobs {
		if _, ok := job["update_every"]; !ok {
			job["update_every"] = m.UpdateEvery
		}

		if _, ok := job["autodetection_retry"]; !ok {
			job["autodetection_retry"] = m.AutoDetectionRetry
		}

		if v, ok := job["update_every"].(int); ok && v < pluginUpdateEvery {
			job["update_every"] = pluginUpdateEvery
		}
	}
}

func loadYAML(conf interface{}, filename string) error {
	file, err := os.Open(filename)
	defer file.Close()

	if err != nil {
		log.Debug("open file ", filename, ": ", err)
		return err
	}

	if err = yaml.NewDecoder(file).Decode(conf); err != nil {
		if err == io.EOF {
			log.Debug("config file is empty")
			return nil
		}
		log.Debug("read YAML ", filename, ": ", err)
		return err
	}

	return nil
}
