package auth

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/cli/config/configStructs"
	"github.com/up9inc/mizu/cli/uiUtils"
	"github.com/up9inc/mizu/shared/logger"
	"golang.org/x/oauth2"
)

const loginTimeoutInMin = 2

// Ports are configured in keycloak "cli" client as valid redirect URIs. A change here must be reflected there as well.
var listenPorts = []int{3141, 4001, 5002, 6003, 7004, 8005, 9006, 10007}

func Login() error {
	token, loginErr := loginInteractively()
	if loginErr != nil {
		return fmt.Errorf("failed login interactively, err: %v", loginErr)
	}

	authConfig := configStructs.AuthConfig{
		EnvName: config.Config.Auth.EnvName,
		Token:   token.AccessToken,
	}

	configFile, defaultConfigErr := config.GetConfigWithDefaults()
	if defaultConfigErr != nil {
		return fmt.Errorf("failed getting config with defaults, err: %v", defaultConfigErr)
	}

	if err := config.LoadConfigFile(config.Config.ConfigFilePath, configFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed getting config file, err: %v", err)
	}

	configFile.Auth = authConfig

	if err := config.WriteConfig(configFile); err != nil {
		return fmt.Errorf("failed writing config with auth, err: %v", err)
	}

	config.Config.Auth = authConfig

	logger.Log.Infof("Login successfully, token stored in config path: %s", fmt.Sprintf(uiUtils.Purple, config.Config.ConfigFilePath))
	return nil
}

func loginInteractively() (*oauth2.Token, error) {
	tokenChannel := make(chan *oauth2.Token)
	errorChannel := make(chan error)

	server := http.Server{}
	go startLoginServer(tokenChannel, errorChannel, &server)

	defer func() {
		if err := server.Shutdown(context.Background()); err != nil {
			logger.Log.Debugf("Error shutting down server, err: %v", err)
		}
	}()

	select {
	case <-time.After(loginTimeoutInMin * time.Minute):
		return nil, errors.New("auth timed out")
	case err := <-errorChannel:
		return nil, err
	case token := <-tokenChannel:
		return token, nil
	}
}

func startLoginServer(tokenChannel chan *oauth2.Token, errorChannel chan error, server *http.Server) {
	for _, port := range listenPorts {
		var authConfig = &oauth2.Config{
			ClientID:    "cli",
			RedirectURL: fmt.Sprintf("http://localhost:%v/callback", port),
			Endpoint: oauth2.Endpoint{
				AuthURL:  fmt.Sprintf("https://auth.%s/auth/realms/testr/protocol/openid-connect/auth", config.Config.Auth.EnvName),
				TokenURL: fmt.Sprintf("https://auth.%s/auth/realms/testr/protocol/openid-connect/token", config.Config.Auth.EnvName),
			},
		}

		state := uuid.New()

		mux := http.NewServeMux()
		server.Handler = mux
		mux.Handle("/callback", loginCallbackHandler(tokenChannel, errorChannel, authConfig, state))

		listener, listenErr := net.Listen("tcp", fmt.Sprintf("%s:%d", "127.0.0.1", port))
		if listenErr != nil {
			logger.Log.Debugf("failed to start listening on port %v, err: %v", port, listenErr)
			continue
		}

		authorizationUrl := authConfig.AuthCodeURL(state.String())
		uiUtils.OpenBrowser(authorizationUrl)

		serveErr := server.Serve(listener)
		if serveErr == http.ErrServerClosed {
			logger.Log.Debugf("received server shutdown, server on port %v is closed", port)
			return
		} else if serveErr != nil {
			logger.Log.Debugf("failed to start serving on port %v, err: %v", port, serveErr)
			continue
		}

		logger.Log.Debugf("didn't receive server closed on port %v", port)
		return
	}

	errorChannel <- fmt.Errorf("failed to start serving on all listen ports, ports: %v", listenPorts)
}

func loginCallbackHandler(tokenChannel chan *oauth2.Token, errorChannel chan error, authConfig *oauth2.Config, state uuid.UUID) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if err := request.ParseForm(); err != nil {
			errorMsg := fmt.Sprintf("failed to parse form, err: %v", err)
			http.Error(writer, errorMsg, http.StatusBadRequest)
			errorChannel <- fmt.Errorf(errorMsg)
			return
		}

		requestState := request.Form.Get("state")
		if requestState != state.String() {
			errorMsg := fmt.Sprintf("state invalid, requestState: %v, authState:%v", requestState, state.String())
			http.Error(writer, errorMsg, http.StatusBadRequest)
			errorChannel <- fmt.Errorf(errorMsg)
			return
		}

		code := request.Form.Get("code")
		if code == "" {
			errorMsg := "code not found"
			http.Error(writer, errorMsg, http.StatusBadRequest)
			errorChannel <- fmt.Errorf(errorMsg)
			return
		}

		token, err := authConfig.Exchange(context.Background(), code)
		if err != nil {
			errorMsg := fmt.Sprintf("failed to create token, err: %v", err)
			http.Error(writer, errorMsg, http.StatusInternalServerError)
			errorChannel <- fmt.Errorf(errorMsg)
			return
		}

		tokenChannel <- token

		http.Redirect(writer, request, fmt.Sprintf("https://%s/CliLogin", config.Config.Auth.EnvName), http.StatusFound)
	})
}
