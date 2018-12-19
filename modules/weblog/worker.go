package weblog

import (
	"strconv"
	"strings"

	"github.com/netdata/go.d.plugin/modules"
)

func newWorker() *worker {
	return &worker{
		tailFactory: newFollower,
		reqParser: newCSVParser(csvPattern{
			{keyMethod, 0},
			{keyURL, 1},
			{keyVersion, 2},
		}),
		stopCh:  make(chan struct{}, 1),
		pauseCh: make(chan struct{}),
		timings: timings{
			keyRespTime:         &timing{},
			keyRespTimeUpstream: &timing{},
		},
		histograms:     make(map[string]histogram),
		uniqIPs:        make(map[string]bool),
		uniqIPsAllTime: make(map[string]bool),
		metrics: map[string]int64{
			"successful_requests":      0,
			"redirects":                0,
			"bad_requests":             0,
			"server_errors":            0,
			"other_requests":           0,
			"2xx":                      0,
			"5xx":                      0,
			"3xx":                      0,
			"4xx":                      0,
			"1xx":                      0,
			"0xx":                      0,
			"unmatched":                0,
			"bytes_sent":               0,
			"resp_length":              0,
			"resp_time_min":            0,
			"resp_time_max":            0,
			"resp_time_avg":            0,
			"resp_time_upstream_min":   0,
			"resp_time_upstream_max":   0,
			"resp_time_upstream_avg":   0,
			"unique_current_poll_ipv4": 0,
			"unique_current_poll_ipv6": 0,
			"unique_all_time_ipv4":     0,
			"unique_all_time_ipv6":     0,
			"req_ipv4":                 0,
			"req_ipv6":                 0,
			"GET":                      0, // GET should be green on the dashboard
		},
	}
}

type chartUpdateTask struct {
	id  string
	dim *Dim
}

type worker struct {
	webLog *WebLog

	tailFactory func(string) (follower, error)
	tail        follower
	filter      matcher
	parser      parser
	reqParser   parser

	matchedURL string // for chart per url

	urlCats  []*category
	userCats []*category

	stopCh  chan struct{}
	pauseCh chan struct{}

	timings        timings
	histograms     map[string]histogram
	uniqIPs        map[string]bool
	uniqIPsAllTime map[string]bool

	chartUpdate []chartUpdateTask

	metrics map[string]int64
}

func (w *worker) stop() {
	w.stopCh <- struct{}{}
}

func (w *worker) pause() {
	w.pauseCh <- struct{}{}
}

func (w *worker) unpause() {
	<-w.pauseCh
}

func (w *worker) cleanup() {
	w.tail.stop()
}

func (w *worker) parseLoop() {
	lines := w.tail.lines()
LOOP:
	for {
		select {
		case <-w.stopCh:
			w.cleanup()
			break LOOP
		case <-w.pauseCh:
			w.pauseCh <- struct{}{}
		case line := <-lines:
			if w.filter.match(line.Text) {
				w.parseLine(line.Text)
			}
		}
	}
}

func (w *worker) parseLine(line string) {
	gm, ok := w.parser.parse(line)

	if !ok {
		w.metrics["unmatched"]++
		return
	}

	w.codeFam(gm)

	w.codeStatus(gm)

	if w.webLog.DoCodesDetailed {
		w.codeDetailed(gm)
	}

	if _, ok := gm.lookup(keyVhost); ok {
		w.vhost(gm)
	}

	if _, ok := gm.lookup(keyRequest); ok {
		w.request(gm)
	}

	if _, ok := gm.lookup(keyUserDefined); ok && len(w.userCats) > 0 {
		w.userCategory(gm)
	}

	if _, ok := gm.lookup(keyBytesSent); ok {
		w.bytesSent(gm)
	}

	if _, ok := gm.lookup(keyRespLength); ok {
		w.respLength(gm)
	}

	if _, ok := gm.lookup(keyRespTime); ok {
		w.respTime(gm)
	}

	if _, ok := gm.lookup(keyRespTimeUpstream); ok {
		w.respTimeUpstream(gm)
	}

	if _, ok := gm.lookup(keyAddress); ok {
		w.ipProto(gm)
	}

	if w.webLog.DoPerURLCharts && w.matchedURL != "" {
		w.urlCategoryStats(gm)
	}

}

func (w *worker) vhost(gm groupMap) {
	vhost := gm.get(keyVhost)
	dimID := vhostReplacer.Replace(vhost)

	if _, ok := w.metrics[dimID]; !ok {
		dim := &Dim{ID: dimID, Name: vhost, Algo: modules.Incremental}
		w.chartUpdate = append(w.chartUpdate, chartUpdateTask{id: requestsPerVhost.ID, dim: dim})
	}

	w.metrics[dimID]++
}

func (w *worker) codeFam(gm groupMap) {
	fam := gm.get(keyCode)[:1] + "xx"

	if _, ok := w.metrics[fam]; ok {
		w.metrics[fam]++
	} else {
		w.metrics["0xx"]++
	}
}

func (w *worker) codeDetailed(gm groupMap) {
	code := gm.get(keyCode)

	if _, ok := w.metrics[code]; ok {
		w.metrics[code]++
		return
	}

	var chartID string

	if w.webLog.DoCodesAggregate {
		chartID = responseCodesDetailed.ID
	} else {
		v := "other"
		if code[0] <= 53 {
			v = code[:1] + "xx"
		}
		chartID = responseCodesDetailed.ID + "_" + v
	}
	dim := &Dim{ID: code, Algo: modules.Incremental}
	w.chartUpdate = append(w.chartUpdate, chartUpdateTask{id: chartID, dim: dim})

	w.metrics[code]++
}

func (w *worker) codeStatus(gm groupMap) {
	code, fam := gm.get(keyCode), gm.get(keyCode)[:1]

	switch {
	case fam == "2", code == "304", fam == "1":
		w.metrics["successful_requests"]++
	case fam == "3":
		w.metrics["redirects"]++
	case fam == "4":
		w.metrics["bad_requests"]++
	case fam == "5":
		w.metrics["server_errors"]++
	default:
		w.metrics["other_requests"]++
	}
}

func (w *worker) request(gm groupMap) {
	var ok bool

	if gm, ok = w.reqParser.parse(gm.get(keyRequest)); !ok {
		return
	}

	w.httpMethod(gm)
	w.urlCategory(gm)
	w.httpVersion(gm)
}

func (w *worker) httpMethod(gm groupMap) {
	method := gm.get(keyMethod)

	if _, ok := w.metrics[method]; !ok {
		dim := &Dim{ID: method, Algo: modules.Incremental}
		w.chartUpdate = append(w.chartUpdate, chartUpdateTask{id: requestsPerHTTPMethod.ID, dim: dim})
	}

	w.metrics[method]++
}

func (w *worker) urlCategory(gm groupMap) {
	url := gm.get(keyURL)

	for _, v := range w.urlCats {
		if v.match(url) {
			w.metrics[v.name]++
			w.matchedURL = v.name
			return
		}
	}
	w.matchedURL = ""
}

func (w *worker) userCategory(gm groupMap) {
	userDefined := gm.get(keyUserDefined)

	for _, cat := range w.userCats {
		if cat.match(userDefined) {
			w.metrics[cat.name]++
			return
		}
	}
}

func (w *worker) httpVersion(gm groupMap) {
	version := gm.get(keyVersion)
	dimID := httpMethodReplacer.Replace(version)

	if _, ok := w.metrics[dimID]; !ok {
		dim := &Dim{ID: dimID, Name: version, Algo: modules.Incremental}
		w.chartUpdate = append(w.chartUpdate, chartUpdateTask{id: requestsPerHTTPVersion.ID, dim: dim})
	}

	w.metrics[dimID]++
}

func (w *worker) bytesSent(gm groupMap) {
	w.metrics[keyBytesSent] += toInt(gm.get(keyBytesSent))
}

func (w *worker) respLength(gm groupMap) {
	w.metrics[keyRespLength] += toInt(gm.get(keyRespLength))
}

func (w *worker) ipProto(gm groupMap) {
	var (
		address = gm.get(keyAddress)
		proto   = "ipv4"
	)

	if strings.Contains(address, ":") {
		proto = "ipv6"
	}

	w.metrics["req_"+proto]++

	if _, ok := w.uniqIPs[address]; !ok {
		w.uniqIPs[address] = true
		w.metrics["unique_current_poll_"+proto]++
	}

	if !w.webLog.DoAllTimeIPs {
		return
	}

	if _, ok := w.uniqIPsAllTime[address]; !ok {
		w.uniqIPsAllTime[address] = true
		w.metrics["unique_all_time_"+proto]++
	}

}

func (w *worker) respTime(gm groupMap) {
	i := w.timings.get(keyRespTime).set(gm.get(keyRespTime))

	if h, ok := w.histograms[keyRespTimeHistogram]; ok {
		h.set(i)
	}
}

func (w *worker) respTimeUpstream(gm groupMap) {
	i := w.timings.get(keyRespTimeUpstream).set(gm.get(keyRespTimeUpstream))

	if h, ok := w.histograms[keyRespTimeUpstreamHistogram]; ok {
		h.set(i)
	}
}

func (w *worker) urlCategoryStats(gm groupMap) {
	code := gm.get(keyCode)
	id := w.matchedURL + "_" + code

	if _, ok := w.metrics[id]; !ok {
		dim := &Dim{ID: id, Name: code, Algo: modules.Incremental}
		w.chartUpdate = append(w.chartUpdate, chartUpdateTask{id: responseCodesDetailed.ID + "_" + w.matchedURL, dim: dim})
	}

	w.metrics[id]++

	if v, ok := gm.lookup(keyBytesSent); ok {
		w.metrics[w.matchedURL+"_bytes_sent"] += toInt(v)
	}

	if v, ok := gm.lookup(keyRespLength); ok {
		w.metrics[w.matchedURL+"_resp_length"] += toInt(v)
	}

	if v, ok := gm.lookup(keyRespTime); ok {
		w.timings.get(w.matchedURL).set(v)
	}
}

// toInt used in bytesSent and respLength
func toInt(s string) int64 {
	if s == "-" {
		return 0
	}
	v, _ := strconv.Atoi(s)

	return int64(v)
}

var (
	httpMethodReplacer = strings.NewReplacer("/", "_", ".", "_")
	vhostReplacer      = strings.NewReplacer(":", "_", ".", "_", "-", "_")
)
