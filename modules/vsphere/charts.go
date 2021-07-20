package vsphere

import (
	"fmt"
	"strings"

	rs "github.com/netdata/go.d.plugin/modules/vsphere/resources"

	"github.com/netdata/go.d.plugin/agent/module"
)

type (
	// Charts is an alias for module.Charts
	Charts = module.Charts
	// Chart is an alias for module.Chart
	Chart = module.Chart
	// Dims is an alias for module.Dims
	Dims = module.Dims
)

const (
	hostPrio = module.Priority
	vmPrio   = hostPrio + 200
	datastorePrio = vmPrio + 200
)

var (
	datastoreCharts = func() Charts{
		cs := Charts{}
		panicIf(cs.Add(datastoreStorageCharts...))
		return cs
	}()
	// Ref : https://www.vmware.com/support/developer/converter-sdk/conv51_apireference/disk_storutil_counters.html
	datastoreStorageCharts = Charts{
		{
			ID:      "%s_disk_metrics",
			Title:   "Disk Metrics",
			Units:   "KiB",
			Fam:     "datastore %s",
			Ctx:     "vsphere.disk_metrics",
			Dims: Dims{
				{ID: "%s_disk.used.latest", Name:"used"},
				{ID: "%s_disk.provisioned.latest", Name:"provisioned"},
				{ID: "%s_disk.capacity.latest", Name:"capacity"},
				{ID: "%s_disk.capacity.provisioned.average", Name:"provisioned average"},
			},
		},
	}

	vmCharts = func() Charts {
		cs := Charts{}
		panicIf(cs.Add(vmCPUCharts...))
		panicIf(cs.Add(vmMemCharts...))
		panicIf(cs.Add(vmNetCharts...))
		panicIf(cs.Add(vmDiskCharts...))
		panicIf(cs.Add(vmSystemCharts...))
		return cs
	}()

	vmCPUCharts = Charts{
		{
			ID:    "%s_cpu_usage_total",
			Title: "Cpu Usage Total",
			Units: "percentage",
			Fam:   "vm %s (%s)",
			Ctx:   "vsphere.cpu_usage_total",
			Type:  module.Area,
			Dims: Dims{
				{ID: "%s_cpu.usage.average", Name: "used", Div: 100},
			},
		},
	}
	// Ref: https://www.vmware.com/support/developer/converter-sdk/conv51_apireference/memory_counters.html
	vmMemCharts = Charts{
		{
			ID:    "%s_mem_usage_percentage",
			Title: "Memory Usage Percentage",
			Units: "percentage",
			Fam:   "vm %s (%s)",
			Ctx:   "vsphere.vm_mem_usage_percentage",
			Type:  module.Area,
			Dims: Dims{
				{ID: "%s_mem.usage.average", Name: "used", Div: 100},
			},
		},
		{
			ID:    "%s_mem_usage",
			Title: "Memory Usage",
			Units: "KiB",
			Fam:   "vm %s (%s)",
			Ctx:   "vsphere.vm_mem_usage",
			Dims: Dims{
				{ID: "%s_mem.granted.average", Name: "granted"},
				{ID: "%s_mem.consumed.average", Name: "consumed"},
				{ID: "%s_mem.active.average", Name: "active"},
				{ID: "%s_mem.shared.average", Name: "shared"},
			},
		},
		{
			ID:    "%s_mem_swap_rate",
			Title: "VMKernel Memory Swap Rate",
			Units: "KiB/s",
			Fam:   "vm %s (%s)",
			Ctx:   "vsphere.vm_mem_swap_rate",
			Type:  module.Area,
			Dims: Dims{
				{ID: "%s_mem.swapinRate.average", Name: "in"},
				{ID: "%s_mem.swapoutRate.average", Name: "out"},
			},
		},
		{
			ID:    "%s_mem_swap",
			Title: "VMKernel Memory Swap",
			Units: "KiB",
			Fam:   "vm %s (%s)",
			Ctx:   "vsphere.vm_mem_swap",
			Type:  module.Area,
			Dims: Dims{
				{ID: "%s_mem.swapped.average", Name: "swapped"},
			},
		},
	}
	vmNetCharts = Charts{
		{
			ID:    "%s_net_bandwidth_total",
			Title: "Network Bandwidth Total",
			Units: "KiB/s",
			Fam:   "vm %s (%s)",
			Ctx:   "vsphere.net_bandwidth_total",
			Type:  module.Area,
			Dims: Dims{
				{ID: "%s_net.bytesRx.average", Name: "rx"},
				{ID: "%s_net.bytesTx.average", Name: "tx", Mul: -1},
			},
		},
		{
			ID:    "%s_net_packets_total",
			Title: "Network Packets Total",
			Units: "packets",
			Fam:   "vm %s (%s)",
			Ctx:   "vsphere.net_packets_total",
			Dims: Dims{
				{ID: "%s_net.packetsRx.summation", Name: "rx"},
				{ID: "%s_net.packetsTx.summation", Name: "tx", Mul: -1},
			},
		},
		{
			ID:    "%s_net_drops_total",
			Title: "Network Drops Total",
			Units: "packets",
			Fam:   "vm %s (%s)",
			Ctx:   "vsphere.net_drops_total",
			Dims: Dims{
				{ID: "%s_net.droppedRx.summation", Name: "rx"},
				{ID: "%s_net.droppedTx.summation", Name: "tx", Mul: -1},
			},
		},
	}
	vmDiskCharts = Charts{
		{
			ID:    "%s_disk_usage_total",
			Title: "Disk Usage Total",
			Units: "KiB/s",
			Fam:   "vm %s (%s)",
			Ctx:   "vsphere.disk_usage_total",
			Type:  module.Area,
			Dims: Dims{
				{ID: "%s_disk.read.average", Name: "read"},
				{ID: "%s_disk.write.average", Name: "write", Mul: -1},
			},
		},
		{
			ID:    "%s_disk_max_latency",
			Title: "Disk Max Latency",
			Units: "ms",
			Fam:   "vm %s (%s)",
			Ctx:   "vsphere.disk_max_latency",
			Dims: Dims{
				{ID: "%s_disk.maxTotalLatency.latest", Name: "latency"},
			},
		},
	}
	vmSystemCharts = Charts{
		{
			ID:    "%s_overall_status",
			Title: "Overall Alarm Status",
			Units: "status",
			Fam:   "vm %s (%s)",
			Ctx:   "vsphere.overall_status",
			Dims: Dims{
				{ID: "%s_overall.status", Name: "status"},
			},
		},
		{
			ID:    "%s_system_uptime",
			Title: "System Uptime",
			Units: "seconds",
			Fam:   "vm %s (%s)",
			Ctx:   "vsphere.system_uptime",
			Dims: Dims{
				{ID: "%s_sys.uptime.latest", Name: "time"},
			},
		},
	}
)

var (
	hostCharts = func() Charts {
		cs := Charts{}
		panicIf(cs.Add(hostCPUCharts...))
		panicIf(cs.Add(hostMemCharts...))
		panicIf(cs.Add(hostNetCharts...))
		panicIf(cs.Add(hostDiskCharts...))
		panicIf(cs.Add(hostSystemCharts...))
		return cs
	}()
	hostCPUCharts = Charts{
		{
			ID:    "%s_cpu_usage_total",
			Title: "Cpu Usage Total",
			Units: "percentage",
			Fam:   "host %s",
			Ctx:   "vsphere.cpu_usage_total",
			Type:  module.Area,
			Dims: Dims{
				{ID: "%s_cpu.usage.average", Name: "used", Div: 100},
			},
		},
	}
	// Ref: https://www.vmware.com/support/developer/converter-sdk/conv51_apireference/memory_counters.html
	hostMemCharts = Charts{
		{
			ID:    "%s_mem_usage_percentage",
			Title: "Memory Usage Percentage",
			Units: "percentage",
			Fam:   "host %s",
			Ctx:   "vsphere.host_mem_usage_percentage",
			Type:  module.Area,
			Dims: Dims{
				{ID: "%s_mem.usage.average", Name: "used", Div: 100},
			},
		},
		{
			ID:    "%s_mem_usage",
			Title: "Memory Usage",
			Units: "KiB",
			Fam:   "host %s",
			Ctx:   "vsphere.host_mem_usage",
			Dims: Dims{
				{ID: "%s_mem.granted.average", Name: "granted"},
				{ID: "%s_mem.consumed.average", Name: "consumed"},
				{ID: "%s_mem.active.average", Name: "active"},
				{ID: "%s_mem.shared.average", Name: "shared"},
				{ID: "%s_mem.sharedcommon.average", Name: "sharedcommon"},
			},
		},
		{
			ID:    "%s_mem_swap_rate",
			Title: "VMKernel Memory Swap Rate",
			Units: "KiB/s",
			Fam:   "host %s",
			Ctx:   "vsphere.host_mem_swap_rate",
			Type:  module.Area,
			Dims: Dims{
				{ID: "%s_mem.swapinRate.average", Name: "in"},
				{ID: "%s_mem.swapoutRate.average", Name: "out"},
			},
		},
	}
	hostNetCharts = Charts{
		{
			ID:    "%s_net_bandwidth_total",
			Title: "Network Bandwidth Total",
			Units: "KiB/s",
			Fam:   "host %s",
			Ctx:   "vsphere.net_bandwidth_total",
			Type:  module.Area,
			Dims: Dims{
				{ID: "%s_net.bytesRx.average", Name: "rx"},
				{ID: "%s_net.bytesTx.average", Name: "tx", Mul: -1},
			},
		},
		{
			ID:    "%s_net_packets_total",
			Title: "Network Packets Total",
			Units: "packets",
			Fam:   "host %s",
			Ctx:   "vsphere.net_packets_total",
			Dims: Dims{
				{ID: "%s_net.packetsRx.summation", Name: "rx"},
				{ID: "%s_net.packetsTx.summation", Name: "tx", Mul: -1},
			},
		},
		{
			ID:    "%s_net_drops_total",
			Title: "Network Drops Total",
			Units: "packets",
			Fam:   "host %s",
			Ctx:   "vsphere.net_drops_total",
			Dims: Dims{
				{ID: "%s_net.droppedRx.summation", Name: "rx"},
				{ID: "%s_net.droppedTx.summation", Name: "tx", Mul: -1},
			},
		},
		{
			ID:    "%s_net_errors_total",
			Title: "Network Errors Total",
			Units: "errors",
			Fam:   "host %s",
			Ctx:   "vsphere.net_errors_total",
			Dims: Dims{
				{ID: "%s_net.errorsRx.summation", Name: "rx"},
				{ID: "%s_net.errorsTx.summation", Name: "tx", Mul: -1},
			},
		},
	}
	hostDiskCharts = Charts{
		{
			ID:    "%s_disk_usage_total",
			Title: "Disk Usage Total",
			Units: "KiB/s",
			Fam:   "host %s",
			Ctx:   "vsphere.disk_usage_total",
			Type:  module.Area,
			Dims: Dims{
				{ID: "%s_disk.read.average", Name: "read"},
				{ID: "%s_disk.write.average", Name: "write", Mul: -1},
			},
		},
		{
			ID:    "%s_disk_max_latency",
			Title: "Disk Max Latency",
			Units: "ms",
			Fam:   "host %s",
			Ctx:   "vsphere.disk_max_latency",
			Dims: Dims{
				{ID: "%s_disk.maxTotalLatency.latest", Name: "latency"},
			},
		},
	}
	hostSystemCharts = Charts{
		{
			ID:    "%s_overall_status",
			Title: "Overall Alarm Status",
			Units: "status",
			Fam:   "host %s",
			Ctx:   "vsphere.overall_status",
			Dims: Dims{
				{ID: "%s_overall.status", Name: "status"},
			},
		},
		{
			ID:    "%s_system_uptime",
			Title: "System Uptime",
			Units: "seconds",
			Fam:   "host %s",
			Ctx:   "vsphere.system_uptime",
			Dims: Dims{
				{ID: "%s_sys.uptime.latest", Name: "time"},
			},
		},
	}
)

func (vs *VSphere) updateHostsCharts(collected map[string]string) {
	for id, userID := range collected {
		if vs.charted[userID] {
			continue
		}
		h := vs.resources.Hosts.Get(id)
		if h == nil {
			continue
		}
		vs.charted[userID] = true

		cs := newHostCharts(h, userID)
		if err := vs.charts.Add(*cs...); err != nil {
			vs.Error(err)
		}
	}
}

func newHostCharts(host *rs.Host, userID string) *Charts {
	cs := hostCharts.Copy()
	for i, c := range *cs {
		setHostChart(c, host, userID, hostPrio+i)
	}
	return cs
}

func setHostChart(chart *Chart, host *rs.Host, userID string, prio int) {
	chart.Priority = prio
	chart.ID = fmt.Sprintf(chart.ID, userID)
	chart.Fam = fmt.Sprintf(chart.Fam, host.Name)
	for _, d := range chart.Dims {
		d.ID = fmt.Sprintf(d.ID, host.ID)
	}
}

func (vs *VSphere) updateVMsCharts(collected map[string]string) {
	for id, userID := range collected {
		if vs.charted[userID] {
			continue
		}
		vm := vs.resources.VMs.Get(id)
		if vm == nil {
			continue
		}
		vs.charted[userID] = true

		cs := newVMCHarts(vm, userID)
		if err := vs.charts.Add(*cs...); err != nil {
			vs.Error(err)
		}
	}
}

func newVMCHarts(vm *rs.VM, userID string) *Charts {
	cs := vmCharts.Copy()
	for i, c := range *cs {
		setVMChart(c, vm, userID, vmPrio+i)
	}
	return cs
}

func setVMChart(chart *Chart, vm *rs.VM, userID string, prio int) {
	chart.Priority = prio
	chart.ID = fmt.Sprintf(chart.ID, userID)
	chart.Fam = fmt.Sprintf(chart.Fam, vm.Name, vm.Hier.Host.Name)
	for _, d := range chart.Dims {
		d.ID = fmt.Sprintf(d.ID, vm.ID)
	}
}

func (vs *VSphere) updateDatastoresCharts(collected map[string]string) {
	for id, userID := range collected {
		if vs.charted[userID] {
			continue
		}
		d := vs.resources.Datastores.Get(id)
		if d == nil {
			continue
		}
		vs.charted[userID] = true

		cs := newDatastoreCharts(d, userID)
		if err := vs.charts.Add(*cs...); err != nil {
			vs.Error(err)
		}
	}
}

func newDatastoreCharts(datastore *rs.Datastore, userID string) *Charts {
	cs := datastoreCharts.Copy()
	for i, c := range *cs {
		setDatastoreChart(c, datastore, userID, datastorePrio+i)
	}
	return cs
}

func setDatastoreChart(chart *Chart, datastore *rs.Datastore, userID string, prio int) {
	chart.Priority = prio
	chart.ID = fmt.Sprintf(chart.ID, userID)
	chart.Fam = fmt.Sprintf(chart.Fam, datastore.Name)
	for _, d := range chart.Dims {
		d.ID = fmt.Sprintf(d.ID, datastore.ID)
	}
}

func (vs *VSphere) removeFromCharts(prefix string) {
	for _, c := range *vs.charts {
		if strings.HasPrefix(c.ID, prefix) {
			c.MarkRemove()
			c.MarkNotCreated()
		}
	}
}

//func findMetricSeriesByPrefix(ms []performance.MetricSeries, prefix string) []performance.MetricSeries {
//	from := sort.Search(len(ms), func(i int) bool { return ms[i].Name >= prefix })
//
//	if from == len(ms) || !strings.HasPrefix(ms[from].Name, prefix) {
//		return nil
//	}
//
//	until := from + 1
//	for until < len(ms) && strings.HasPrefix(ms[until].Name, prefix) {
//		until++
//	}
//	return ms[from:until]
//}

func panicIf(err error) {
	if err != nil {
		panic(err)
	}
}
