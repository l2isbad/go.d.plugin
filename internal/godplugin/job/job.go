package job

import (
	"sync"
	"time"

	"github.com/l2isbad/go.d.plugin/internal/modules"
	"github.com/l2isbad/go.d.plugin/internal/pkg/logger"
)

type Job struct {
	*Config
	*logger.Logger
	Module       modules.Module
	Tick         chan int
	shutdownHook chan int
	observer     *observer

	timers
	retries int
}

func New(module modules.Module, config *Config) *Job {
	return &Job{
		Module:       module,
		Config:       config,
		Tick:         make(chan int),
		shutdownHook: make(chan int),
		observer:     newObserver(config),
	}
}

func (j *Job) Init() error {
	l := logger.New(j.RealModuleName, j.JobName())
	j.Logger = l

	j.Module.SetUpdateEvery(j.UpdateEvery)
	j.Module.SetModuleName(j.RealModuleName)
	j.Module.SetLogger(l)

	return j.Module.Init()
}

func (j *Job) Check() bool {
	okCh := make(chan bool)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				j.Errorf("PANIC: %v", r)
				okCh <- false
			}
		}()
		okCh <- j.Module.Check()
	}()

	var ok bool
	select {
	case ok = <-okCh:
	case time.After(5 * time.Second):
		j.Error("check timeout")
	}
	return ok
}

func (j *Job) MainLoop() {
LOOP:
	for {
		select {
		case <-j.shutdownHook:
			break LOOP
		case t := <-j.Tick:
			if t%j.UpdateEvery != 0 {
				continue LOOP
			}
		}
		data := j.getData()
		if data == nil {
			continue
		}
	}
}

func (j *Job) Start(wg *sync.WaitGroup) {
Done:
	for {

		sleep := j.nextIn()
		j.Debugf("sleeping for %s to reach frequency of %d sec", sleep, j.UpdateEvery)
		time.Sleep(sleep)

		j.curRun = time.Now()
		if !j.lastRun.IsZero() {
			j.sinceLast.Duration = j.curRun.Sub(j.lastRun)
		}

		if ok := j.update(); ok {
			j.retries, j.penalty, j.lastRun = 0, 0, j.curRun
			j.spentOnRun.Duration = time.Since(j.lastRun)

		} else if !ok && !j.handleRetries() {
			j.Errorf("stopped after %d collection failures in a row", j.MaxRetries)
			break Done
		}

	}
	wg.Done()
}

func (j *Job) update() bool {

	data := j.getData()

	if data == nil {
		j.Debug("getData() failed")
		return false
	}

	var (
		updated    int
		active     int
		suppressed int
	)

	for _, v := range *j.observer.charts {
		if _, ok := j.observer.items[v.ID]; !ok {
			j.observer.add(v)
		}
		chart := j.observer.items[v.ID]

		if chart.obsoleted {
			if canBeUpdated(*chart, data) {
				chart.refresh()
			} else {
				suppressed++
				continue
			}
		}

		if j.ChartCleanup > 0 && chart.retries >= j.ChartCleanup {
			j.Errorf("item '%s' was suppressed due to non updating", chart.item.ID)
			chart.obsolete()
			suppressed++
			continue
		}

		active++
		if chart.update(data, j.sinceLast.ConvertTo(time.Microsecond)) {
			updated++
		}
	}

	j.Debugf("update items: updated:%d, active:%d, suppressed:%d", updated, active, suppressed)
	return updated > 0
}

func (j *Job) getData() (result map[string]int64) {
	defer func() {
		if r := recover(); r != nil {
			j.Errorf("PANIC: %v", r)
			result = nil
		}
	}()
	result = j.Module.GetData()
	return
}

func (j *Job) nextIn() time.Duration {
	start := time.Now()
	next := start.Add(time.Duration(j.UpdateEvery) * time.Second).Add(j.penalty).Truncate(time.Second)
	return time.Duration(next.UnixNano() - start.UnixNano())
}

func (j *Job) handleRetries() bool {
	j.retries++

	if j.retries%5 != 0 {
		return true
	}

	j.penalty = time.Duration(j.retries*j.UpdateEvery/2) * time.Second
	j.Warningf(
		"added %.0f seconds penalty after %d failed updates in a row",
		j.penalty.Seconds(),
		j.retries,
	)
	return j.retries < j.MaxRetries
}
