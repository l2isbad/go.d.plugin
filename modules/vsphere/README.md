<!--
title: "vCenter Server monitoring with Netdata"
description: "Monitor the health and performance of vCenter servers with zero configuration, per-second metric granularity, and interactive visualizations."
custom_edit_url: https://github.com/netdata/go.d.plugin/edit/master/modules/vsphere/README.md
sidebar_label: "vCenter Servers"
-->

# vCenter Server monitoring with Netdata

[`VMware vCenter Server`](https://www.vmware.com/products/vcenter-server.html) is advanced server management software
that provides a centralized platform for controlling your VMware vSphere environments.

This module collects hosts and vms performance statistics from one or more `vCenter` servers depending on configuration.

## Charts

It produces the following charts:

#### Virtual Machine

- Cpu Usage Total in `%`
- Memory Usage Percentage in `%`
- Memory Usage in `KiB`
- VMKernel Memory Swap Rate in `KiB/s`
- VMKernel Memory Swap in `KiB`
- Network Bandwidth Total in `KiB/s`
- Network Packets Total in `packets`
- Network Drops Total in `packets`
- Disk Usage Total in `KiB/s`
- Disk Max Latency in `ms`
- Overall Alarm Status in `status`
- System Uptime in `seconds`

#### Host

- Cpu Usage Total in `%`
- Memory Usage Percentage in `%`
- Memory Usage in `KiB`
- VMKernel Memory Swap Rate in `KiB/s`
- VMKernel Memory Swap in `KiB`
- Network Bandwidth Total in `KiB/s`
- Network Packets Total in `packets`
- Network Drops Total in `packets`
- Network Errors Total in `errors`
- Disk Usage Total in `KiB/s`
- Disk Max Latency in `ms`
- Overall Alarm Status in `status`
- System Uptime in `seconds`

#### Datastore

- Storage Capacity in `KiB`
- Storage Provisioned in `KiB`
- Storage used in `KiB`
- Storage Provisioned average in `KiB`

## Configuration

Edit the `go.d/vsphere.conf` configuration file using `edit-config` from the
Netdata [config directory](https://learn.netdata.cloud/docs/configure/nodes), which is typically at `/etc/netdata`.

```bash
cd /etc/netdata # Replace this path with your Netdata config directory
sudo ./edit-config go.d/vsphere.conf
```

Needs only `url`, `username` and `password`. Here is an example for 2 servers:

```yaml
jobs:
  - name: vcenter1
    url: https://203.0.113.0
    username: admin@vsphere.local
    password: somepassword
    host_include: [ '/*' ]
    vm_include: [ '/*' ]
    datastore_include: ['/*']

  - name: vcenter2
    url: https://203.0.113.10
    username: admin@vsphere.local
    password: somepassword
    host_include: [ '/*' ]
    vm_include: [ '/*' ]
    datastore_include: ['/*']
```

For all available options please see
module [configuration file](https://github.com/netdata/go.d.plugin/blob/master/config/go.d/vsphere.conf).

## Hosts/vms/datastores filtering

Module supports filtering hosts and vms. Filtering options are `host_include` and `vm_include`.

- `host_include` is a list of match patterns: `/Dc pattern[/Cluster pattern/Host pattern]`.
- `vm_include` is a list of match patterns: `/Dc pattern[/Cluster pattern/Host pattern/VM name]`.
- `datastore_include` is a list of match patterns: `/Dc pattern[/Datastore pattern]`.

Pattern should start with `/`. It matches name,
syntax: [simple patterns](https://docs.netdata.cloud/libnetdata/simple_pattern/).

Examples:

```yaml
    host_include: # filter all hosts
      - '/!*'
    vm_include: # allow all vms
      - '/*'
    datastore_include:
      - '/*'
```

```yaml

host_include: # allow all DC1 datacenter hosts and DC2 datacenter hosts except HOST2
  - '/DC1/*'
  - '/DC2/*/!HOST2 *'
vm_include: # allow all vms from datacenters whose names starts with DC1 and from all hosts except HOST1 and HOST2
  - '/DC1*/*/!HOST1 !HOST2 */*'
```  

## Update every

Default `update_every` is 20 seconds and it doesnt make sense to decrease the value. **VMware real-time statistics are
generated at the 20-seconds specificity**.

It is likely that 20 seconds is not enough for big installations and the value should be tuned.

To get better view we recommend to run the collector in debug mode and see how much time it will take to collect
metrics.

Example (all not related debug lines were removed):

```
[ilyam@pc]$ ./go.d.plugin -d -m vsphere
[ DEBUG ] vsphere[vsphere] discover.go:94 discovering : starting resource discovering process
[ DEBUG ] vsphere[vsphere] discover.go:102 discovering : found 3 dcs, process took 49.329656ms
[ DEBUG ] vsphere[vsphere] discover.go:109 discovering : found 12 folders, process took 49.538688ms
[ DEBUG ] vsphere[vsphere] discover.go:116 discovering : found 3 clusters, process took 47.722692ms
[ DEBUG ] vsphere[vsphere] discover.go:123 discovering : found 2 hosts, process took 52.966995ms
[ DEBUG ] vsphere[vsphere] discover.go:130 discovering : found 2 vms, process took 49.832979ms
[ DEBUG ] vsphere[vsphere] discover.go:130 discovering : found 4 datastores, process took 65.457101ms
[ INFO  ] vsphere[vsphere] discover.go:140 discovering : found 3 dcs, 12 folders, 3 clusters (2 dummy), 2 hosts, 3 vms, 4/4 datastores, process took 249.655993ms
[ DEBUG ] vsphere[vsphere] build.go:12 discovering : building : starting building resources process
[ INFO  ] vsphere[vsphere] build.go:23 discovering : building : built 3/3 dcs, 12/12 folders, 3/3 clusters, 2/2 hosts, 3/3 vms, 4/4 datastores, process took 63.3µs
[ DEBUG ] vsphere[vsphere] hierarchy.go:10 discovering : hierarchy : start setting resources hierarchy process
[ INFO  ] vsphere[vsphere] hierarchy.go:18 discovering : hierarchy : set 3/3 clusters, 2/2 hosts, 3/3 vms, 4/4 datastores, process took 6.522µs
[ DEBUG ] vsphere[vsphere] filter.go:24 discovering : filtering : starting filtering resources process
[ DEBUG ] vsphere[vsphere] filter.go:45 discovering : filtering : removed 0 unmatched hosts
[ DEBUG ] vsphere[vsphere] filter.go:56 discovering : filtering : removed 0 unmatched vms
[ DEBUG ] vsphere[vsphere] filter.go:56 discovering : filtering : removed 0 unmatched datastores
[ INFO  ] vsphere[vsphere] filter.go:29 discovering : filtering : filtered 0/2 hosts, 0/3 vms, 0/4 datastores, process took 42.973µs
[ DEBUG ] vsphere[vsphere] metric_lists.go:14 discovering : metric lists : starting resources metric lists collection process
[ INFO  ] vsphere[vsphere] metric_lists.go:30 discovering : metric lists : collected metric lists for 2/2 hosts, 3/3 vms, 4/4 datastores, process took 275.60764ms
[ INFO  ] vsphere[vsphere] discover.go:74 discovering : discovered 2/2 hosts, 3/3 vms, 4/4 datastores, the whole process took 525.614041ms
[ INFO  ] vsphere[vsphere] discover.go:11 starting discovery process, will do discovery every 5m0s
[ DEBUG ] vsphere[vsphere] collect.go:11 starting collection process
[ DEBUG ] vsphere[vsphere] scrape.go:48 scraping : scraped metrics for 2/2 hosts, process took 96.257374ms
[ DEBUG ] vsphere[vsphere] scrape.go:60 scraping : scraped metrics for 3/3 vms, process took 57.879697ms
[ DEBUG ] vsphere[vsphere] scrape.go:60 scraping : scraped metrics for 4/4 datastores, process took 57.879697ms
[ DEBUG ] vsphere[vsphere] collect.go:23 metrics collected, process took 154.77997ms

```

There you can see that discovering took `525.614041ms`, collecting metrics took `154.77997ms`. Discovering is a separate
thread, it doesnt affect collecting.

`update_every` and `timeout` parameters should be adjusted based on these numbers.

## Troubleshooting

To troubleshoot issues with the `vsphere` collector, run the `go.d.plugin` with the debug option enabled. The output
should give you clues as to why the collector isn't working.

First, navigate to your plugins directory, usually at `/usr/libexec/netdata/plugins.d/`. If that's not the case on your
system, open `netdata.conf` and look for the setting `plugins directory`. Once you're in the plugin's directory, switch
to the `netdata` user.

```bash
cd /usr/libexec/netdata/plugins.d/
sudo -u netdata -s
```

You can now run the `go.d.plugin` to debug the collector:

```bash
./go.d.plugin -d -m vsphere
```
