package up9

import (
	"fmt"
	"net/http"
	"net/url"
)

func IsTokenValid(tokenString string, envName string) bool {
	whoAmIUrl, _ := url.Parse(fmt.Sprintf("https://trcc.%s/admin/whoami", envName))

	req := &http.Request{
		Method: http.MethodGet,
		URL:    whoAmIUrl,
		Header: map[string][]string{
			"Authorization": {fmt.Sprintf("bearer %s", tokenString)},
		},
	}

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return false
	}

	return true
}
