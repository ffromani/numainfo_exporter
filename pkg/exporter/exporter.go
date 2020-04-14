/*
 * This file is part of the numainfo_exporter project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by explicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2020 Red Hat, Inc.
 */

package exporter

import (
	"net/http"
	"os"
	"runtime"

	"github.com/spf13/pflag"

	"k8s.io/klog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/fromanirh/numainfo_exporter/pkg/reader/kubeletcheckpoint"

	"github.com/fromanirh/numainfo_exporter/internal/pkg/k8sutils"
	verinfo "github.com/fromanirh/numainfo_exporter/internal/pkg/version"
)

const (
	defaultPort = 8443
	defaultHost = "0.0.0.0"
)

type Exporter struct {
	TLSInfo       *k8sutils.TLSInfo
	ListenAddress string
	DumpMode      bool
	Nodename      string
	SysFSDir      string
	KubeStateDir  string
}

func printVersion() {
	klog.Infof("Server Version: %s", verinfo.Version)
	klog.Infof("Go Version: %s", runtime.Version())
	klog.Infof("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
}

func NewExporter() *Exporter {
	return &Exporter{
		TLSInfo: &k8sutils.TLSInfo{},
	}
}

func getHostname() (string, error) {
	var err error
	var hostname string

	if hostname = os.Getenv("NUMAINFO_NODENAME"); hostname != "" {
		return hostname, nil
	}
	if hostname, err = os.Hostname(); err == nil {
		return hostname, nil
	}
	return "", err
}

func (exp *Exporter) ParseFlags() error {
	pflag.StringVarP(&exp.TLSInfo.CertFilePath, "cert-file", "c", "", "override path to TLS certificate - you need also the key to enable TLS")
	pflag.StringVarP(&exp.TLSInfo.KeyFilePath, "key-file", "k", "", "override path to TLS key - you need also the cert to enable TLS")
	pflag.StringVarP(&exp.ListenAddress, "listen-address", "L", ":19091", "listening address for the server")
	pflag.BoolVarP(&exp.DumpMode, "dump-metrics", "M", false, "dump the available metrics and exit")
	pflag.StringVarP(&exp.Nodename, "node-name", "N", "", "node identifier.")
	pflag.StringVarP(&exp.SysFSDir, "sysfs", "Y", "/sys", "base root directory where sysfs is mounted")
	pflag.StringVarP(&exp.KubeStateDir, "kube-state", "K", "/var/lib/kubelet", "base root directory where the kubelet state files are stored")

	pflag.Parse()

	if exp.Nodename == "" {
		hostname, err := getHostname()
		if err != nil {
			return err
		}
		exp.Nodename = hostname
	}
	return nil
}

func (exp *Exporter) Run() {
	printVersion()

	exp.TLSInfo.UpdateFromK8S()
	defer exp.TLSInfo.Clean()

	var err error
	klReader, err := kubeletcheckpoint.NewReader(exp.KubeStateDir)
	if err != nil {
		klog.V(1).Infof("error creating the kubelet check point reader: %v", err)
		os.Exit(1)
	}

	co, err := NewCollector(exp.Nodename, exp.SysFSDir, klReader)
	if err != nil {
		klog.V(1).Infof("error creating the collector: %v", err)
		os.Exit(1)
	}
	prometheus.MustRegister(co)

	if exp.DumpMode {
		DumpMetrics(os.Stderr)
		os.Exit(1)
	}

	klog.V(2).Info("%s started", verinfo.Component)
	defer klog.V(2).Infof("%s stopped", verinfo.Component)

	http.Handle("/metrics", promhttp.Handler())
	if exp.TLSInfo.IsEnabled() {
		klog.V(1).Infof("TLS configured, serving over HTTPS")
		klog.V(1).Infof("%s", http.ListenAndServeTLS(exp.ListenAddress, exp.TLSInfo.CertFilePath, exp.TLSInfo.KeyFilePath, nil))
	} else {
		klog.V(1).Infof("TLS *NOT* configured, serving over HTTP")
		klog.V(1).Infof("%s", http.ListenAndServe(exp.ListenAddress, nil))
	}
}
