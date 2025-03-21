package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/creasty/defaults"
	"github.com/kubeshark/kubeshark/config/configStructs"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// pcapDumpCmd represents the consolidated pcapdump command
var pcapDumpCmd = &cobra.Command{
	Use:   "pcapdump",
	Short: "Store all captured traffic (including decrypted TLS) in a PCAP file.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Retrieve the kubeconfig path from the flag
		kubeconfig, _ := cmd.Flags().GetString(configStructs.PcapKubeconfig)

		// If kubeconfig is not provided, use the default location
		if kubeconfig == "" {
			if home := homedir.HomeDir(); home != "" {
				kubeconfig = filepath.Join(home, ".kube", "config")
			} else {
				return errors.New("kubeconfig flag not provided and no home directory available for default config location")
			}
		}

		debugEnabled, _ := cmd.Flags().GetBool("debug")
		if debugEnabled {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
			log.Debug().Msg("Debug logging enabled")
		} else {
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
		}

		// Use the current context in kubeconfig
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return fmt.Errorf("Error building kubeconfig: %w", err)
		}

		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			return fmt.Errorf("Error creating Kubernetes client: %w", err)
		}

		// Parse the `--time` flag
		timeIntervalStr, _ := cmd.Flags().GetString("time")
		var cutoffTime *time.Time // Use a pointer to distinguish between provided and not provided
		if timeIntervalStr != "" {
			duration, err := time.ParseDuration(timeIntervalStr)
			if err != nil {
				return fmt.Errorf("Invalid format %w", err)
			}
			tempCutoffTime := time.Now().Add(-duration)
			cutoffTime = &tempCutoffTime
		}

		// Test the dest dir if provided
		destDir, _ := cmd.Flags().GetString(configStructs.PcapDest)
		if destDir != "" {
			info, err := os.Stat(destDir)
			if os.IsNotExist(err) {
				return fmt.Errorf("Directory does not exist: %s", destDir)
			}
			if err != nil {
				return fmt.Errorf("Error checking dest directory: %w", err)
			}
			if !info.IsDir() {
				return fmt.Errorf("Dest path is not a directory: %s", destDir)
			}
			tempFile, err := os.CreateTemp(destDir, "write-test-*")
			if err != nil {
				return fmt.Errorf("Directory %s is not writable", destDir)
			}
			_ = os.Remove(tempFile.Name())
		}

		log.Info().Msg("Copying PCAP files")
		err = copyPcapFiles(clientset, config, destDir, cutoffTime)
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(pcapDumpCmd)

	defaultPcapDumpConfig := configStructs.PcapDumpConfig{}
	if err := defaults.Set(&defaultPcapDumpConfig); err != nil {
		log.Debug().Err(err).Send()
	}

	pcapDumpCmd.Flags().String(configStructs.PcapTime, "", "Time interval (e.g., 10m, 1h) in the past for which the pcaps are copied")
	pcapDumpCmd.Flags().String(configStructs.PcapDest, "", "Local destination path for copied PCAP files (can not be used together with --enabled)")
	pcapDumpCmd.Flags().String(configStructs.PcapKubeconfig, "", "Path for kubeconfig (if not provided the default location will be checked)")
	pcapDumpCmd.Flags().Bool("debug", false, "Enable debug logging")
}
