package kubernetes

import (
	"testing"

	"github.com/kubeshark/kubeshark/semver"
)

func TestValidateKubernetesVersionRejectsNumericVersionBelowMinimum(t *testing.T) {
	serverVersion := semver.SemVersion("v1.9.0")

	if err := ValidateKubernetesVersion(&serverVersion); err == nil {
		t.Fatalf("expected Kubernetes version %s to be rejected", serverVersion)
	}
}

func TestValidateKubernetesVersionAcceptsMinimumVersion(t *testing.T) {
	serverVersion := semver.SemVersion(MinKubernetesServerVersion)

	if err := ValidateKubernetesVersion(&serverVersion); err != nil {
		t.Fatalf("expected Kubernetes version %s to be accepted: %v", serverVersion, err)
	}
}
