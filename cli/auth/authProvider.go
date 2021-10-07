package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
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

func IsTokenExpired(tokenString string) (bool, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return true, fmt.Errorf("failed to parse token, err: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return true, fmt.Errorf("can't convert token's claims to standard claims")
	}

	expiry := time.Unix(int64(claims["exp"].(float64)), 0)

	return time.Now().After(expiry), nil
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

func loginCallbackHandler(tokenChannel chan *oauth2.Token, errorChannel chan error, config *oauth2.Config, envName string, state uuid.UUID) http.Handler {
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

		token, err := config.Exchange(context.Background(), code)
		if err != nil {
			errorMsg := fmt.Sprintf("failed to create token, err: %v", err)
			http.Error(writer, errorMsg, http.StatusInternalServerError)
			errorChannel <- fmt.Errorf(errorMsg)
			return
		}

		tokenChannel <- token

		http.Redirect(writer, request, fmt.Sprintf("https://%s/CliLogin", envName), http.StatusFound)
	})
}
