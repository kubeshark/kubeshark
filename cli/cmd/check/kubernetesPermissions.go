package check

import (
	"context"
	"embed"
	"fmt"
	"strings"

	"github.com/up9inc/mizu/cli/bucket"
	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/cli/uiUtils"
	"github.com/up9inc/mizu/logger"
	"github.com/up9inc/mizu/shared/kubernetes"
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(data, nil, nil)
	if err != nil {
		logger.Log.Errorf("%v error while checking kubernetes permissions, err: %v", fmt.Sprintf(uiUtils.Red, "✗"), err)
		return false
	}

	switch resource := obj.(type) {
	case *rbac.Role:
		return checkRulesPermissions(ctx, kubernetesProvider, resource.Rules, config.Config.MizuResourcesNamespace)
	case *rbac.ClusterRole:
		return checkRulesPermissions(ctx, kubernetesProvider, resource.Rules, "")
	}

	logger.Log.Errorf("%v error while checking kubernetes permissions, err: resource of type 'Role' or 'ClusterRole' not found in permission files", fmt.Sprintf(uiUtils.Red, "✗"))
	return false
}

func InstallKubernetesPermissions(ctx context.Context, kubernetesProvider *kubernetes.Provider) bool {
	logger.Log.Infof("\nkubernetes-permissions\n--------------------")

	bucketProvider := bucket.NewProvider(config.Config.Install.TemplateUrl, bucket.DefaultTimeout)
	installTemplate, err := bucketProvider.GetInstallTemplate(config.Config.Install.TemplateName)
	if err != nil {
		logger.Log.Errorf("%v error while checking kubernetes permissions, err: %v", fmt.Sprintf(uiUtils.Red, "✗"), err)
		return false
	}

	resourcesTemplate := strings.Split(installTemplate, "---")[1:]

	permissionsExist := true

	decode := scheme.Codecs.UniversalDeserializer().Decode
	for _, resourceTemplate := range resourcesTemplate {
		obj, _, err := decode([]byte(resourceTemplate), nil, nil)
		if err != nil {
			logger.Log.Errorf("%v error while checking kubernetes permissions, err: %v", fmt.Sprintf(uiUtils.Red, "✗"), err)
			return false
		}

		groupVersionKind := obj.GetObjectKind().GroupVersionKind()
		resource := fmt.Sprintf("%vs", strings.ToLower(groupVersionKind.Kind))
		permissionsExist = checkCreatePermission(ctx, kubernetesProvider, resource, groupVersionKind.Group, obj.(metav1.Object).GetNamespace()) && permissionsExist

		switch resourceObj := obj.(type) {
		case *rbac.Role:
			permissionsExist = checkRulesPermissions(ctx, kubernetesProvider, resourceObj.Rules, resourceObj.Namespace) && permissionsExist
		case *rbac.ClusterRole:
			permissionsExist = checkRulesPermissions(ctx, kubernetesProvider, resourceObj.Rules, "") && permissionsExist
		}
	}

	return permissionsExist
}

func checkCreatePermission(ctx context.Context, kubernetesProvider *kubernetes.Provider, resource string, group string, namespace string) bool {
	exist, err := kubernetesProvider.CanI(ctx, namespace, resource, "create", group)
	return checkPermissionExist(group, resource, "create", namespace, exist, err)
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
		logger.Log.Errorf("%v error checking permission for %v %v %v, err: %v", fmt.Sprintf(uiUtils.Red, "✗"), verb, resource, groupAndNamespace, err)
		return false
	} else if !exist {
		logger.Log.Errorf("%v can't %v %v %v", fmt.Sprintf(uiUtils.Red, "✗"), verb, resource, groupAndNamespace)
		return false
	}

	logger.Log.Infof("%v can %v %v %v", fmt.Sprintf(uiUtils.Green, "√"), verb, resource, groupAndNamespace)
	return true
}
