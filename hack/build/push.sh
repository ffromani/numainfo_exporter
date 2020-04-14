#!/bin/bash

set -e

VERSION="${1:-devel}"

podman build -t quay.io/fromani/numainfo_exporter:$VERSION .
podman push quay.io/fromani/numainfo_exporter:$VERSION

