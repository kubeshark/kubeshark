package check

import (
	"context"
	"fmt"

	"github.com/kubeshark/kubeshark/kubernetes"
	"github.com/rs/zerolog/log"
	rbac "k8s.io/api/rbac/v1"
)

func KubernetesPermissions(ctx context.Context, kubernetesProvider *kubernetes.Provider) bool {
	log.Info().Str("procedure", "kubernetes-permissions").Msg("Checking:")
	return checkRulesPermissions(ctx, kubernetesProvider, kubernetesProvider.BuildClusterRole().Rules, "")
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
