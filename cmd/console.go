package cmd

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/creasty/defaults"
	"github.com/gorilla/websocket"
	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/config/configStructs"
	"github.com/kubeshark/kubeshark/kubernetes"
	"github.com/kubeshark/kubeshark/utils"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var consoleCmd = &cobra.Command{
	Use:   "console",
	Short: "Stream the scripting console logs into shell",
	RunE: func(cmd *cobra.Command, args []string) error {
		runConsole()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(consoleCmd)

	defaultTapConfig := configStructs.TapConfig{}
	if err := defaults.Set(&defaultTapConfig); err != nil {
		log.Debug().Err(err).Send()
	}

	consoleCmd.Flags().Uint16(configStructs.ProxyFrontPortLabel, defaultTapConfig.Proxy.Front.Port, "Provide a custom port for the Kubeshark")
	consoleCmd.Flags().String(configStructs.ProxyHostLabel, defaultTapConfig.Proxy.Host, "Provide a custom host for the Kubeshark")
	consoleCmd.Flags().StringP(configStructs.ReleaseNamespaceLabel, "s", defaultTapConfig.Release.Namespace, "Release namespace of Kubeshark")
}

func runConsole() {
	hubUrl := kubernetes.GetHubUrl()
	response, err := http.Get(fmt.Sprintf("%s/echo", hubUrl))
	if err != nil || response.StatusCode != 200 {
		log.Info().Msg(fmt.Sprintf(utils.Yellow, "Couldn't connect to Hub. Establishing proxy..."))
		runProxy(false, true)
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	log.Info().Str("host", config.Config.Tap.Proxy.Host).Str("url", hubUrl).Msg("Connecting to:")
	u := url.URL{
		Scheme: "ws",
		Host:   fmt.Sprintf("%s:%d/api", config.Config.Tap.Proxy.Host, config.Config.Tap.Proxy.Front.Port),
		Path:   "/scripts/logs",
	}
	headers := http.Header{}
	headers.Set(utils.X_KUBESHARK_CAPTURE_HEADER_KEY, utils.X_KUBESHARK_CAPTURE_HEADER_IGNORE_VALUE)
	headers.Set("License-Key", config.Config.License)

	c, _, err := websocket.DefaultDialer.Dial(u.String(), headers)
	if err != nil {
		log.Error().Err(err).Send()
		return
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Error().Err(err).Send()
				return
			}

			msg := string(message)
			if strings.Contains(msg, ":ERROR]") {
				msg = fmt.Sprintf(utils.Red, msg)
				fmt.Fprintln(os.Stderr, msg)
			} else {
				fmt.Fprintln(os.Stdout, msg)
			}
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-interrupt:
			log.Warn().Msg(fmt.Sprintf(utils.Yellow, "Received interrupt, exiting..."))

			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Error().Err(err).Send()
				return
			}

			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
