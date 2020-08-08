wanted to add a dynamic ps1 prompt for remote k8s clusters but calling kubectl everytime a bash prompt is created is far too costly.

i.e 
```spokesd --kubeconfig kubeconfigfile [--config file]```

my intent with spokesd is to have it serve as a out-of-cluster (systemd-managed hence the d???) aggregator of k8s clusters events.
either have it write aggregated/filtered events to an unix socket or some other ipc-optimized data sink

does golang have native support for named-semaphores?

add a spokesd client that scrapes the configured spokesd.sock or what have you and serve info to $ps1 or wherever
i.e

```spokesctl get events [--type type] [--cluster cluster]```

can also integrate with Msft Graph API for tenant-managed services like Teams or Outlook?



