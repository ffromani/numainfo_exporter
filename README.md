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

## license
(C) 2020 Red Hat Inc and licensed under the Apache License v2

`internal/pkg/certutil` is (C) 2014 the Kubernetes Authors

## build
just run
```bash
make
```

## Container image
TBD


