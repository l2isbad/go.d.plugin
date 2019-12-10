package squidlog

import (
	"github.com/netdata/go.d.plugin/pkg/logs"

	"github.com/netdata/go-orchestrator/module"
)

func init() {
	creator := module.Creator{
		Defaults: module.Defaults{
			Disabled: true,
		},
		Create: func() module.Module { return New() },
	}

	module.Register("squid_log", creator)
}

func New() *SquidLog {
	cfg := logs.ParserConfig{
		LogType: logs.TypeCSV,
		CSV: logs.CSVConfig{
			FieldsPerRecord:  -1,
			Delimiter:        ' ',
			TrimLeadingSpace: false,
			//CheckField:       checkCSVFormatField,
		},
		LTSV: logs.LTSVConfig{
			FieldDelimiter: '\t',
			ValueDelimiter: ':',
		},
		RegExp: logs.RegExpConfig{},
	}
	return &SquidLog{
		Config: Config{
			ExcludePath:    "*.gz",
			GroupRespCodes: true,
			Parser:         cfg,
		},
	}
}

type (
	Config struct {
		Parser         logs.ParserConfig `yaml:",inline"`
		Path           string            `yaml:"path"`
		ExcludePath    string            `yaml:"exclude_path"`
		GroupRespCodes bool              `yaml:"group_response_codes"`
	}

	SquidLog struct {
		module.Base
		Config `yaml:",inline"`

		file   *logs.Reader
		parser logs.Parser
		line   *logLine

		mx     *metricsData
		charts *module.Charts
	}
)

func (s *SquidLog) Init() bool {
	s.createLogLine()
	s.mx = newMetricsData(s.Config)
	return true
}

func (s *SquidLog) Check() bool {
	// Note: these inits are here to make auto detection retry working
	if err := s.createLogReader(); err != nil {
		s.Warning("check failed: ", err)
		return false
	}

	if err := s.createParser(); err != nil {
		s.Warning("check failed: ", err)
		return false
	}

	if err := s.createCharts(s.line); err != nil {
		s.Warning("check failed: ", err)
	}
	return true
}

func (s *SquidLog) Charts() *module.Charts {
	return s.charts
}

func (s *SquidLog) Collect() map[string]int64 {
	mx, err := s.collect()
	if err != nil {
		s.Error(err)
	}

	if len(mx) == 0 {
		return nil
	}
	return mx
}

func (s *SquidLog) Cleanup() {
	if s.file != nil {
		_ = s.file.Close()
	}
}