package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/kubernetes"
	"github.com/kubeshark/kubeshark/misc"
	"github.com/kubeshark/kubeshark/misc/fsUtils"
	"github.com/kubeshark/kubeshark/utils"
	"github.com/ohler55/ojg/jp"
	"github.com/ohler55/ojg/oj"
	"github.com/otiai10/copy"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var helmChartCmd = &cobra.Command{
	Use:   "helm-chart",
	Short: "Generate Helm chart of Kubeshark",
	RunE: func(cmd *cobra.Command, args []string) error {
		runHelmChart()
		return nil
	},
}

// Maintainer describes a Chart maintainer.
type Maintainer struct {
	// Name is a user name or organization name
	Name string `json:"name,omitempty"`
	// Email is an optional email address to contact the named maintainer
	Email string `json:"email,omitempty"`
	// URL is an optional URL to an address for the named maintainer
	URL string `json:"url,omitempty"`
}

// Metadata for a Chart file. This models the structure of a Chart.yaml file.
type Metadata struct {
	// The name of the chart. Required.
	Name string `json:"name,omitempty"`
	// The URL to a relevant project page, git repo, or contact person
	Home string `json:"home,omitempty"`
	// Source is the URL to the source code of this chart
	Sources []string `json:"sources,omitempty"`
	// A SemVer 2 conformant version string of the chart. Required.
	Version string `json:"version,omitempty"`
	// A one-sentence description of the chart
	Description string `json:"description,omitempty"`
	// A list of string keywords
	Keywords []string `json:"keywords,omitempty"`
	// A list of name and URL/email address combinations for the maintainer(s)
	Maintainers []*Maintainer `json:"maintainers,omitempty"`
	// The URL to an icon file.
	Icon string `json:"icon,omitempty"`
	// The API Version of this chart. Required.
	APIVersion string `json:"apiVersion,omitempty"`
	// The condition to check to enable chart
	Condition string `json:"condition,omitempty"`
	// The tags to check to enable chart
	Tags string `json:"tags,omitempty"`
	// The version of the application enclosed inside of this chart.
	AppVersion string `json:"appVersion,omitempty"`
	// Whether or not this chart is deprecated
	Deprecated bool `json:"deprecated,omitempty"`
	// Annotations are additional mappings uninterpreted by Helm,
	// made available for inspection by other applications.
	Annotations map[string]string `json:"annotations,omitempty"`
	// KubeVersion is a SemVer constraint specifying the version of Kubernetes required.
	KubeVersion string `json:"kubeVersion,omitempty"`
	// Dependencies are a list of dependencies for a chart.
	Dependencies []*Dependency `json:"dependencies,omitempty"`
	// Specifies the chart type: application or library
	Type string `json:"type,omitempty"`
}

// Dependency describes a chart upon which another chart depends.
//
// Dependencies can be used to express developer intent, or to capture the state
// of a chart.
type Dependency struct {
	// Name is the name of the dependency.
	//
	// This must mach the name in the dependency's Chart.yaml.
	Name string `json:"name"`
	// Version is the version (range) of this chart.
	//
	// A lock file will always produce a single version, while a dependency
	// may contain a semantic version range.
	Version string `json:"version,omitempty"`
	// The URL to the repository.
	//
	// Appending `index.yaml` to this string should result in a URL that can be
	// used to fetch the repository index.
	Repository string `json:"repository"`
	// A yaml path that resolves to a boolean, used for enabling/disabling charts (e.g. subchart1.enabled )
	Condition string `json:"condition,omitempty"`
	// Tags can be used to group charts for enabling/disabling together
	Tags []string `json:"tags,omitempty"`
	// Enabled bool determines if chart should be loaded
	Enabled bool `json:"enabled,omitempty"`
	// ImportValues holds the mapping of source values to parent key to be imported. Each item can be a
	// string or pair of child/parent sublist items.
	ImportValues []interface{} `json:"import-values,omitempty"`
	// Alias usable alias to be used for the chart
	Alias string `json:"alias,omitempty"`
}

var namespaceMappings = map[string]interface{}{
	"metadata.name": "{{ .Values.tap.selfnamespace }}",
}
var serviceAccountMappings = map[string]interface{}{
	"metadata.namespace": "{{ .Values.tap.selfnamespace }}",
}
var clusterRoleMappings = serviceAccountMappings
var clusterRoleBindingMappings = map[string]interface{}{
	"metadata.namespace":    "{{ .Values.tap.selfnamespace }}",
	"subjects[0].namespace": "{{ .Values.tap.selfnamespace }}",
}
var hubPodMappings = map[string]interface{}{
	"metadata.namespace": "{{ .Values.tap.selfnamespace }}",
	"spec.containers[0].env": []map[string]interface{}{
		{
			"name":  "POD_REGEX",
			"value": "{{ .Values.tap.regex }}",
		},
		{
			"name":  "NAMESPACES",
			"value": "{{ gt (len .Values.tap.namespaces) 0 | ternary (join \",\" .Values.tap.namespaces) \"\" }}",
		},
		{
			"name":  "LICENSE",
			"value": "{{ .Values.license }}",
		},
		{
			"name":  "SCRIPTING_ENV",
			"value": "{}",
		},
		{
			"name":  "SCRIPTING_SCRIPTS",
			"value": "[]",
		},
		{
			"name":  "AUTH_APPROVED_DOMAINS",
			"value": "{{ gt (len .Values.tap.ingress.auth.approvedDomains) 0 | ternary (join \",\" .Values.tap.ingress.auth.approvedDomains) \"\" }}",
		},
	},
	"spec.containers[0].image":                     "{{ .Values.tap.docker.registry }}/hub:{{ .Values.tap.docker.tag }}",
	"spec.containers[0].imagePullPolicy":           "{{ .Values.tap.docker.imagepullpolicy }}",
	"spec.containers[0].resources.limits.cpu":      "{{ .Values.tap.resources.hub.limits.cpu }}",
	"spec.containers[0].resources.limits.memory":   "{{ .Values.tap.resources.hub.limits.memory }}",
	"spec.containers[0].resources.requests.cpu":    "{{ .Values.tap.resources.hub.requests.cpu }}",
	"spec.containers[0].resources.requests.memory": "{{ .Values.tap.resources.hub.requests.memory }}",
	"spec.containers[0].command[0]":                "{{ .Values.tap.debug | ternary \"./hub -debug\" \"./hub\" }}",
}
var hubServiceMappings = serviceAccountMappings
var frontPodMappings = map[string]interface{}{
	"metadata.namespace":                 "{{ .Values.tap.selfnamespace }}",
	"spec.containers[0].image":           "{{ .Values.tap.docker.registry }}/front:{{ .Values.tap.docker.tag }}",
	"spec.containers[0].imagePullPolicy": "{{ .Values.tap.docker.imagepullpolicy }}",
}
var frontServiceMappings = serviceAccountMappings
var persistentVolumeMappings = map[string]interface{}{
	"metadata.namespace":              "{{ .Values.tap.selfnamespace }}",
	"spec.resources.requests.storage": "{{ .Values.tap.storagelimit }}",
	"spec.storageClassName":           "{{ .Values.tap.storageclass }}",
}
var workerDaemonSetMappings = map[string]interface{}{
	"metadata.namespace":                                         "{{ .Values.tap.selfnamespace }}",
	"spec.template.spec.containers[0].image":                     "{{ .Values.tap.docker.registry }}/worker:{{ .Values.tap.docker.tag }}",
	"spec.template.spec.containers[0].imagePullPolicy":           "{{ .Values.tap.docker.imagepullpolicy }}",
	"spec.template.spec.containers[0].resources.limits.cpu":      "{{ .Values.tap.resources.worker.limits.cpu }}",
	"spec.template.spec.containers[0].resources.limits.memory":   "{{ .Values.tap.resources.worker.limits.memory }}",
	"spec.template.spec.containers[0].resources.requests.cpu":    "{{ .Values.tap.resources.worker.requests.cpu }}",
	"spec.template.spec.containers[0].resources.requests.memory": "{{ .Values.tap.resources.worker.requests.memory }}",
	"spec.template.spec.containers[0].command[0]":                "{{ .Values.tap.debug | ternary \"./worker -debug\" \"./worker\" }}",
	"spec.template.spec.containers[0].command[4]":                "{{ .Values.tap.proxy.worker.srvport }}",
	"spec.template.spec.containers[0].command[6]":                "{{ .Values.tap.packetcapture }}",
}
var ingressClassMappings = serviceAccountMappings
var ingressMappings = map[string]interface{}{
	"metadata.namespace": "{{ .Values.tap.selfnamespace }}",
	"metadata.annotations[\"certmanager.k8s.io/cluster-issuer\"]": "{{ .Values.tap.ingress.certManager }}",
	"spec.rules[0].host": "{{ .Values.tap.ingress.host }}",
	"spec.tls":           "{{ .Values.tap.ingress.tls | toYaml }}",
}

func init() {
	rootCmd.AddCommand(helmChartCmd)
}

func runHelmChart() {
	namespace,
		serviceAccount,
		clusterRole,
		clusterRoleBinding,
		hubPod,
		hubService,
		frontPod,
		frontService,
		persistentVolume,
		workerDaemonSet,
		ingressClass,
		ingress,
		err := generateManifests()
	if err != nil {
		log.Error().Err(err).Send()
		return
	}

	err = dumpHelmChart(map[string]interface{}{
		"00-namespace.yaml":               template(namespace, namespaceMappings),
		"01-service-account.yaml":         template(serviceAccount, serviceAccountMappings),
		"02-cluster-role.yaml":            template(clusterRole, clusterRoleMappings),
		"03-cluster-role-binding.yaml":    template(clusterRoleBinding, clusterRoleBindingMappings),
		"04-hub-pod.yaml":                 template(hubPod, hubPodMappings),
		"05-hub-service.yaml":             template(hubService, hubServiceMappings),
		"06-front-pod.yaml":               template(frontPod, frontPodMappings),
		"07-front-service.yaml":           template(frontService, frontServiceMappings),
		"08-persistent-volume-claim.yaml": template(persistentVolume, persistentVolumeMappings),
		"09-worker-daemon-set.yaml":       template(workerDaemonSet, workerDaemonSetMappings),
		"10-ingress-class.yaml":           template(ingressClass, ingressClassMappings),
		"11-ingress.yaml":                 template(ingress, ingressMappings),
	})
	if err != nil {
		log.Error().Err(err).Send()
		return
	}
}

func template(object interface{}, mappings map[string]interface{}) (template interface{}) {
	var err error
	var data []byte
	data, err = json.Marshal(object)
	if err != nil {
		log.Error().Err(err).Send()
		return
	}

	var obj interface{}
	obj, err = oj.Parse(data)
	if err != nil {
		log.Error().Err(err).Send()
		return
	}

	for path, value := range mappings {
		var x jp.Expr
		x, err = jp.ParseString(path)
		if err != nil {
			log.Error().Err(err).Send()
			return
		}

		err = x.Set(obj, value)
		if err != nil {
			log.Error().Err(err).Send()
			return
		}
	}

	newJson := oj.JSON(obj)

	err = json.Unmarshal([]byte(newJson), &template)
	if err != nil {
		log.Error().Err(err).Send()
		return
	}

	return
}

func handleHubPod(manifest string) string {
	lines := strings.Split(manifest, "\n")

	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "hostPort:") {
			lines[i] = "          hostPort: {{ .Values.tap.proxy.hub.srvport }}"
		}
	}

	return strings.Join(lines, "\n")
}

func handleFrontPod(manifest string) string {
	lines := strings.Split(manifest, "\n")

	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "hostPort:") {
			lines[i] = "          hostPort: {{ .Values.tap.proxy.front.srvport }}"
		}
	}

	return strings.Join(lines, "\n")
}

func handlePVCManifest(manifest string) string {
	return fmt.Sprintf("{{- if .Values.tap.persistentstorage }}\n%s{{- end }}\n", manifest)
}

func handleDaemonSetManifest(manifest string) string {
	lines := strings.Split(manifest, "\n")

	for i, line := range lines {
		if strings.TrimSpace(line) == "- mountPath: /app/data" {
			lines[i] = fmt.Sprintf("{{- if .Values.tap.persistentstorage }}\n%s", line)
		}

		if strings.TrimSpace(line) == "name: kubeshark-persistent-volume" {
			lines[i] = fmt.Sprintf("%s\n{{- end }}", line)
		}

		if strings.TrimSpace(line) == "- name: kubeshark-persistent-volume" {
			lines[i] = fmt.Sprintf("{{- if .Values.tap.persistentstorage }}\n%s", line)
		}

		if strings.TrimSpace(line) == "claimName: kubeshark-persistent-volume-claim" {
			lines[i] = fmt.Sprintf("%s\n{{- end }}", line)
		}

		if strings.HasPrefix(strings.TrimSpace(line), "- containerPort:") {
			lines[i] = "            - containerPort: {{ .Values.tap.proxy.worker.srvport }}"
		}

		if strings.HasPrefix(strings.TrimSpace(line), "hostPort:") {
			lines[i] = "              hostPort: {{ .Values.tap.proxy.worker.srvport }}"
		}
	}

	return strings.Join(lines, "\n")
}

func handleIngressClass(manifest string) string {
	return fmt.Sprintf("{{- if .Values.tap.ingress.enabled }}\n%s{{- end }}\n", manifest)
}

func handleIngress(manifest string) string {
	manifest = strings.Replace(manifest, "'{{ .Values.tap.ingress.tls | toYaml }}'", "{{ .Values.tap.ingress.tls | toYaml }}", 1)

	return handleIngressClass(manifest)
}

func dumpHelmChart(objects map[string]interface{}) error {
	folder := filepath.Join(".", "helm-chart")
	templatesFolder := filepath.Join(folder, "templates")

	err := fsUtils.RemoveFilesByExtension(templatesFolder, "yaml")
	if err != nil {
		return err
	}

	err = os.MkdirAll(templatesFolder, os.ModePerm)
	if err != nil {
		return err
	}

	// Sort by filenames
	filenames := make([]string, 0)
	for filename := range objects {
		filenames = append(filenames, filename)
	}
	sort.Strings(filenames)

	// Generate templates
	for _, filename := range filenames {
		manifest, err := utils.PrettyYamlOmitEmpty(objects[filename])
		if err != nil {
			return err
		}

		if filename == "04-hub-pod.yaml" {
			manifest = handleHubPod(manifest)
		}

		if filename == "06-front-pod.yaml" {
			manifest = handleFrontPod(manifest)
		}

		if filename == "08-persistent-volume-claim.yaml" {
			manifest = handlePVCManifest(manifest)
		}

		if filename == "09-worker-daemon-set.yaml" {
			manifest = handleDaemonSetManifest(manifest)
		}

		if filename == "10-ingress-class.yaml" {
			manifest = handleIngressClass(manifest)
		}

		if filename == "11-ingress.yaml" {
			manifest = handleIngress(manifest)
		}

		path := filepath.Join(templatesFolder, filename)
		err = os.WriteFile(path, []byte(manifestHeader+manifest), 0644)
		if err != nil {
			return err
		}
		log.Info().Msgf("Helm chart template generated: %s", path)
	}

	// Copy LICENSE
	licenseSrcPath := filepath.Join(".", "LICENSE")
	licenseDstPath := filepath.Join(folder, "LICENSE")
	err = copy.Copy(licenseSrcPath, licenseDstPath)
	if err != nil {
		log.Warn().Err(err).Str("path", licenseSrcPath).Msg("Couldn't find the license:")
	} else {
		log.Info().Msgf("Helm chart license copied: %s", licenseDstPath)
	}

	// Generate Chart.yaml
	chartMetadata := Metadata{
		APIVersion:  "v2",
		Name:        misc.Program,
		Description: misc.Description,
		Home:        misc.Website,
		Sources:     []string{"https://github.com/kubeshark/kubeshark/tree/master/helm-chart"},
		Keywords: []string{
			"kubeshark",
			"packet capture",
			"traffic capture",
			"traffic analyzer",
			"network sniffer",
			"observability",
			"devops",
			"microservice",
			"forensics",
			"api",
		},
		Maintainers: []*Maintainer{
			{
				Name:  misc.Software,
				Email: misc.Email,
				URL:   misc.Website,
			},
		},
		Version:     misc.Ver,
		AppVersion:  misc.Ver,
		KubeVersion: fmt.Sprintf(">= %s-0", kubernetes.MinKubernetesServerVersion),
		Type:        "application",
	}

	chart, err := utils.PrettyYamlOmitEmpty(chartMetadata)
	if err != nil {
		return err
	}

	path := filepath.Join(folder, "Chart.yaml")
	err = os.WriteFile(path, []byte(chart), 0644)
	if err != nil {
		return err
	}
	log.Info().Msgf("Helm chart Chart.yaml generated: %s", path)

	// Generate values.yaml
	values, err := utils.PrettyYaml(config.Config)
	if err != nil {
		return err
	}
	path = filepath.Join(folder, "values.yaml")
	err = os.WriteFile(path, []byte(values), 0644)
	if err != nil {
		return err
	}
	log.Info().Msgf("Helm chart values.yaml generated: %s", path)

	return nil
}
