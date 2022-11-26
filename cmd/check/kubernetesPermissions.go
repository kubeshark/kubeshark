package check

import (
	"context"
	"embed"
	"fmt"
	"log"
	"strings"

	"github.com/kubeshark/kubeshark/bucket"
	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/kubernetes"
	"github.com/kubeshark/kubeshark/utils"
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

func TapKubernetesPermissions(ctx context.Context, embedFS embed.FS, kubernetesProvider *kubernetes.Provider) bool {
	log.Printf("\nkubernetes-permissions\n--------------------")

	var filePath string
	if config.Config.IsNsRestrictedMode() {
		filePath = "permissionFiles/permissions-ns-tap.yaml"
	} else {
		filePath = "permissionFiles/permissions-all-namespaces-tap.yaml"
	}

	data, err := embedFS.ReadFile(filePath)
	if err != nil {
		log.Printf("%v error while checking kubernetes permissions, err: %v", fmt.Sprintf(utils.Red, "✗"), err)
		return false
	}

	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(data, nil, nil)
	if err != nil {
		log.Printf("%v error while checking kubernetes permissions, err: %v", fmt.Sprintf(utils.Red, "✗"), err)
		return false
	}

	switch resource := obj.(type) {
	case *rbac.Role:
		return checkRulesPermissions(ctx, kubernetesProvider, resource.Rules, config.Config.KubesharkResourcesNamespace)
	case *rbac.ClusterRole:
		return checkRulesPermissions(ctx, kubernetesProvider, resource.Rules, "")
	}

	log.Printf("%v error while checking kubernetes permissions, err: resource of type 'Role' or 'ClusterRole' not found in permission files", fmt.Sprintf(utils.Red, "✗"))
	return false
}

func InstallKubernetesPermissions(ctx context.Context, kubernetesProvider *kubernetes.Provider) bool {
	log.Printf("\nkubernetes-permissions\n--------------------")

	bucketProvider := bucket.NewProvider(config.Config.Install.TemplateUrl, bucket.DefaultTimeout)
	installTemplate, err := bucketProvider.GetInstallTemplate(config.Config.Install.TemplateName)
	if err != nil {
		log.Printf("%v error while checking kubernetes permissions, err: %v", fmt.Sprintf(utils.Red, "✗"), err)
		return false
	}

	resourcesTemplate := strings.Split(installTemplate, "---")[1:]

	permissionsExist := true

	decode := scheme.Codecs.UniversalDeserializer().Decode
	for _, resourceTemplate := range resourcesTemplate {
		obj, _, err := decode([]byte(resourceTemplate), nil, nil)
		if err != nil {
			log.Printf("%v error while checking kubernetes permissions, err: %v", fmt.Sprintf(utils.Red, "✗"), err)
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
		log.Printf("%v error checking permission for %v %v %v, err: %v", fmt.Sprintf(utils.Red, "✗"), verb, resource, groupAndNamespace, err)
		return false
	} else if !exist {
		log.Printf("%v can't %v %v %v", fmt.Sprintf(utils.Red, "✗"), verb, resource, groupAndNamespace)
		return false
	}

	log.Printf("%v can %v %v %v", fmt.Sprintf(utils.Green, "√"), verb, resource, groupAndNamespace)
	return true
}
