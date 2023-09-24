package kubernetes

import (
	"context"

	"github.com/kubeshark/kubeshark/config"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	SUFFIX_SECRET = "secret"
)

func SetSecret(provider *Provider, key string, value string) (err error) {
	var secret *v1.Secret
	secret, err = provider.clientSet.CoreV1().Secrets(config.Config.Tap.Release.Namespace).Get(context.TODO(), SelfResourcesPrefix+SUFFIX_SECRET, metav1.GetOptions{})
	if err != nil {
		return
	}

	secret.StringData[key] = value

	_, err = provider.clientSet.CoreV1().Secrets(config.Config.Tap.Release.Namespace).Update(context.TODO(), secret, metav1.UpdateOptions{})
	return
}
