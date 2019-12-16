# kubelet

The [`Kubelet`](https://kubernetes.io/docs/reference/command-line-tools-reference/kubelet/)  is the primary “node agent” that runs on each node in the kubernetes environment.

This module will monitor one or more `Kubelet` instances depending on configuration.

## Charts

It produces the following charts:

-   API Server Audit Requests in `requests/s`
-   API Server Failed Data Encryption Key(DEK) Generation Operations in `events/s`
-   API Server Latencies Of Data Encryption Key(DEK) Generation Operations in `observes/s`
-   API Server Latencies Of Data Encryption Key(DEK) Generation Operations Percentage in `%`
-   API Server Storage Envelope Transformation Cache Misses` in `events/s`
-   Number Of Containers Currently Running in `containers`
-   Number Of Pods Currently Running in `pods`
-   Bytes Used By The Pod Logs On The Filesystem in `bytes`
-   Runtime Operations By Type in `operations/s`
-   Docker Operations By Type in `operations/s`
-   Docker Operations Errors By Type in `operations/s`
-   Node Configuration-Related Error in `bool`
-   PLEG Relisting Interval Summary in `microseconds`
-   PLEG Relisting Latency Summary in `microseconds`
-   Token Requests To The Alternate Token Source in `requests/s`
-   REST Client HTTP Requests By Status Code in `requests/s`
-   REST Client HTTP Requests By Method in `requests/s`

Per every plugin:

-   Volume Manager State Of The World in `state`
 
## Configuration

Needs only `url` to kubelet metric-address. Here is an example for 2 instances:

```yaml
jobs:
  - name: local
    url : http://127.0.0.1:10255/metrics
      
  - name: remote
    url : http://203.0.113.10:10255/metrics
```

For all available options please see module [configuration file](https://github.com/netdata/go.d.plugin/blob/master/config/go.d/k8s_kubelet.conf).

## Troubleshooting

Check the module debug output. Run the following command as `netdata` user:

> ./go.d.plugin -d -m k8s_kubelet
