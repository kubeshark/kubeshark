package bucket

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kubeshark/kubeshark/utils"
)

type Provider struct {
	url    string
	client *http.Client
}

const DefaultTimeout = 2 * time.Second

func NewProvider(url string, timeout time.Duration) *Provider {
	return &Provider{
		url: url,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (provider *Provider) GetInstallTemplate(templateName string) (string, error) {
	url := fmt.Sprintf("%s/%v", provider.url, templateName)
	response, err := utils.Get(url, provider.client)
	if err != nil {
		return "", err
	}

	defer response.Body.Close()

	installTemplate, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return string(installTemplate), nil
}
