package k8sutils

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// ServiceAccountNamespaceFile is the default service account namespace file path
const ServiceAccountNamespaceFile = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"

// DefaultNamespace is the fallback namespace
const DefaultNamespace = "default"

// GetNamespace returns the current namespace on which the server is running
func GetNamespace() (string, error) {
	if data, err := ioutil.ReadFile(ServiceAccountNamespaceFile); err == nil {
		if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
			return ns, nil
		}
	} else if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to determine namespace from %s: %v", ServiceAccountNamespaceFile, err)
	}
	return DefaultNamespace, nil
}
