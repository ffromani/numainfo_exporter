#!/bin/sh

set -e

TAG=${1:-devel}
#TODO: validate tag is vX.Y.Z
VERSIONDIR="internal/pkg/version"
VERSIONFILE="${VERSIONDIR}/version.go"

mkdir -p ${VERSIONDIR} && ./hack/build/genver.sh ${TAG} > ${VERSIONFILE}
