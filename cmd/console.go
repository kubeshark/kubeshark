package cmd

import (
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/creasty/defaults"
	"github.com/gorilla/websocket"
	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/config/configStructs"
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

	consoleCmd.Flags().Uint16(configStructs.ProxyHubPortLabel, defaultTapConfig.Proxy.Hub.SrcPort, "Provide a custom port for the Hub.")
	consoleCmd.Flags().String(configStructs.ProxyHostLabel, defaultTapConfig.Proxy.Host, "Provide a custom host for the Hub.")
}

func runConsole() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	log.Info().Str("host", config.Config.Tap.Proxy.Host).Uint16("port", config.Config.Tap.Proxy.Hub.SrcPort).Msg("Connecting to:")
	u := url.URL{
		Scheme: "ws",
		Host:   fmt.Sprintf("%s:%d", config.Config.Tap.Proxy.Host, config.Config.Tap.Proxy.Hub.SrcPort),
		Path:   "/scripts/logs",
	}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Error().Err(err).Send()
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
			}
			log.Info().Msg(msg)
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
