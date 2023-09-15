package cmd

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/creasty/defaults"
	"github.com/kubeshark/kubeshark/config/configStructs"
	"github.com/kubeshark/kubeshark/internal/connect"
	"github.com/kubeshark/kubeshark/kubernetes"
	"github.com/kubeshark/kubeshark/utils"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Exports the captured traffic into a TAR file that contains PCAP files",
	RunE: func(cmd *cobra.Command, args []string) error {
		runExport()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)

	defaultTapConfig := configStructs.TapConfig{}
	if err := defaults.Set(&defaultTapConfig); err != nil {
		log.Debug().Err(err).Send()
	}

	exportCmd.Flags().Uint16(configStructs.ProxyFrontPortLabel, defaultTapConfig.Proxy.Front.Port, "Provide a custom port for the Kubeshark")
	exportCmd.Flags().String(configStructs.ProxyHostLabel, defaultTapConfig.Proxy.Host, "Provide a custom host for the Kubeshark")
	exportCmd.Flags().StringP(configStructs.ReleaseNamespaceLabel, "s", defaultTapConfig.Release.Namespace, "Release namespace of Kubeshark")
}

func runExport() {
	hubUrl := kubernetes.GetHubUrl()
	response, err := http.Get(fmt.Sprintf("%s/echo", hubUrl))
	if err != nil || response.StatusCode != 200 {
		log.Info().Msg(fmt.Sprintf(utils.Yellow, "Couldn't connect to Hub. Establishing proxy..."))
		runProxy(false, true)
	}

	dstPath, err := filepath.Abs(fmt.Sprintf("./%d.tar.gz", time.Now().Unix()))
	if err != nil {
		panic(err)
	}

	out, err := os.Create(dstPath)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	connector := connect.NewConnector(kubernetes.GetHubUrl(), connect.DefaultRetries, connect.DefaultTimeout)
	connector.PostPcapsMerge(out)
}
