package kubernetes

import (
	"github.com/kubeshark/kubeshark/config"
)

func buildWithDefaultLabels(labels map[string]string, provider *Provider) map[string]string {
	labels[LabelManagedBy] = provider.managedBy
	labels[LabelCreatedBy] = provider.createdBy

	for k, v := range config.Config.Tap.ResourceLabels {
		labels[k] = v
	}

	return labels
}
