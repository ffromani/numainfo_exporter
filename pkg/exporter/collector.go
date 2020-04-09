package exporter

import (
	"fmt"
	"runtime"

	"k8s.io/klog"

	"github.com/prometheus/client_golang/prometheus"

	verinfo "github.com/fromanirh/numainfo_exporter/internal/pkg/version"

	"github.com/fromanirh/numainfo_exporter/pkg/reader/kubeletcheckpoint"
	"github.com/fromanirh/numainfo_exporter/pkg/reader/sysfs"
)

var (
	// see https://www.robustperception.io/exposing-the-software-version-to-prometheus
	version = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "numainfo_version",
			Help: "Version information",
			ConstLabels: prometheus.Labels{
				"branch":      verinfo.Branch,
				"goversion":   runtime.Version(),
				"revision":    verinfo.Revision,
				"kubeversion": verinfo.Version, // TODO: clarify what this really represents
				"version":     "1",
			},
		},
	)

	// TODO: maybe redundant?
	numaNodesDesc = prometheus.NewDesc(
		"numainfo_node_count",
		"NUMA nodes per node, count.",
		[]string{
			"node",
		},
		nil,
	)

	coreCountDesc = prometheus.NewDesc(
		"numainfo_core_count",
		"CPU cores per NUMA node, count.",
		[]string{
			"node",
			"numanode",
			"type", // "capacity" or "allocation"
		},
		nil,
	)

	pciDeviceCountDesc = prometheus.NewDesc(
		"numainfo_pcidevice_count",
		"PCI device resources per NUMA node, count.",
		[]string{
			"node",
			"numanode",
			"resource",
			"type", // "capacity" or "allocation"
		},
		nil,
	)
)

func init() {
	prometheus.MustRegister(version)

	version.Set(1)
}

type Collector struct {
	nodeName    string
	sysFsDir    string
	kubeletInfo *kubeletcheckpoint.Reader
	// pciResources: pciaddr -> resource
	pciResources map[string]string
}

func NewCollector(nodeName, sysFsDir string, klReader *kubeletcheckpoint.Reader) (*Collector, error) {
	return &Collector{
		nodeName:     nodeName,
		sysFsDir:     sysFsDir,
		kubeletInfo:  klReader,
		pciResources: make(map[string]string),
	}, nil
}

func (co Collector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(co, ch)
}

// Note that Collect could be called concurrently
func (co Collector) Collect(ch chan<- prometheus.Metric) {
	tp, err := sysfs.NewTopologyFromSysFs(co.sysFsDir, co.pciResources)
	if err != nil {
		klog.Warningf("failed to extract topology from sysfs (at %s): %v", co.sysFsDir, err)
		return
	}

	coresAlloc, err := co.kubeletInfo.GetCoresAllocation(tp)
	if err != nil {
		klog.Warningf("failed to get the current cores allocation: %v", err)
		return
	}

	m, err := prometheus.NewConstMetric(
		numaNodesDesc,
		prometheus.GaugeValue,
		float64(tp.NUMANodeCount),
		co.nodeName,
	)
	if err != nil {
		klog.Warningf("failed to create the NUMA nodes metric: %v", err)
		return
	}
	ch <- m

	for _, node := range tp.NUMANodes {
		m, err := prometheus.NewConstMetric(
			coreCountDesc,
			prometheus.GaugeValue,
			float64(len(node.Cores)),
			co.nodeName,
			fmt.Sprintf("node%02d", node.Id),
			"capacity",
		)
		if err != nil {
			klog.Warningf("failed to create the CPU cores capacity: %v", err)
			continue
		}
		ch <- m

		m, err = prometheus.NewConstMetric(
			coreCountDesc,
			prometheus.GaugeValue,
			float64(coresAlloc[node.Id]),
			co.nodeName,
			fmt.Sprintf("node%02d", node.Id),
			"allocation",
		)
		if err != nil {
			klog.Warningf("failed to create the CPU cores allocation: %v", err)
			continue
		}
		ch <- m
	}
}
