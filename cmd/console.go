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

func runConsoleWithoutProxy() {
	log.Info().Msg("Starting scripting console ...")
	time.Sleep(5 * time.Second)
	hubUrl := kubernetes.GetHubUrl()
	for {

		// Attempt to connect to the Hub every second
		response, err := http.Get(fmt.Sprintf("%s/echo", hubUrl))
		if err != nil || response.StatusCode != 200 {
			log.Info().Msg(fmt.Sprintf(utils.Yellow, "Couldn't connect to Hub."))
			time.Sleep(5 * time.Second)
			continue
		}

		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt)

		log.Info().Str("host", config.Config.Tap.Proxy.Host).Str("url", hubUrl).Msg("Connecting to:")
		u := url.URL{
			Scheme: "ws",
			Host:   fmt.Sprintf("%s:%d", config.Config.Tap.Proxy.Host, config.Config.Tap.Proxy.Front.Port),
			Path:   "/api/scripts/logs",
		}
		headers := http.Header{}
		headers.Set(utils.X_KUBESHARK_CAPTURE_HEADER_KEY, utils.X_KUBESHARK_CAPTURE_HEADER_IGNORE_VALUE)
		headers.Set("License-Key", config.Config.License)

		c, _, err := websocket.DefaultDialer.Dial(u.String(), headers)
		if err != nil {
			log.Error().Err(err).Msg("Websocket dial error, retrying in 5 seconds...")
			time.Sleep(5 * time.Second) // Delay before retrying
			continue
		}
		defer c.Close()

		done := make(chan struct{})

		go func() {
			defer close(done)
			for {
				_, message, err := c.ReadMessage()
				if err != nil {
					log.Error().Err(err).Msg("Error reading websocket message, reconnecting...")
					break // Break to reconnect
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

		select {
		case <-done:
			log.Warn().Msg(fmt.Sprintf(utils.Yellow, "Connection closed, reconnecting..."))
			time.Sleep(5 * time.Second) // Delay before reconnecting
			continue                    // Reconnect after error
		case <-interrupt:
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Error().Err(err).Send()
				continue
			}

			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

func runConsole() {
	go runConsoleWithoutProxy()

	// Create interrupt channel and setup signal handling once
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	done := make(chan struct{})

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-interrupt:
			// Handle interrupt and exit gracefully
			log.Warn().Msg(fmt.Sprintf(utils.Yellow, "Received interrupt, exiting..."))
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return

		case <-ticker.C:
			// Attempt to connect to the Hub every second
			hubUrl := kubernetes.GetHubUrl()
			response, err := http.Get(fmt.Sprintf("%s/echo", hubUrl))
			if err != nil || response.StatusCode != 200 {
				log.Info().Msg(fmt.Sprintf(utils.Yellow, "Couldn't connect to Hub. Establishing proxy..."))
				runProxy(false, true)
			}
		}
	}
}
