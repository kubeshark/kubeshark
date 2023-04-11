package cmd

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/kubeshark/kubeshark/utils"
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
		workerDaemonSet,
		err := generateManifests()
	if err != nil {
		log.Error().Err(err).Send()
		return
	}

	err = dumpHelmChart(map[string]interface{}{
		"00-namespace.yaml":            namespace,
		"01-service-account.yaml":      serviceAccount,
		"02-cluster-role.yaml":         clusterRole,
		"03-cluster-role-binding.yaml": clusterRoleBinding,
		"04-hub-pod.yaml":              hubPod,
		"05-hub-service.yaml":          hubService,
		"06-front-pod.yaml":            frontPod,
		"07-front-service.yaml":        frontService,
		"08-worker-daemon-set.yaml":    workerDaemonSet,
	})
	if err != nil {
		log.Error().Err(err).Send()
		return
	}
}

func dumpHelmChart(objects map[string]interface{}) error {
	folder := filepath.Join(".", "helm-chart/templates")
	err := os.MkdirAll(folder, os.ModePerm)
	if err != nil {
		return err
	}

	// Sort by filenames
	filenames := make([]string, 0)
	for filename := range objects {
		filenames = append(filenames, filename)
	}
	sort.Strings(filenames)

	for _, filename := range filenames {
		manifest, err := utils.PrettyYamlOmitEmpty(objects[filename])
		if err != nil {
			return err
		}

		path := filepath.Join(folder, filename)
		err = os.WriteFile(path, []byte(manifest), 0644)
		if err != nil {
			return err
		}
		log.Info().Msgf("Helm chart template generated: %s", path)
	}

	return nil
}
