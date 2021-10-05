package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/up9inc/mizu/cli/logger"
	"github.com/up9inc/mizu/cli/uiUtils"
	"golang.org/x/oauth2"
	"net"
	"net/http"
	"time"
)

const loginTimeoutInMin = 2

// Ports are configured in keycloak "cli" client as valid redirect URIs. A change here must be reflected there as well.
var listenPorts = []int{3141, 4001, 5002, 6003, 7004, 8005, 9006, 10007}

func LoginInteractively(envName string) (*oauth2.Token, error) {
	tokenChannel := make(chan *oauth2.Token)
	errorChannel := make(chan error)

	server := http.Server{}
	go startLoginServer(tokenChannel, errorChannel, envName, &server)

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

func startLoginServer(tokenChannel chan *oauth2.Token, errorChannel chan error, envName string, server *http.Server) {
	for _, port := range listenPorts {
		var config = &oauth2.Config{
			ClientID:    "cli",
			RedirectURL: fmt.Sprintf("http://localhost:%v/callback", port),
			Endpoint: oauth2.Endpoint{
				AuthURL:  fmt.Sprintf("https://auth.%s/auth/realms/testr/protocol/openid-connect/auth", envName),
				TokenURL: fmt.Sprintf("https://auth.%s/auth/realms/testr/protocol/openid-connect/token", envName),
			},
		}

		state := uuid.New()

		mux := http.NewServeMux()
		server.Handler = mux
		mux.Handle("/callback", loginCallbackHandler(tokenChannel, errorChannel, config, envName, state))

		listener, listenErr := net.Listen("tcp", fmt.Sprintf("%s:%d", "127.0.0.1", port))
		if listenErr != nil {
			logger.Log.Debugf("failed to start listening on port %v, err: %v", port, listenErr)
			continue
		}

		authorizationUrl := config.AuthCodeURL(state.String())
		uiUtils.OpenBrowser(authorizationUrl)

		serveErr := server.Serve(listener)
		if serveErr == http.ErrServerClosed {
			logger.Log.Debugf("Received server shutdown, server on port %v is closed", port)
		} else if serveErr != nil {
			logger.Log.Debugf("failed to start serving on port %v, err: %v", port, serveErr)
			continue
		}

		return
	}

	errorChannel <- fmt.Errorf("failed to start serving on all listen ports")
}

func loginCallbackHandler(tokenChannel chan *oauth2.Token, errorChannel chan error, config *oauth2.Config, envName string, state uuid.UUID) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if err := request.ParseForm(); err != nil {
			errorChannel <- fmt.Errorf("failed to parse form, err: %v", err)
			http.Error(writer, fmt.Sprintf("failed to parse form, err: %v", err), http.StatusBadRequest)
			return
		}

		requestState := request.Form.Get("state")
		if requestState != state.String() {
			errorChannel <- fmt.Errorf("state invalid, requestState: %v, authState:%v", requestState, state.String())
			http.Error(writer, fmt.Sprintf("state invalid, requestState: %v, authState:%v", requestState, state.String()), http.StatusBadRequest)
			return
		}

		code := request.Form.Get("code")
		if code == "" {
			errorChannel <- fmt.Errorf("code not found")
			http.Error(writer, "code not found", http.StatusBadRequest)
			return
		}

		token, err := config.Exchange(context.Background(), code)
		if err != nil {
			errorChannel <- fmt.Errorf("failed to create token, err: %v", err)
			http.Error(writer, fmt.Sprintf("failed to create token, err: %v", err), http.StatusInternalServerError)
			return
		}

		tokenChannel <- token

		http.Redirect(writer, request, fmt.Sprintf("https://%s/CliLogin", envName), http.StatusFound)
	})
}
