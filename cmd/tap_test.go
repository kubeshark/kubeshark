package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/pflag"

	"github.com/kubeshark/kubeshark/config/configStructs"
)

func TestTLSFlagConfiguration(t *testing.T) {
	tlsFlag := requireTLSFlag(t)

	t.Run("is a Boolean flag", func(t *testing.T) {
		if got, want := tlsFlag.Value.Type(), "bool"; got != want {
			t.Errorf("unexpected --%s type: got %q, want %q", configStructs.TlsLabel, got, want)
		}
	})

	t.Run("is enabled by default", func(t *testing.T) {
		if got, want := tlsFlag.DefValue, "true"; got != want {
			t.Errorf("unexpected --%s default: got %q, want %q", configStructs.TlsLabel, got, want)
		}
	})

	t.Run("can be enabled without an explicit value", func(t *testing.T) {
		if got, want := tlsFlag.NoOptDefVal, "true"; got != want {
			t.Errorf("unexpected bare --%s value: got %q, want %q", configStructs.TlsLabel, got, want)
		}
	})
}

func TestTLSFlagUsage(t *testing.T) {
	tlsFlag := requireTLSFlag(t)

	t.Run("uses the canonical usage text", func(t *testing.T) {
		if got, want := tlsFlag.Usage, tlsFlagUsage; got != want {
			t.Errorf("unexpected --%s usage: got %q, want %q", configStructs.TlsLabel, got, want)
		}
	})

	t.Run("renders the usage text in command help", func(t *testing.T) {
		if help := tapCmd.Flags().FlagUsages(); !strings.Contains(help, tlsFlagUsage) {
			t.Errorf("expected tap help to contain %q; got %q", tlsFlagUsage, help)
		}
	})
}

func TestTLSFlagUsageDocumentsSupportedTechnologies(t *testing.T) {
	tlsFlag := requireTLSFlag(t)

	supportedTechnologies := []struct {
		name string
		text string
	}{
		{name: "OpenSSL", text: "OpenSSL"},
		{name: "BoringSSL", text: "BoringSSL"},
		{name: "Envoy", text: "Envoy"},
		{name: "Istio", text: "Istio"},
		{name: "Go crypto TLS", text: "Go crypto/tls"},
	}

	for _, technology := range supportedTechnologies {
		t.Run("documents "+technology.name, func(t *testing.T) {
			if !strings.Contains(tlsFlag.Usage, technology.text) {
				t.Errorf("expected --%s help to mention %q; got %q", configStructs.TlsLabel, technology.text, tlsFlag.Usage)
			}
		})
	}
}

func requireTLSFlag(t *testing.T) *pflag.Flag {
	t.Helper()

	tlsFlag := tapCmd.Flags().Lookup(configStructs.TlsLabel)
	if tlsFlag == nil {
		t.Fatalf("expected --%s flag to be registered", configStructs.TlsLabel)
	}

	return tlsFlag
}
