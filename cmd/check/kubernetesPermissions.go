package check

import (
	"context"
	"embed"
	"fmt"

	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/kubernetes"
	"github.com/rs/zerolog/log"
	rbac "k8s.io/api/rbac/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

func KubernetesPermissions(ctx context.Context, embedFS embed.FS, kubernetesProvider *kubernetes.Provider) bool {
	log.Info().Str("procedure", "kubernetes-permissions").Msg("Checking:")

	var filePath string
	if config.Config.IsNsRestrictedMode() {
		filePath = "permissionFiles/permissions-ns-tap.yaml"
	} else {
		filePath = "permissionFiles/permissions-all-namespaces-tap.yaml"
	}

	data, err := embedFS.ReadFile(filePath)
	if err != nil {
		log.Error().Err(err).Msg("While checking Kubernetes permissions!")
		return false
	}

	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(data, nil, nil)
	if err != nil {
		log.Error().Err(err).Msg("While checking Kubernetes permissions!")
		return false
	}

	switch resource := obj.(type) {
	case *rbac.Role:
		return checkRulesPermissions(ctx, kubernetesProvider, resource.Rules, config.Config.ResourcesNamespace)
	case *rbac.ClusterRole:
		return checkRulesPermissions(ctx, kubernetesProvider, resource.Rules, "")
	}

	log.Error().Msg("While checking Kubernetes permissions! Resource of types 'Role' or 'ClusterRole' are not found in permission files.")
	return false
}

func checkRulesPermissions(ctx context.Context, kubernetesProvider *kubernetes.Provider, rules []rbac.PolicyRule, namespace string) bool {
	permissionsExist := true

	for _, rule := range rules {
		for _, group := range rule.APIGroups {
			for _, resource := range rule.Resources {
				for _, verb := range rule.Verbs {
					exist, err := kubernetesProvider.CanI(ctx, namespace, resource, verb, group)
					permissionsExist = checkPermissionExist(group, resource, verb, namespace, exist, err) && permissionsExist
				}
			}
		}
	}

	return permissionsExist
}

func checkPermissionExist(group string, resource string, verb string, namespace string, exist bool, err error) bool {
	var groupAndNamespace string
	if group != "" && namespace != "" {
		groupAndNamespace = fmt.Sprintf("in api group '%v' and namespace '%v'", group, namespace)
	} else if group != "" {
		groupAndNamespace = fmt.Sprintf("in api group '%v'", group)
	} else if namespace != "" {
		groupAndNamespace = fmt.Sprintf("in namespace '%v'", namespace)
	}

	if err != nil {
		log.Error().
			Str("verb", verb).
			Str("resource", resource).
			Str("group-and-namespace", groupAndNamespace).
			Err(err).
			Msg("While checking Kubernetes permissions!")
		return false
	} else if !exist {
		log.Error().Msg(fmt.Sprintf("Can't %v %v %v", verb, resource, groupAndNamespace))
		return false
	}

	log.Info().Msg(fmt.Sprintf("Can %v %v %v", verb, resource, groupAndNamespace))
	return true
}
