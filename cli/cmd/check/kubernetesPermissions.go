package check

import (
	"context"
	"embed"
	"fmt"
	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/cli/uiUtils"
	"github.com/up9inc/mizu/shared/kubernetes"
	"github.com/up9inc/mizu/shared/logger"
	rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
)

func TapKubernetesPermissions(ctx context.Context, embedFS embed.FS, kubernetesProvider *kubernetes.Provider) bool {
	logger.Log.Infof("\nkubernetes-permissions\n--------------------")

	var filePath string
	if config.Config.IsNsRestrictedMode() {
		filePath = "permissionFiles/permissions-ns-tap.yaml"
	} else {
		filePath = "permissionFiles/permissions-all-namespaces-tap.yaml"
	}

	data, err := embedFS.ReadFile(filePath)
	if err != nil {
		logger.Log.Errorf("%v error while checking kubernetes permissions, err: %v", fmt.Sprintf(uiUtils.Red, "✗"), err)
		return false
	}

	obj, err := getDecodedObject(data)
	if err != nil {
		logger.Log.Errorf("%v error while checking kubernetes permissions, err: %v", fmt.Sprintf(uiUtils.Red, "✗"), err)
		return false
	}

	var rules []rbac.PolicyRule
	if config.Config.IsNsRestrictedMode() {
		rules = obj.(*rbac.Role).Rules
	} else {
		rules = obj.(*rbac.ClusterRole).Rules
	}

	return checkPermissions(ctx, kubernetesProvider, rules)
}

func getDecodedObject(data []byte) (runtime.Object, error) {
	decode := scheme.Codecs.UniversalDeserializer().Decode

	obj, _, err := decode(data, nil, nil)
	if err != nil {
		return nil, err
	}

	return obj, nil
}

func checkPermissions(ctx context.Context, kubernetesProvider *kubernetes.Provider, rules []rbac.PolicyRule) bool {
	permissionsExist := true

	for _, rule := range rules {
		for _, group := range rule.APIGroups {
			for _, resource := range rule.Resources {
				for _, verb := range rule.Verbs {
					exist, err := kubernetesProvider.CanI(ctx, config.Config.MizuResourcesNamespace, resource, verb, group)
					permissionsExist = checkPermissionExist(group, resource, verb, exist, err) && permissionsExist
				}
			}
		}
	}

	return permissionsExist
}

func checkPermissionExist(group string, resource string, verb string, exist bool, err error) bool {
	if err != nil {
		logger.Log.Errorf("%v error checking permission for %v %v in group '%v', err: %v", fmt.Sprintf(uiUtils.Red, "✗"), verb, resource, group, err)
		return false
	} else if !exist {
		logger.Log.Errorf("%v can't %v %v in group '%v'", fmt.Sprintf(uiUtils.Red, "✗"), verb, resource, group)
		return false
	}

	logger.Log.Infof("%v can %v %v in group '%v'", fmt.Sprintf(uiUtils.Green, "√"), verb, resource, group)
	return true
}
