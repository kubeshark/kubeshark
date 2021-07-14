package cmd

import (
	"errors"
	"fmt"
	"github.com/up9inc/mizu/shared/units"
	"regexp"
	"strings"

	"github.com/up9inc/mizu/cli/mizu"

	"github.com/spf13/cobra"
)

type MizuTapOptions struct {
	GuiPort                uint16
	Namespace              string
	AllNamespaces          bool
	Analyze                bool
	AnalyzeDestination     string
	KubeConfigPath         string
	MizuImage              string
	PlainTextFilterRegexes []string
	TapOutgoing            bool
	HideHealthChecks       bool
	MaxEntriesDBSizeBytes  int64
	SleepIntervalSec       uint16
}

var mizuTapOptions = &MizuTapOptions{}
var direction string
var humanMaxEntriesDBSize string

const maxEntriesDBSizeFlagName = "max-entries-db-size"

var tapCmd = &cobra.Command{
	Use:   "tap [POD REGEX]",
	Short: "Record ingoing traffic of a kubernetes pod",
	Long: `Record the ingoing traffic of a kubernetes pod.
Supported protocols are HTTP and gRPC.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("POD REGEX argument is required")
		} else if len(args) > 1 {
			return errors.New("unexpected number of arguments")
		}

		regex, err := regexp.Compile(args[0])
		if err != nil {
			return errors.New(fmt.Sprintf("%s is not a valid regex %s", args[0], err))
		}

		mizuTapOptions.MaxEntriesDBSizeBytes = 200 * 1000 * 1000
		mizuTapOptions.MaxEntriesDBSizeBytes, err = units.HumanReadableToBytes(humanMaxEntriesDBSize)
		if err != nil {
			return errors.New(fmt.Sprintf("Could not parse --max-entries-db-size value %s", humanMaxEntriesDBSize))
		} else if cmd.Flags().Changed(maxEntriesDBSizeFlagName) {
			// We're parsing human readable file sizes here so its best to be unambiguous
			fmt.Printf("Setting max entries db size to %s\n", units.BytesToHumanReadable(mizuTapOptions.MaxEntriesDBSizeBytes))
		}

		directionLowerCase := strings.ToLower(direction)
		if directionLowerCase == "any" {
			mizuTapOptions.TapOutgoing = true
		} else if directionLowerCase == "in" {
			mizuTapOptions.TapOutgoing = false
		} else {
			return errors.New(fmt.Sprintf("%s is not a valid value for flag --direction. Acceptable values are in/any.", direction))
		}

		RunMizuTap(regex, mizuTapOptions)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(tapCmd)

	tapCmd.Flags().Uint16VarP(&mizuTapOptions.GuiPort, "gui-port", "p", 8899, "Provide a custom port for the web interface webserver")
	tapCmd.Flags().StringVarP(&mizuTapOptions.Namespace, "namespace", "n", "", "Namespace selector")
	tapCmd.Flags().BoolVar(&mizuTapOptions.Analyze, "analyze", false, "Uploads traffic to UP9 for further analysis (Beta)")
	tapCmd.Flags().StringVar(&mizuTapOptions.AnalyzeDestination, "dest", "up9.app", "Destination environment")
	tapCmd.Flags().Uint16VarP(&mizuTapOptions.SleepIntervalSec, "upload-interval", "", 10, "Interval in seconds for uploading data to UP9")
	tapCmd.Flags().BoolVarP(&mizuTapOptions.AllNamespaces, "all-namespaces", "A", false, "Tap all namespaces")
	tapCmd.Flags().StringVarP(&mizuTapOptions.KubeConfigPath, "kube-config", "k", "", "Path to kube-config file")
	tapCmd.Flags().StringVarP(&mizuTapOptions.MizuImage, "mizu-image", "", fmt.Sprintf("gcr.io/up9-docker-hub/mizu/%s:latest", mizu.Branch), "Custom image for mizu collector")
	tapCmd.Flags().StringArrayVarP(&mizuTapOptions.PlainTextFilterRegexes, "regex-masking", "r", nil, "List of regex expressions that are used to filter matching values from text/plain http bodies")
	tapCmd.Flags().StringVarP(&direction, "direction", "", "in", "Record traffic that goes in this direction (relative to the tapped pod): in/any")
	tapCmd.Flags().BoolVar(&mizuTapOptions.HideHealthChecks, "hide-healthchecks", false, "hides requests with kube-probe or prometheus user-agent headers")
	tapCmd.Flags().StringVarP(&humanMaxEntriesDBSize, maxEntriesDBSizeFlagName, "", "200MB", "override the default max entries db size of 200mb")
}
