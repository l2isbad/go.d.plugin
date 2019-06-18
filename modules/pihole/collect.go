package pihole

import (
	"sync"
)

func (p *Pihole) collect() (map[string]int64, error) {
	pmx := p.scrapePihole(true)
	mx := make(map[string]int64)

	// non auth
	collectSummary(mx, pmx)
	// auth
	collectQueryTypes(mx, pmx)
	collectForwardDestination(mx, pmx)
	collectTopClients(mx, pmx)
	collectTopDomains(mx, pmx)
	collectTopAdvertisers(mx, pmx)

	p.updateCharts(pmx)

	return mx, nil
}

func collectSummary(mx map[string]int64, pmx *piholeMetrics) {
	if !pmx.hasSummary() {
		return
	}
}

func collectQueryTypes(mx map[string]int64, pmx *piholeMetrics) {
	if !pmx.hasQueryTypes() {
		return
	}

	mx["A"] = int64(pmx.queryTypes.A * 100)
	mx["AAAA"] = int64(pmx.queryTypes.AAAA * 100)
	mx["ANY"] = int64(pmx.queryTypes.ANY * 100)
	mx["PTR"] = int64(pmx.queryTypes.PTR * 100)
	mx["SOA"] = int64(pmx.queryTypes.SOA * 100)
	mx["SRV"] = int64(pmx.queryTypes.SRV * 100)
	mx["TXT"] = int64(pmx.queryTypes.TXT * 100)
}

func collectForwardDestination(mx map[string]int64, pmx *piholeMetrics) {
	if !pmx.hasForwardDestinations() {
		return
	}
	for _, v := range *pmx.forwarders {
		mx["target_"+v.Name] = int64(v.Percent * 100)
	}
}

func collectTopClients(mx map[string]int64, pmx *piholeMetrics) {
	if !pmx.hasTopClients() {
		return
	}
	for _, v := range *pmx.topClients {
		mx["top_client_"+v.Name] = v.Requests
	}
}

func collectTopDomains(mx map[string]int64, pmx *piholeMetrics) {
	if !pmx.hasTopQueries() {
		return
	}
	for _, v := range *pmx.topQueries {
		mx["top_domain_"+v.Name] = v.Hits
	}
}

func collectTopAdvertisers(mx map[string]int64, pmx *piholeMetrics) {
	if !pmx.hasTopAdvertisers() {
		return
	}
	for _, v := range *pmx.topAds {
		mx["top_ad_"+v.Name] = v.Hits
	}
}

func (p *Pihole) scrapePihole(doConcurrently bool) *piholeMetrics {
	pmx := new(piholeMetrics)

	taskSummary := func() {
		var err error
		pmx.summary, err = p.client.SummaryRaw()
		if err != nil {
			p.Error(err)
		}
	}
	taskQueryTypes := func() {
		var err error
		pmx.queryTypes, err = p.client.QueryTypes()
		if err != nil {
			p.Error(err)
		}
	}
	taskForwardDestinations := func() {
		var err error
		pmx.forwarders, err = p.client.ForwardDestinations()
		if err != nil {
			p.Error(err)
		}
	}
	taskTopClients := func() {
		var err error
		pmx.topClients, err = p.client.TopClients(defaultTopClients)
		if err != nil {
			p.Error(err)
		}
	}
	taskTopItems := func() {
		topItems, err := p.client.TopItems(defaultTopItems)
		if err != nil {
			p.Error(err)
		}
		if topItems != nil {
			pmx.topQueries = &topItems.TopQueries
			pmx.topAds = &topItems.TopAds
		}
	}

	var tasks = []func(){taskSummary}
	if p.Password != "" {
		tasks = []func(){
			taskSummary,
			taskQueryTypes,
			taskForwardDestinations,
			taskTopClients,
			taskTopItems,
		}
	}

	wg := &sync.WaitGroup{}

	wrap := func(call func()) func() {
		return func() {
			call()
			wg.Done()
		}
	}

	for _, task := range tasks {
		if doConcurrently {
			wg.Add(1)
			task = wrap(task)
			go task()
		} else {
			task()
		}
	}

	wg.Wait()

	return pmx
}
