# numainfo_exporter - a NUMA-aware resource exporter for kubernetes

`numainfo_exporter` exports on `prometheus` the kubernetes node resources, per-NUMA node.

## Description

`numainfo_exporter` is a deaemon which exports on prometheus the current per-NUMA node capacity and allocation
of the resources exposed by a kubernetes node.
To do so, `numainfo_exporter` reads the kubelet checkpoints file and compute the per-NUMA node allocation on its own.
This works because the kubelet records each state change on a checkpoint file. However, this is not meant to be a
stable long-term solutions; a better alternative would be to have the kubelet export the data in an official and
supported way, and we will contribute to the design and implementation effort in this direction.

## Rationale

* this project reads the kubelet checkpoints files because this is the best (and only) way we can reliable peek in
  the kubelet state counters _without patching the kubelet code to export data_. This allows the tool to run in pristine
  clusters. This also means:
  - for safety, `numainfo_exporter` *must* access the state files in read-only mode, and take the utmost care to minimize
    (better: avoid) any interference with the kubelet.
  - because of the previous requirements, `numainfo_exporter` may end up reading partially-written, corrupt checkpoint files.
    Races are unavoidable because how the kubelet works. `numainfo_exporter` shoudld just warn and give up in this case,
    reporting no data and trying again next time.
* to assess the per-NUMA node capacity, `numainfo_exporter` reads the data under `/sys` the same way the kubelet does.
  This is safe since the data is read-only and the interface stable. Furtermore, the node _capacity_ changes rarely if ever
  during the node lifecycle
* `numainfo_exporter` exports the gathered data on a prometheus interface. There is no real constraint to expose a prometheus
  interface -the tool can for example also feed a CRD-, but doing so allows to integrate nicely with the ongoing (202004) efforts
  to enhance the kubernetes scheduler, like the [telemetry-aware scheduling](https://github.com/intel/telemetry-aware-scheduling).

## Open issues

1. In rapidly-changing (pods created/destroyed frequently) the data reported by `numainfo_exporter` can get stale fast because
   the reporting is completely asynchronous with respect to the kubelet.
   This cannot be fixed, and can only be improved by making the `numainfo_exporter` push updates, which doesn't fit the prometheus model.
2. There is no way to avoid races with kubelet. `numainfo_exporter`. This means some pull cycles may unpredictably return no fresh data.
   This cannot be fixed. The only way to avoid that is extend the kubelet to export the same data `numainfo_exporter` provides (e.g. new API).
3. There is a dependency on the (private) checkpoint files internal format. Each version of `numainfo_exporter` is tied to a specific
   kubernetes version. This can be improved somehow, but never really solved.

Please note that only the bullet point 3 above is `numainfo_exporter` specific. The other bullet points affect any *external* tool which
peeks into the kubelet state.

## Requirements:

`numainfo_exporter` works with and requires kubernetes >= 1.18.0.

## Available metrics

To get the currently available metrics, you can run
```bash
$ ./numainfo_exporter -K /var/tmp -N test.k8s.io -M 2>&1 | grep numainfo
```

Example output:
```bash
# HELP numainfo_core_count CPU cores per NUMA node, count.
# TYPE numainfo_core_count gauge
numainfo_core_count{node="test.k8s.io",numanode="node00",type="capacity"} 4
# HELP numainfo_node_count NUMA nodes per node, count.
# TYPE numainfo_node_count gauge
numainfo_node_count{node="test.k8s.io"} 1
# HELP numainfo_version Version information
# TYPE numainfo_version gauge
numainfo_version{branch="master",goversion="go1.13.9",kubeversion="devel",revision="15ced69",version="1"} 1

```

In this case the "allocation" label is missing, because this output was taken from a developer laptop with no kubelet running
(this is also the reason for the option `-K` in the aforementioned example).

Another example:
```bash
# HELP numainfo_core_count CPU cores per NUMA node, count.
# TYPE numainfo_core_count gauge
numainfo_core_count{node="kind-worker",numanode="node00",type="allocation"} 1
numainfo_core_count{node="kind-worker",numanode="node00",type="capacity"} 12
# HELP numainfo_node_count NUMA nodes per node, count.
# TYPE numainfo_node_count gauge
numainfo_node_count{node="kind-worker"} 1
# HELP numainfo_version Version information
# TYPE numainfo_version gauge
numainfo_version{branch="kube118",goversion="go1.13.9",kubeversion="devel",revision="434f808",version="1"} 1
```

This is from a live multi-master, multi-worker [kind](https://kubernetes.io/docs/setup/learning-environment/kind/) cluster.


## license
(C) 2020 Red Hat Inc and licensed under the Apache License v2

`internal/pkg/certutil` is (C) 2014 the Kubernetes Authors

## build
just run
```bash
make
```

## Container image

```bash
podman pull quay.io/fromani/numainfo_exporter:devel
```
