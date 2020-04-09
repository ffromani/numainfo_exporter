package main

import (
	"flag"

	"k8s.io/klog"

	"github.com/spf13/pflag"

	"github.com/fromanirh/numainfo_exporter/pkg/exporter"
)

func main() {
	// Add klog flags
	klog.InitFlags(nil)

	// Add flags registered by imported packages
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	app := exporter.NewExporter()
	err := app.ParseFlags()
	if err != nil {
		klog.Exit("error parsing flags: %v", err)
	}

	app.Run()
}
