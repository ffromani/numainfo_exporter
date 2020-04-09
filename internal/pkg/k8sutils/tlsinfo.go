package k8sutils

import (
	"io/ioutil"
	"os"

	"k8s.io/klog"

	"k8s.io/client-go/rest"

	"github.com/fromanirh/numainfo_exporter/internal/pkg/certutil"
	"github.com/fromanirh/numainfo_exporter/internal/pkg/version"
)

// TLSInfo holds the certs and key data
type TLSInfo struct {
	CertFilePath   string
	KeyFilePath    string
	certsDirectory string
}

// Clean removes all data managed by this TLSInfo instance
func (ti *TLSInfo) Clean() {
	if ti.certsDirectory != "" {
		os.RemoveAll(ti.certsDirectory)
		ti.certsDirectory = ""
		klog.V(4).Infof("TLSInfo cleaned up!")
	}
}

// IsEnabled tells if TLS is enabled
func (ti *TLSInfo) IsEnabled() bool {
	return ti.CertFilePath != "" && ti.KeyFilePath != ""
}

// UpdateFromK8S fetches all the relevant config from K8S
func (ti *TLSInfo) UpdateFromK8S() error {
	var err error
	if _, err = rest.InClusterConfig(); err != nil {
		// is not a real error, rather a supported case. So, let's swallow the error
		klog.V(3).Infof("running outside a K8S cluster")
		return nil
	}
	if ti.IsEnabled() {
		klog.V(3).Infof("TLSInfo already fully set")
		return nil
	}

	// at least one between cert and key need to be set
	ti.certsDirectory, err = ioutil.TempDir("", "certsdir")
	if err != nil {
		return err
	}
	namespace, err := GetNamespace()
	if err != nil {
		klog.Warningf("Error searching for namespace: %v", err)
		return err
	}
	certStore, err := certutil.GenerateSelfSignedCert(ti.certsDirectory, version.Component, namespace)
	if err != nil {
		klog.Warningf("unable to generate certificates: %v", err)
		return err
	}

	if ti.CertFilePath == "" {
		ti.CertFilePath = certStore.CurrentPath()
	} else {
		klog.V(2).Infof("NOT overriding cert file %s with %s", ti.CertFilePath, certStore.CurrentPath())
	}
	if ti.KeyFilePath == "" {
		ti.KeyFilePath = certStore.CurrentPath()
	} else {
		klog.V(2).Infof("NOT overriding key file %s with %s", ti.KeyFilePath, certStore.CurrentPath())
	}
	klog.V(3).Infof("running in a K8S cluster: with configuration %#v", *ti)
	return nil
}
