/*
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2020 Red Hat, Inc.
 */

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
