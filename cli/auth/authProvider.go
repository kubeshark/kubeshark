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

// Ports are configured in keycloak "cli" client as valid redirect URIs. A change here must be reflected there as well.
var listenPorts = []int{3141, 4001, 5002, 6003, 7004, 8005, 9006, 10007}

func LoginInteractively(envName string) (*oauth2.Token, error) {
	tokenChannel := make(chan *oauth2.Token)

	go func() {
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
			server := http.Server{Handler: mux}

			mux.HandleFunc("/callback", func(writer http.ResponseWriter, request *http.Request) {
				if err := request.ParseForm(); err != nil {
					logger.Log.Errorf("Failed to parse form, err: %v", err)
					http.Error(writer, "Failed to parse form", http.StatusBadRequest)
					return
				}

				requestState := request.Form.Get("state")
				if requestState != state.String() {
					logger.Log.Errorf("State invalid, requestState: %v, authState:", requestState, state.String())
					http.Error(writer, "State invalid", http.StatusBadRequest)
					return
				}

				code := request.Form.Get("code")
				if code == "" {
					logger.Log.Errorf("Code not found")
					http.Error(writer, "Code not found", http.StatusBadRequest)
					return
				}

				token, err := config.Exchange(context.Background(), code)
				if err != nil {
					logger.Log.Errorf("Failed to create token, err: %v", err)
					http.Error(writer, "Failed to create token", http.StatusInternalServerError)
					return
				}

				http.Redirect(writer, request, fmt.Sprintf("https://%s/CliLogin", envName), http.StatusFound)

				flusher, ok := writer.(http.Flusher)
				if !ok {
					logger.Log.Errorf("No flush support")
					http.Error(writer, "No flush support", http.StatusInternalServerError)
					return
				}

				flusher.Flush()

				tokenChannel <- token

				if err := server.Shutdown(context.Background()); err != nil {
					logger.Log.Warningf("Error shutting down server, err: %v", err)
				}
			})

			listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", "127.0.0.1", port))
			if err != nil {
				continue
			}

			authorizationUrl := config.AuthCodeURL(state.String())
			uiUtils.OpenBrowser(authorizationUrl)

			if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
				continue
			}

			return
		}

		tokenChannel <- nil
	}()

	select {
	case <-time.After(2 * time.Minute):
	case token := <-tokenChannel:
		if token == nil {
			return nil, errors.New("failed to start serving on all listen ports")
		}

		return token, nil
	}

	return nil, errors.New("auth timed out")
}
