package couchbase

import (
	"github.com/netdata/go.d.plugin/agent/module"
)

type (
	Charts = module.Charts
	Chart  = module.Chart
	Dim    = module.Dim
)

var dbPercentCharts = Chart{
	ID:    "quota_percent_used_stats",
	Title: "Quota Percent Used Per Bucket",
	Units: "%",
	Fam:   "quota percent used",
	Ctx:   "couchbase.quota_percent_used_stats",
}

var opsPerSecCharts = Chart{
	ID:    "ops_per_sec_stats",
	Title: "Operations Per Second",
	Units: "num",
	Fam:   "ops per sec",
	Ctx:   "couchbase.ops_per_sec_stats",
}

var diskFetchesCharts = Chart{
	ID:    "disk_fetches_stats",
	Title: "Disk Fetches",
	Units: "num/s",
	Fam:   "disk fetches",
	Ctx:   "couchbase.disk_fetches_stats",
}

var itemCountCharts = Chart{
	ID:    "item_count_stats",
	Title: "Item Count",
	Units: "items",
	Fam:   "item count",
	Ctx:   "couchbase.item_count_stats",
	Type:  module.Stacked,
}

var diskUsedCharts = Chart{
	ID:    "disk_used_stats",
	Title: "Disk Used Per Bucket",
	Units: "KiB/s",
	Fam:   "disk used",
	Ctx:   "couchbase.disk_used_stats",
}

var dataUsedCharts = Chart{
	ID:    "data_used_stats",
	Title: "Data Used Per Bucket",
	Units: "KiB/s",
	Fam:   "data used",
	Ctx:   "couchbase.data_used_stats",
}

var memUsedCharts = Chart{
	ID:    "mem_used_stats",
	Title: "Memory Used Per Bucket",
	Units: "KiB/s",
	Fam:   "memory used",
	Ctx:   "couchbase.mem_used_stats",
}

var vbActiveNumNonResidentCharts = Chart{
	ID:    "vb_active_num_non_resident_stats",
	Title: "Number Of Non-Resident Items",
	Units: "num/s",
	Fam:   "vb active num non resident",
	Ctx:   "couchbase.vb_active_num_non_resident_stats",
}
