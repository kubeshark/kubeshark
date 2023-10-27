package kubernetes

import (
	"context"

	"github.com/kubeshark/kubeshark/config"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	SUFFIX_SECRET                = "secret"
	SUFFIX_CONFIG_MAP            = "config-map"
	SECRET_LICENSE               = "LICENSE"
	CONFIG_POD_REGEX             = "POD_REGEX"
	CONFIG_NAMESPACES            = "NAMESPACES"
	CONFIG_SCRIPTING_ENV         = "SCRIPTING_ENV"
	CONFIG_AUTH_ENABLED          = "AUTH_ENABLED"
	CONFIG_AUTH_APPROVED_EMAILS  = "AUTH_APPROVED_EMAILS"
	CONFIG_AUTH_APPROVED_DOMAINS = "AUTH_APPROVED_DOMAINS"
	CONFIG_AUTH_APPROVED_TENANTS = "AUTH_APPROVED_TENANTS"
)

func SetSecret(provider *Provider, key string, value string) (updated bool, err error) {
	var secret *v1.Secret
	secret, err = provider.clientSet.CoreV1().Secrets(config.Config.Tap.Release.Namespace).Get(context.TODO(), SELF_RESOURCES_PREFIX+SUFFIX_SECRET, metav1.GetOptions{})
	if err != nil {
		return
	}

	if secret.StringData[key] != value {
		updated = true
	}
	secret.Data[key] = []byte(value)

	_, err = provider.clientSet.CoreV1().Secrets(config.Config.Tap.Release.Namespace).Update(context.TODO(), secret, metav1.UpdateOptions{})
	if err == nil {
		if updated {
			log.Info().Str("secret", key).Str("value", value).Msg("Updated:")
		}
	} else {
		log.Error().Str("secret", key).Err(err).Send()
	}
	return
}

func SetConfig(provider *Provider, key string, value string) (updated bool, err error) {
	var configMap *v1.ConfigMap
	configMap, err = provider.clientSet.CoreV1().ConfigMaps(config.Config.Tap.Release.Namespace).Get(context.TODO(), SELF_RESOURCES_PREFIX+SUFFIX_CONFIG_MAP, metav1.GetOptions{})
	if err != nil {
		return
	}

	if configMap.Data[key] != value {
		updated = true
	}
	configMap.Data[key] = value

	_, err = provider.clientSet.CoreV1().ConfigMaps(config.Config.Tap.Release.Namespace).Update(context.TODO(), configMap, metav1.UpdateOptions{})
	if err == nil {
		if updated {
			log.Info().Str("config", key).Str("value", value).Msg("Updated:")
		}
	} else {
		log.Error().Str("config", key).Err(err).Send()
	}
	return
}
