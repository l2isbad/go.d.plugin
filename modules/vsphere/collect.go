package vsphere

import (
	"errors"
	"fmt"
	"strings"
	"time"

	rs "github.com/netdata/go.d.plugin/modules/vsphere/resources"

	"github.com/vmware/govmomi/performance"
)

func (vs *VSphere) collect() (map[string]int64, error) {
	vs.collectionLock.Lock()
	defer vs.collectionLock.Unlock()
	defer vs.removeStale()

	vs.Debug("starting collection process")
	t := time.Now()
	mx := make(map[string]int64)
	
	err := vs.collectHosts(mx)
	if err != nil {
		return mx, err
	}

	err = vs.collectVMs(mx)
	if err != nil {
		return mx, err
	}

	err = vs.collectDatastores(mx)
	if err != nil {
		return mx, err
	}

	vs.Debugf("metrics collected, process took %s", time.Since(t))
	return mx, nil
}

func (vs *VSphere) collectHosts(mx map[string]int64) error {
	if len(vs.resources.Hosts) == 0 {
		return nil
	}
	// NOTE: returns unsorted if at least one types.PerfMetricId Instance is not ""
	metrics := vs.ScrapeHosts(vs.resources.Hosts)
	if len(metrics) == 0 {
		return errors.New("failed to scrape hosts metrics")
	}

	hosts := vs.collectHostsMetrics(mx, metrics)
	vs.updateDiscoveredHosts(hosts)
	vs.updateHostsCharts(hosts)
	return nil
}

func (vs *VSphere) collectHostsMetrics(mx map[string]int64, metrics []performance.EntityMetric) map[string]string {
	hosts := make(map[string]string)
	for _, m := range metrics {
		host := vs.resources.Hosts.Get(m.Entity.Value)
		if host == nil {
			continue
		}
		writeHostMetrics(mx, host, m.Value)
		hosts[host.ID] = vs.hostID(host)
	}
	return hosts
}

func writeHostMetrics(mx map[string]int64, host *rs.Host, metrics []performance.MetricSeries) {
	for _, m := range metrics {
		if len(m.Value) == 0 || m.Value[0] == -1 {
			continue
		}
		key := fmt.Sprintf("%s_%s", host.ID, m.Name)
		mx[key] = m.Value[0]
	}
	key := fmt.Sprintf("%s_overall.status", host.ID)
	mx[key] = overallStatusToInt(host.OverallStatus)
}

func (vs *VSphere) updateDiscoveredHosts(collected map[string]string) {
	for _, h := range vs.resources.Hosts {
		id := vs.hostID(h)
		if v, ok := collected[h.ID]; !ok || id != v {
			vs.discoveredHosts[id] += 1
		} else {
			vs.discoveredHosts[id] = 0
		}
	}
}

func (vs VSphere) hostID(host *rs.Host) (id string) {
	id = host.ID
	if vs.HostMetrics.Name {
		id = join(id, "name", host.Name)
	}
	if vs.HostMetrics.Cluster {
		id = join(id, "cluster", host.Hier.Cluster.Name)
	}
	if vs.HostMetrics.DataCenter {
		id = join(id, "datacenter", host.Hier.DC.Name)
	}
	return cleanID(id)
}

func (vs *VSphere) collectVMs(mx map[string]int64) error {
	if len(vs.resources.VMs) == 0 {
		return nil
	}
	// NOTE: returns unsorted if at least one types.PerfMetricId Instance is not ""
	ems := vs.ScrapeVMs(vs.resources.VMs)
	if len(ems) == 0 {
		return errors.New("failed to scrape vms metrics")
	}

	vms := vs.collectVMsMetrics(mx, ems)
	vs.updateDiscoveredVMs(vms)
	vs.updateVMsCharts(vms)
	return nil
}

func (vs *VSphere) collectVMsMetrics(mx map[string]int64, ems []performance.EntityMetric) map[string]string {
	vms := make(map[string]string)
	for _, em := range ems {
		vm := vs.resources.VMs.Get(em.Entity.Value)
		if vm == nil {
			continue
		}
		writeVMMetrics(mx, vm, em.Value)
		vms[vm.ID] = vs.vmID(vm)
	}
	return vms
}

func writeVMMetrics(mx map[string]int64, vm *rs.VM, metrics []performance.MetricSeries) {
	for _, m := range metrics {
		if len(m.Value) == 0 || m.Value[0] == -1 {
			continue
		}
		key := fmt.Sprintf("%s_%s", vm.ID, m.Name)
		mx[key] = m.Value[0]
	}
	key := fmt.Sprintf("%s_overall.status", vm.ID)
	mx[key] = overallStatusToInt(vm.OverallStatus)
}

func (vs *VSphere) updateDiscoveredVMs(collected map[string]string) {
	for _, vm := range vs.resources.VMs {
		id := vs.vmID(vm)
		if v, ok := collected[vm.ID]; !ok || id != v {
			vs.discoveredVMs[id] += 1
		} else {
			vs.discoveredVMs[id] = 0
		}
	}
}

func (vs VSphere) vmID(vm *rs.VM) (id string) {
	id = vm.ID
	if vs.VMMetrics.Name {
		id = join(id, "name", vm.Name)
	}
	if vs.VMMetrics.Host {
		id = join(id, "host", vm.Hier.Host.Name)
	}
	if vs.VMMetrics.Cluster {
		id = join(id, "cluster", vm.Hier.Cluster.Name)
	}
	if vs.VMMetrics.DataCenter {
		id = join(id, "datacenter", vm.Hier.DC.Name)
	}
	return cleanID(id)
}

func (vs *VSphere) collectDatastores(mx map[string]int64) error {
	if len(vs.resources.Datastores) == 0 {
		return nil
	}
	// NOTE: returns unsorted if at least one types.PerfMetricId Instance is not ""
	metrics := vs.ScrapeDatastores(vs.resources.Datastores)
	if len(metrics) == 0 {
		return errors.New("failed to scrape datastores metrics")
	}
	datastores := vs.collectDatastoresMetrics(mx, metrics)
	vs.updateDiscoveredDatastores(datastores)
	vs.updateDatastoresCharts(datastores)
	return nil
}

func (vs *VSphere) collectDatastoresMetrics(mx map[string]int64, metrics []performance.EntityMetric) map[string]string {
	datastores := make(map[string]string)
	for _, m := range metrics {
		datastore := vs.resources.Datastores.Get(m.Entity.Value)
		if datastore == nil {
			continue
		}
		writeDatastoreMetrics(mx, datastore, m.Value)
		datastores[datastore.ID] = vs.datastoreID(datastore)
	}
	return datastores
}

func writeDatastoreMetrics(mx map[string]int64, datastore *rs.Datastore, metrics []performance.MetricSeries) {
	for _, m := range metrics {
		if len(m.Value) == 0 || m.Value[0] == -1 {
			continue
		}
		key := fmt.Sprintf("%s_%s", datastore.ID, m.Name)
		mx[key] = m.Value[0]
	}
	key := fmt.Sprintf("%s_accessible", datastore.ID)
	mx[key] = accessibleToInt(datastore.Accessible)
}

func (vs *VSphere) updateDiscoveredDatastores(collected map[string]string) {
	for _, h := range vs.resources.Datastores {
		id := vs.datastoreID(h)
		if v, ok := collected[h.ID]; !ok || id != v {
			vs.discoveredDatastores[id] += 1
		} else {
			vs.discoveredDatastores[id] = 0
		}
	}
}

func (vs VSphere) datastoreID(datastore *rs.Datastore) (id string) {
	id = datastore.ID
	if vs.DatastoreMetrics.Name {
		id = join(id, "name", datastore.Name)
	}
	return cleanID(id)
}

const failedUpdatesLimit = 10

func (vs *VSphere) removeStale() {
	for userHostID, fails := range vs.discoveredHosts {
		if fails < failedUpdatesLimit {
			continue
		}
		vs.removeFromCharts(userHostID)
		delete(vs.charted, userHostID)
		delete(vs.discoveredHosts, userHostID)
	}
	for userVMID, fails := range vs.discoveredVMs {
		if fails < failedUpdatesLimit {
			continue
		}
		vs.removeFromCharts(userVMID)
		delete(vs.charted, userVMID)
		delete(vs.discoveredVMs, userVMID)
	}
	for userDatastoreID, fails := range vs.discoveredDatastores {
		if fails < failedUpdatesLimit {
			continue
		}
		vs.removeFromCharts(userDatastoreID)
		delete(vs.charted, userDatastoreID)
		delete(vs.discoveredDatastores, userDatastoreID)
	}
}

func join(a, prefix, value string) string {
	if value == "" {
		value = "unknown"
	}
	return a + "_" + prefix + "-" + value
}

func cleanID(id string) string {
	return r.Replace(id)
}

var r = strings.NewReplacer(" ", "_", ".", "_")

func overallStatusToInt(status string) int64 {
	// ManagedEntityStatus
	switch status {
	default:
		return 0
	case "grey":
		return 0
	case "green":
		return 1
	case "yellow":
		return 2
	case "red":
		return 3
	}
}

func accessibleToInt(accessible bool) int64 {
	if accessible == true {
		return 1
	}else{
		return 3
	}
}
