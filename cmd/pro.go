package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/creasty/defaults"
	"github.com/gin-gonic/gin"
	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/config/configStructs"
	"github.com/kubeshark/kubeshark/internal/connect"
	"github.com/kubeshark/kubeshark/kubernetes"
	"github.com/kubeshark/kubeshark/utils"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var proCmd = &cobra.Command{
	Use:   "pro",
	Short: "Acquire a Pro license",
	RunE: func(cmd *cobra.Command, args []string) error {
		acquireLicense()
		return nil
	},
}

const (
	PRO_URL  = "https://master.d2hqi2yb1n0lz7.amplifyapp.com/"
	PRO_PORT = 5252
)

func init() {
	rootCmd.AddCommand(proCmd)

	defaultTapConfig := configStructs.TapConfig{}
	if err := defaults.Set(&defaultTapConfig); err != nil {
		log.Debug().Err(err).Send()
	}

	proCmd.Flags().Uint16(configStructs.ProxyHubPortLabel, defaultTapConfig.Proxy.Hub.SrcPort, "Provide a custom port for the Hub.")
	proCmd.Flags().String(configStructs.ProxyHostLabel, defaultTapConfig.Proxy.Host, "Provide a custom host for the Hub.")
}

func acquireLicense() {
	if config.Config.Scripting.Source == "" {
		log.Error().Msg("`scripting.source` field is empty.")
		return
	}

	hubUrl := kubernetes.GetLocalhostOnPort(config.Config.Tap.Proxy.Hub.SrcPort)
	response, err := http.Get(fmt.Sprintf("%s/echo", hubUrl))
	if err != nil || response.StatusCode != 200 {
		log.Info().Msg(fmt.Sprintf(utils.Yellow, "Couldn't connect to Hub. Establishing proxy..."))
		runProxy(false)
	}

	connector = connect.NewConnector(kubernetes.GetLocalhostOnPort(config.Config.Tap.Proxy.Hub.SrcPort), connect.DefaultRetries, connect.DefaultTimeout)

	log.Info().Str("url", PRO_URL).Msg("Opening in the browser:")
	utils.OpenBrowser(PRO_URL)

	runLicenseRecieverServer()
}

func runLicenseRecieverServer() {
	gin.SetMode(gin.ReleaseMode)
	ginApp := gin.New()
	ginApp.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, x-session-token")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Disposition")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	ginApp.POST("/", func(c *gin.Context) {
		data, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			panic(err)
		}

		licenseKey := string(data)

		log.Info().Str("key", licenseKey).Msg("Received license:")

		config.Config.License = licenseKey
		err = config.WriteConfig(&config.Config)
		if err != nil {
			panic(err)
		}

		connector.PostLicense(config.Config.License)

		log.Info().Msg("Updated the license.")

		os.Exit(0)
	})

	if err := ginApp.Run(fmt.Sprintf(":%d", PRO_PORT)); err != nil {
		panic(err)
	}
}
