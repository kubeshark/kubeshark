package bucket

import (
	"fmt"
	"github.com/up9inc/mizu/cli/utils"
	"io/ioutil"
	"net/http"
	"time"
)

type Provider struct {
	url     string
	client  *http.Client
}

const DefaultTimeout = 2 * time.Second

func NewProvider(url string, timeout time.Duration) *Provider {
	return &Provider{
		url:     url,
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

	installTemplate, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return string(installTemplate), nil
}
