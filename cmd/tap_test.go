package cmd

import (
	"strings"
	"testing"

	"github.com/kubeshark/kubeshark/config/configStructs"
)

func TestTLSFlagDocumentsBoringSSLSupport(t *testing.T) {
	flag := tapCmd.Flags().Lookup(configStructs.TlsLabel)
	if flag == nil {
		t.Fatalf("expected --%s flag to be registered", configStructs.TlsLabel)
	}

	for _, supportedLibrary := range []string{"OpenSSL", "BoringSSL", "Envoy/Istio", "Go crypto/tls"} {
		if !strings.Contains(flag.Usage, supportedLibrary) {
			t.Errorf("expected --%s help to mention %q; got %q", configStructs.TlsLabel, supportedLibrary, flag.Usage)
		}
	}
}
