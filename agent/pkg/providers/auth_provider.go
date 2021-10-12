package providers

import (
	"encoding/json"
	"fmt"
	"github.com/up9inc/mizu/shared"
	"os"
)

var authStatus *shared.AuthStatus

func GetAuthStatus() (*shared.AuthStatus, error) {
	if authStatus == nil {
		authStatus = &shared.AuthStatus{}

		authStatusJson := os.Getenv(shared.AuthStatusEnvVar)
		if authStatusJson == "" {
			return authStatus, nil
		}

		err := json.Unmarshal([]byte(authStatusJson), authStatus)
		if err != nil {
			authStatus = nil
			return nil, fmt.Errorf("failed to marshal auth status, err: %v", err)
		}
	}

	return authStatus, nil
}
