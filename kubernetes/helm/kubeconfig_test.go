package helm

import (
	"os"
	"path/filepath"
	"testing"
)

func writeKubeConfig(t *testing.T, dir string, name string, server string) string {
	t.Helper()

	content := `apiVersion: v1
kind: Config
current-context: ` + name + `
clusters:
- name: ` + name + `
  cluster:
    server: ` + server + `
contexts:
- name: ` + name + `
  context:
    cluster: ` + name + `
    user: ` + name + `
users:
- name: ` + name + `
  user:
    token: test
`
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	return path
}

func TestRESTClientGetterMergesMultipleKubeConfigPaths(t *testing.T) {
	dir := t.TempDir()
	first := writeKubeConfig(t, dir, "first", "https://first.example.com")
	second := writeKubeConfig(t, dir, "second", "https://second.example.com")

	getter := newRESTClientGetter(first+string(os.PathListSeparator)+second, "kubeshark")

	rawConfig, err := getter.ToRawKubeConfigLoader().RawConfig()
	if err != nil {
		t.Fatalf("expected a list of kubeconfig paths to load: %v", err)
	}
	for _, name := range []string{"first", "second"} {
		if _, ok := rawConfig.Contexts[name]; !ok {
			t.Errorf("expected context %q from the merged kubeconfig files", name)
		}
	}

	restConfig, err := getter.ToRESTConfig()
	if err != nil {
		t.Fatalf("expected a REST config from the merged kubeconfig files: %v", err)
	}
	if restConfig.Host != "https://first.example.com" {
		t.Errorf("expected the first file's current context, got host %q", restConfig.Host)
	}
}

func TestRESTClientGetterUsesNamespace(t *testing.T) {
	dir := t.TempDir()
	path := writeKubeConfig(t, dir, "only", "https://only.example.com")

	namespace, _, err := newRESTClientGetter(path, "kubeshark").ToRawKubeConfigLoader().Namespace()
	if err != nil {
		t.Fatal(err)
	}
	if namespace != "kubeshark" {
		t.Errorf("expected namespace %q, got %q", "kubeshark", namespace)
	}
}
