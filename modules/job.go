package modules

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/l2isbad/go.d.plugin/logger"
)

type JobConfig struct {
	JobName            string `yaml:"job_name"`
	OverrideName       string `yaml:"name"`
	UpdateEvery        int    `yaml:"update_every" validate:"gte=1"`
	AutoDetectionRetry int    `yaml:"autodetection_retry" validate:"gte=0"`
	MaxRetries         int    `yaml:"Retries" validate:"gte=0"`
}

// JobNewConfig returns JobConfig with default values
func JobNewConfig() *JobConfig {
	return &JobConfig{
		UpdateEvery:        1,
		AutoDetectionRetry: 0,
		MaxRetries:         60,
	}
}

type Job interface {
	ModuleName() string
	Name() string
	AutoDetectionRetry() int
	Panicked() bool
	Inited() bool

	Init() bool
	Check() bool
	PostCheck() bool

	Tick(clock int)
	MainLoop()
	Shutdown()
}

func NewJob(modName string, module Module, config *JobConfig, out io.Writer) Job {
	buf := &bytes.Buffer{}
	return &job{
		moduleName:   modName,
		module:       module,
		JobConfig:    config,
		out:          out,
		tick:         make(chan int),
		shutdownHook: make(chan struct{}),
		buf:          buf,
		priority:     70000,
		apiWriter:    apiWriter{Writer: buf},
	}
}

type job struct {
	*JobConfig
	*logger.Logger
	module Module

	moduleName string
	inited     bool
	panicked   bool

	charts       *Charts
	tick         chan int
	shutdownHook chan struct{}
	out          io.Writer
	buf          *bytes.Buffer
	apiWriter    apiWriter

	priority int
	retries  int
	prevRun  time.Time
}

func (j job) fullName() string {
	if j.ModuleName() == j.Name() {
		return j.ModuleName()
	}
	return fmt.Sprintf("%s_%s", j.ModuleName(), j.Name())
}

func (j job) ModuleName() string {
	return j.moduleName
}

func (j job) Name() string {
	if j.OverrideName != "" {
		return j.OverrideName
	}
	if j.JobName != "" {
		return j.JobName
	}
	return j.ModuleName()
}

func (j job) Inited() bool {
	return j.inited
}

func (j job) Panicked() bool {
	return j.panicked
}

func (j *job) Init() bool {
	j.Logger = logger.New(j.ModuleName(), j.Name())
	j.module.SetUpdateEvery(j.UpdateEvery)
	j.module.SetModuleName(j.ModuleName())
	j.module.SetLogger(j.Logger)

	return j.module.Init()
}

func (j *job) Check() bool {
	defer func() {
		if r := recover(); r != nil {
			j.panicked = true
			j.Errorf("PANIC %v", r)
		}

	}()
	return j.module.Check()
}

func (j *job) PostCheck() bool {
	j.UpdateEvery = j.module.UpdateEvery()
	j.moduleName = j.module.ModuleName()
	logger.SetModName(j.Logger, j.ModuleName())

	charts := j.module.GetCharts()
	if charts == nil {
		j.Error("GetCharts() [FAILED]")
		return false
	}

	j.charts = charts
	return true
}

func (j *job) Tick(clock int) {
	select {
	case j.tick <- clock:
	default:
		j.Errorf("Skip the tick due to previous run hasn't been finished.")
	}
}

func (j *job) MainLoop() {
LOOP:
	for {
		select {
		case <-j.shutdownHook:
			break LOOP
		case t := <-j.tick:
			if t%j.UpdateEvery != 0 {
				continue LOOP
			}

			curTime := time.Now()
			sinceLast := calcSinceLast(curTime, j.prevRun)
			data := j.getData()

			//if j.Panicked {
			//
			//}

			if !j.populateMetrics(data, sinceLast) {
				j.retries++
			} else {
				j.prevRun = curTime
			}

			io.Copy(j.out, j.buf)
			j.buf.Reset()
		}
	}
}

func (j *job) Shutdown() {
	select {
	case j.shutdownHook <- struct{}{}:
		j.module.Cleanup()
	default:
	}
}

func (j job) AutoDetectionRetry() int {
	return j.JobConfig.AutoDetectionRetry
}

func (j *job) getData() (result map[string]int64) {
	defer func() {
		if r := recover(); r != nil {
			j.Errorf("PANIC: %v", r)
			j.panicked = true
		}
	}()
	return j.module.GetData()
}

func (j *job) populateMetrics(data map[string]int64, sinceLast int) bool {
	var totalUpdated int

	for _, chart := range *j.charts {
		if chart.Priority == 0 {
			chart.Priority = j.priority
			j.priority++
		}

		if !chart.pushed {
			j.apiWriter.chart(
				j.fullName(),
				chart.ID,
				chart.OverID,
				chart.Title,
				chart.Units,
				chart.Fam,
				chart.Ctx,
				chart.Type,
				chart.Priority,
				j.UpdateEvery,
				chart.Opts,
				j.moduleName,
			)
			chart.pushed = true
		}

		if data == nil || chart.Obsolete {
			continue
		}

		if j.updateChart(chart, data, sinceLast) {
			totalUpdated++
		}
	}

	return totalUpdated > 0
}

func (j *job) updateChart(chart *Chart, data map[string]int64, sinceLast int) bool {
	if !chart.updated {
		sinceLast = 0
	}

	j.apiWriter.begin(j.fullName(), chart.ID, sinceLast)

	var updated int

	for _, dim := range chart.Dims {
		if v, ok := data[dim.ID]; ok {
			j.apiWriter.set(dim.ID, v)
			updated++
		}
	}
	for _, variable := range chart.Vars {
		if v, ok := data[variable.ID]; ok {
			j.apiWriter.set(variable.ID, v)
		}
	}

	j.apiWriter.end()
	chart.updated = updated > 0

	if chart.updated {
		chart.Retries = 0
	} else {
		chart.Retries++
	}

	return chart.updated
}

func calcSinceLast(curTime, prevRun time.Time) int {
	if prevRun.IsZero() {
		return 0
	}
	return int(int64(curTime.Sub(prevRun)) / (int64(time.Microsecond) / int64(time.Nanosecond)))

}

func calcPenalty(retries, updateEvery int) int {
	if retries%5 == 0 {
		return retries * updateEvery / 2
	}
	return 0
}
