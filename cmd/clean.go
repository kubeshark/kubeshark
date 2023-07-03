package cmd

import (
	"fmt"

	"github.com/creasty/defaults"
	"github.com/kubeshark/kubeshark/config/configStructs"
	"github.com/kubeshark/kubeshark/kubernetes/helm"
	"github.com/kubeshark/kubeshark/misc"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: fmt.Sprintf("Removes all %s resources", misc.Software),
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := helm.NewHelmDefault().Uninstall()
		if err != nil {
			log.Error().Err(err).Send()
		} else {
			log.Info().Msgf("Uninstalled the Helm release: %s", resp.Release.Name)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)

	defaultTapConfig := configStructs.TapConfig{}
	if err := defaults.Set(&defaultTapConfig); err != nil {
		log.Debug().Err(err).Send()
	}

	cleanCmd.Flags().StringP(configStructs.SelfNamespaceLabel, "s", defaultTapConfig.SelfNamespace, "Self-namespace of Kubeshark")
}
