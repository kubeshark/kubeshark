package kubernetes

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"regexp"

	"github.com/op/go-logging"
	"github.com/up9inc/mizu/logger"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/semver"
	"github.com/up9inc/mizu/tap/api"
	auth "k8s.io/api/authorization/v1"
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/apimachinery/pkg/watch"
	applyconfapp "k8s.io/client-go/applyconfigurations/apps/v1"
	applyconfcore "k8s.io/client-go/applyconfigurations/core/v1"
	applyconfmeta "k8s.io/client-go/applyconfigurations/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	watchtools "k8s.io/client-go/tools/watch"
)

type Provider struct {
	clientSet        *kubernetes.Clientset
	kubernetesConfig clientcmd.ClientConfig
	clientConfig     rest.Config
	managedBy        string
	createdBy        string
}

const (
	fieldManagerName = "mizu-manager"
	procfsVolumeName = "proc"
	procfsMountPath  = "/hostproc"
	sysfsVolumeName  = "sys"
	sysfsMountPath   = "/sys"
)

func NewProvider(kubeConfigPath string, contextName string) (*Provider, error) {
	kubernetesConfig := loadKubernetesConfiguration(kubeConfigPath, contextName)
	restClientConfig, err := kubernetesConfig.ClientConfig()
	if err != nil {
		if clientcmd.IsEmptyConfig(err) {
			return nil, fmt.Errorf("couldn't find the kube config file, or file is empty (%s)\n"+
				"you can set alternative kube config file path by adding the kube-config-path field to the mizu config file, err:  %w", kubeConfigPath, err)
		}
		if clientcmd.IsConfigurationInvalid(err) {
			return nil, fmt.Errorf("invalid kube config file (%s)\n"+
				"you can set alternative kube config file path by adding the kube-config-path field to the mizu config file, err:  %w", kubeConfigPath, err)
		}

		return nil, fmt.Errorf("error while using kube config (%s)\n"+
			"you can set alternative kube config file path by adding the kube-config-path field to the mizu config file, err:  %w", kubeConfigPath, err)
	}

	clientSet, err := getClientSet(restClientConfig)
	if err != nil {
		return nil, fmt.Errorf("error while using kube config (%s)\n"+
			"you can set alternative kube config file path by adding the kube-config-path field to the mizu config file, err:  %w", kubeConfigPath, err)
	}

	logger.Log.Debugf("K8s client config, host: %s, api path: %s, user agent: %s", restClientConfig.Host, restClientConfig.APIPath, restClientConfig.UserAgent)

	return &Provider{
		clientSet:        clientSet,
		kubernetesConfig: kubernetesConfig,
		clientConfig:     *restClientConfig,
		managedBy:        LabelValueMizu,
		createdBy:        LabelValueMizuCLI,
	}, nil
}

//NewProviderInCluster Used in another repo that calls this function
func NewProviderInCluster() (*Provider, error) {
	restClientConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	clientSet, err := getClientSet(restClientConfig)
	if err != nil {
		return nil, err
	}

	return &Provider{
		clientSet:        clientSet,
		kubernetesConfig: nil, // not relevant in cluster
		clientConfig:     *restClientConfig,
		managedBy:        LabelValueMizu,
		createdBy:        LabelValueMizuAgent,
	}, nil
}

func (provider *Provider) CurrentNamespace() (string, error) {
	if provider.kubernetesConfig == nil {
		return "", errors.New("kubernetesConfig is nil, mizu cli will not work with in-cluster kubernetes config, use a kubeconfig file when initializing the Provider")
	}
	ns, _, err := provider.kubernetesConfig.Namespace()
	return ns, err
}

func (provider *Provider) WaitUtilNamespaceDeleted(ctx context.Context, name string) error {
	fieldSelector := fmt.Sprintf("metadata.name=%s", name)
	var limit int64 = 1
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			options.FieldSelector = fieldSelector
			options.Limit = limit
			return provider.clientSet.CoreV1().Namespaces().List(ctx, options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			options.FieldSelector = fieldSelector
			options.Limit = limit
			return provider.clientSet.CoreV1().Namespaces().Watch(ctx, options)
		},
	}

	var preconditionFunc watchtools.PreconditionFunc = func(store cache.Store) (bool, error) {
		_, exists, err := store.Get(&core.Namespace{ObjectMeta: metav1.ObjectMeta{Name: name}})
		if err != nil {
			return false, err
		}
		if exists {
			return false, nil
		}
		return true, nil
	}

	conditionFunc := func(e watch.Event) (bool, error) {
		if e.Type == watch.Deleted {
			return true, nil
		}
		return false, nil
	}

	obj := &core.Namespace{}
	_, err := watchtools.UntilWithSync(ctx, lw, obj, preconditionFunc, conditionFunc)

	return err
}

func (provider *Provider) CreateNamespace(ctx context.Context, name string) (*core.Namespace, error) {
	namespaceSpec := &core.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				LabelManagedBy: provider.managedBy,
				LabelCreatedBy: provider.createdBy,
			},
		},
	}
	return provider.clientSet.CoreV1().Namespaces().Create(ctx, namespaceSpec, metav1.CreateOptions{})
}

type ApiServerOptions struct {
	Namespace             string
	PodName               string
	PodImage              string
	KratosImage           string
	KetoImage             string
	ServiceAccountName    string
	IsNamespaceRestricted bool
	MaxEntriesDBSizeBytes int64
	Resources             shared.Resources
	ImagePullPolicy       core.PullPolicy
	LogLevel              logging.Level
	Profiler              bool
}

func (provider *Provider) GetMizuApiServerPodObject(opts *ApiServerOptions, mountVolumeClaim bool, volumeClaimName string, createAuthContainer bool) (*core.Pod, error) {
	configMapVolume := &core.ConfigMapVolumeSource{}
	configMapVolume.Name = ConfigMapName

	cpuLimit, err := resource.ParseQuantity(opts.Resources.CpuLimit)
	if err != nil {
		return nil, fmt.Errorf("invalid cpu limit for %s container", opts.PodName)
	}
	memLimit, err := resource.ParseQuantity(opts.Resources.MemoryLimit)
	if err != nil {
		return nil, fmt.Errorf("invalid memory limit for %s container", opts.PodName)
	}
	cpuRequests, err := resource.ParseQuantity(opts.Resources.CpuRequests)
	if err != nil {
		return nil, fmt.Errorf("invalid cpu request for %s container", opts.PodName)
	}
	memRequests, err := resource.ParseQuantity(opts.Resources.MemoryRequests)
	if err != nil {
		return nil, fmt.Errorf("invalid memory request for %s container", opts.PodName)
	}

	command := []string{
		"./mizuagent",
		"--api-server",
	}

	if opts.Profiler {
		command = append(command, "--profiler")
	}

	if opts.IsNamespaceRestricted {
		command = append(command, "--namespace", opts.Namespace)
	}

	volumeMounts := []core.VolumeMount{
		{
			Name:      ConfigMapName,
			MountPath: shared.ConfigDirPath,
		},
	}
	volumes := []core.Volume{
		{
			Name: ConfigMapName,
			VolumeSource: core.VolumeSource{
				ConfigMap: configMapVolume,
			},
		},
	}

	if mountVolumeClaim {
		volumes = append(volumes, core.Volume{
			Name: volumeClaimName,
			VolumeSource: core.VolumeSource{
				PersistentVolumeClaim: &core.PersistentVolumeClaimVolumeSource{
					ClaimName: volumeClaimName,
				},
			},
		})
		volumeMounts = append(volumeMounts, core.VolumeMount{
			Name:      volumeClaimName,
			MountPath: shared.DataDirPath,
		})
	}

	containers := []core.Container{
		{
			Name:            opts.PodName,
			Image:           opts.PodImage,
			ImagePullPolicy: opts.ImagePullPolicy,
			VolumeMounts:    volumeMounts,
			Command:         command,
			Env: []core.EnvVar{
				{
					Name:  shared.LogLevelEnvVar,
					Value: opts.LogLevel.String(),
				},
			},
			Resources: core.ResourceRequirements{
				Limits: core.ResourceList{
					"cpu":    cpuLimit,
					"memory": memLimit,
				},
				Requests: core.ResourceList{
					"cpu":    cpuRequests,
					"memory": memRequests,
				},
			},
		},
		{
			Name:            "basenine",
			Image:           opts.PodImage,
			ImagePullPolicy: opts.ImagePullPolicy,
			VolumeMounts:    volumeMounts,
			ReadinessProbe: &core.Probe{
				FailureThreshold: 3,
				ProbeHandler: core.ProbeHandler{
					TCPSocket: &core.TCPSocketAction{
						Port: intstr.Parse(shared.BaseninePort),
					},
				},
				PeriodSeconds:    1,
				SuccessThreshold: 1,
				TimeoutSeconds:   1,
			},
			Resources: core.ResourceRequirements{
				Limits: core.ResourceList{
					"cpu":    cpuLimit,
					"memory": memLimit,
				},
				Requests: core.ResourceList{
					"cpu":    cpuRequests,
					"memory": memRequests,
				},
			},
			Command:    []string{"basenine"},
			Args:       []string{"-addr", "0.0.0.0", "-port", shared.BaseninePort, "-persistent"},
			WorkingDir: shared.DataDirPath,
		},
	}

	if createAuthContainer {
		containers = append(containers, core.Container{
			Name:            "kratos",
			Image:           opts.KratosImage,
			ImagePullPolicy: opts.ImagePullPolicy,
			VolumeMounts:    volumeMounts,
			ReadinessProbe: &core.Probe{
				FailureThreshold: 3,
				ProbeHandler: core.ProbeHandler{
					HTTPGet: &core.HTTPGetAction{
						Path:   "/health/ready",
						Port:   intstr.FromInt(4433),
						Scheme: core.URISchemeHTTP,
					},
				},
				PeriodSeconds:    1,
				SuccessThreshold: 1,
				TimeoutSeconds:   1,
			},
			Resources: core.ResourceRequirements{
				Limits: core.ResourceList{
					"cpu":    cpuLimit,
					"memory": memLimit,
				},
				Requests: core.ResourceList{
					"cpu":    cpuRequests,
					"memory": memRequests,
				},
			},
		})

		containers = append(containers, core.Container{
			Name:            "keto",
			Image:           opts.KetoImage,
			ImagePullPolicy: opts.ImagePullPolicy,
			VolumeMounts:    volumeMounts,
			ReadinessProbe: &core.Probe{
				FailureThreshold: 3,
				ProbeHandler: core.ProbeHandler{
					HTTPGet: &core.HTTPGetAction{
						Path:   "/health/ready",
						Port:   intstr.FromInt(4466),
						Scheme: core.URISchemeHTTP,
					},
				},
				PeriodSeconds:    1,
				SuccessThreshold: 1,
				TimeoutSeconds:   1,
			},
			Resources: core.ResourceRequirements{
				Limits: core.ResourceList{
					"cpu":    cpuLimit,
					"memory": memLimit,
				},
				Requests: core.ResourceList{
					"cpu":    cpuRequests,
					"memory": memRequests,
				},
			},
		})
	}

	pod := &core.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: opts.PodName,
			Labels: map[string]string{
				"app":          opts.PodName,
				LabelManagedBy: provider.managedBy,
				LabelCreatedBy: provider.createdBy,
			},
		},
		Spec: core.PodSpec{
			Containers:                    containers,
			Volumes:                       volumes,
			DNSPolicy:                     core.DNSClusterFirstWithHostNet,
			TerminationGracePeriodSeconds: new(int64),
			Tolerations: []core.Toleration{
				{
					Operator: core.TolerationOpExists,
					Effect: core.TaintEffectNoExecute,
				},
				{
					Operator: core.TolerationOpExists,
					Effect: core.TaintEffectNoSchedule,
				},
			},
		},
	}

	//define the service account only when it exists to prevent pod crash
	if opts.ServiceAccountName != "" {
		pod.Spec.ServiceAccountName = opts.ServiceAccountName
	}
	return pod, nil
}

func (provider *Provider) CreatePod(ctx context.Context, namespace string, podSpec *core.Pod) (*core.Pod, error) {
	return provider.clientSet.CoreV1().Pods(namespace).Create(ctx, podSpec, metav1.CreateOptions{})
}

func (provider *Provider) CreateService(ctx context.Context, namespace string, serviceName string, appLabelValue string) (*core.Service, error) {
	service := core.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceName,
			Labels: map[string]string{
				LabelManagedBy: provider.managedBy,
				LabelCreatedBy: provider.createdBy,
			},
		},
		Spec: core.ServiceSpec{
			Ports:    []core.ServicePort{{TargetPort: intstr.FromInt(shared.DefaultApiServerPort), Port: 80, Name: "api"}},
			Type:     core.ServiceTypeClusterIP,
			Selector: map[string]string{"app": appLabelValue},
		},
	}
	return provider.clientSet.CoreV1().Services(namespace).Create(ctx, &service, metav1.CreateOptions{})
}

func (provider *Provider) CanI(ctx context.Context, namespace string, resource string, verb string, group string) (bool, error) {
	selfSubjectAccessReview := &auth.SelfSubjectAccessReview{
		Spec: auth.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &auth.ResourceAttributes{
				Namespace: namespace,
				Resource:  resource,
				Verb:      verb,
				Group:     group,
			},
		},
	}

	response, err := provider.clientSet.AuthorizationV1().SelfSubjectAccessReviews().Create(ctx, selfSubjectAccessReview, metav1.CreateOptions{})
	if err != nil {
		return false, err
	}

	return response.Status.Allowed, nil
}

func (provider *Provider) DoesNamespaceExist(ctx context.Context, name string) (bool, error) {
	namespaceResource, err := provider.clientSet.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	return provider.doesResourceExist(namespaceResource, err)
}

func (provider *Provider) DoesConfigMapExist(ctx context.Context, namespace string, name string) (bool, error) {
	configMapResource, err := provider.clientSet.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
	return provider.doesResourceExist(configMapResource, err)
}

func (provider *Provider) DoesServiceAccountExist(ctx context.Context, namespace string, name string) (bool, error) {
	serviceAccountResource, err := provider.clientSet.CoreV1().ServiceAccounts(namespace).Get(ctx, name, metav1.GetOptions{})
	return provider.doesResourceExist(serviceAccountResource, err)
}

func (provider *Provider) DoesServiceExist(ctx context.Context, namespace string, name string) (bool, error) {
	serviceResource, err := provider.clientSet.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	return provider.doesResourceExist(serviceResource, err)
}

func (provider *Provider) DoesClusterRoleExist(ctx context.Context, name string) (bool, error) {
	clusterRoleResource, err := provider.clientSet.RbacV1().ClusterRoles().Get(ctx, name, metav1.GetOptions{})
	return provider.doesResourceExist(clusterRoleResource, err)
}

func (provider *Provider) DoesClusterRoleBindingExist(ctx context.Context, name string) (bool, error) {
	clusterRoleBindingResource, err := provider.clientSet.RbacV1().ClusterRoleBindings().Get(ctx, name, metav1.GetOptions{})
	return provider.doesResourceExist(clusterRoleBindingResource, err)
}

func (provider *Provider) DoesRoleExist(ctx context.Context, namespace string, name string) (bool, error) {
	roleResource, err := provider.clientSet.RbacV1().Roles(namespace).Get(ctx, name, metav1.GetOptions{})
	return provider.doesResourceExist(roleResource, err)
}

func (provider *Provider) DoesRoleBindingExist(ctx context.Context, namespace string, name string) (bool, error) {
	roleBindingResource, err := provider.clientSet.RbacV1().RoleBindings(namespace).Get(ctx, name, metav1.GetOptions{})
	return provider.doesResourceExist(roleBindingResource, err)
}

func (provider *Provider) doesResourceExist(resource interface{}, err error) (bool, error) {
	// Getting NotFound error is the expected behavior when a resource does not exist.
	if k8serrors.IsNotFound(err) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return resource != nil, nil
}

func (provider *Provider) CreateMizuRBAC(ctx context.Context, namespace string, serviceAccountName string, clusterRoleName string, clusterRoleBindingName string, version string, resources []string) error {
	serviceAccount := &core.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceAccountName,
			Labels: map[string]string{
				"mizu-cli-version": version,
				LabelManagedBy:     provider.managedBy,
				LabelCreatedBy:     provider.createdBy,
			},
		},
	}
	clusterRole := &rbac.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterRoleName,
			Labels: map[string]string{
				"mizu-cli-version": version,
				LabelManagedBy:     provider.managedBy,
				LabelCreatedBy:     provider.createdBy,
			},
		},
		Rules: []rbac.PolicyRule{
			{
				APIGroups: []string{"", "extensions", "apps"},
				Resources: resources,
				Verbs:     []string{"list", "get", "watch"},
			},
		},
	}
	clusterRoleBinding := &rbac.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterRoleBindingName,
			Labels: map[string]string{
				"mizu-cli-version": version,
				LabelManagedBy:     provider.managedBy,
				LabelCreatedBy:     provider.createdBy,
			},
		},
		RoleRef: rbac.RoleRef{
			Name:     clusterRoleName,
			Kind:     "ClusterRole",
			APIGroup: "rbac.authorization.k8s.io",
		},
		Subjects: []rbac.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      serviceAccountName,
				Namespace: namespace,
			},
		},
	}
	_, err := provider.clientSet.CoreV1().ServiceAccounts(namespace).Create(ctx, serviceAccount, metav1.CreateOptions{})
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return err
	}
	_, err = provider.clientSet.RbacV1().ClusterRoles().Create(ctx, clusterRole, metav1.CreateOptions{})
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return err
	}
	_, err = provider.clientSet.RbacV1().ClusterRoleBindings().Create(ctx, clusterRoleBinding, metav1.CreateOptions{})
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func (provider *Provider) CreateMizuRBACNamespaceRestricted(ctx context.Context, namespace string, serviceAccountName string, roleName string, roleBindingName string, version string) error {
	serviceAccount := &core.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceAccountName,
			Labels: map[string]string{
				"mizu-cli-version": version,
				LabelManagedBy:     provider.managedBy,
				LabelCreatedBy:     provider.createdBy,
			},
		},
	}
	role := &rbac.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name: roleName,
			Labels: map[string]string{
				"mizu-cli-version": version,
				LabelManagedBy:     provider.managedBy,
				LabelCreatedBy:     provider.createdBy,
			},
		},
		Rules: []rbac.PolicyRule{
			{
				APIGroups: []string{"", "extensions", "apps"},
				Resources: []string{"pods", "services", "endpoints"},
				Verbs:     []string{"list", "get", "watch"},
			},
		},
	}
	roleBinding := &rbac.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: roleBindingName,
			Labels: map[string]string{
				"mizu-cli-version": version,
				LabelManagedBy:     provider.managedBy,
				LabelCreatedBy:     provider.createdBy,
			},
		},
		RoleRef: rbac.RoleRef{
			Name:     roleName,
			Kind:     "Role",
			APIGroup: "rbac.authorization.k8s.io",
		},
		Subjects: []rbac.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      serviceAccountName,
				Namespace: namespace,
			},
		},
	}
	_, err := provider.clientSet.CoreV1().ServiceAccounts(namespace).Create(ctx, serviceAccount, metav1.CreateOptions{})
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return err
	}
	_, err = provider.clientSet.RbacV1().Roles(namespace).Create(ctx, role, metav1.CreateOptions{})
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return err
	}
	_, err = provider.clientSet.RbacV1().RoleBindings(namespace).Create(ctx, roleBinding, metav1.CreateOptions{})
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func (provider *Provider) RemoveNamespace(ctx context.Context, name string) error {
	err := provider.clientSet.CoreV1().Namespaces().Delete(ctx, name, metav1.DeleteOptions{})
	return provider.handleRemovalError(err)
}

func (provider *Provider) RemoveClusterRole(ctx context.Context, name string) error {
	err := provider.clientSet.RbacV1().ClusterRoles().Delete(ctx, name, metav1.DeleteOptions{})
	return provider.handleRemovalError(err)
}

func (provider *Provider) RemoveClusterRoleBinding(ctx context.Context, name string) error {
	err := provider.clientSet.RbacV1().ClusterRoleBindings().Delete(ctx, name, metav1.DeleteOptions{})
	return provider.handleRemovalError(err)
}

func (provider *Provider) RemoveRoleBinding(ctx context.Context, namespace string, name string) error {
	err := provider.clientSet.RbacV1().RoleBindings(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	return provider.handleRemovalError(err)
}

func (provider *Provider) RemoveRole(ctx context.Context, namespace string, name string) error {
	err := provider.clientSet.RbacV1().Roles(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	return provider.handleRemovalError(err)
}

func (provider *Provider) RemoveServiceAccount(ctx context.Context, namespace string, name string) error {
	err := provider.clientSet.CoreV1().ServiceAccounts(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	return provider.handleRemovalError(err)
}

func (provider *Provider) RemovePod(ctx context.Context, namespace string, podName string) error {
	err := provider.clientSet.CoreV1().Pods(namespace).Delete(ctx, podName, metav1.DeleteOptions{})
	return provider.handleRemovalError(err)
}

func (provider *Provider) RemoveConfigMap(ctx context.Context, namespace string, configMapName string) error {
	err := provider.clientSet.CoreV1().ConfigMaps(namespace).Delete(ctx, configMapName, metav1.DeleteOptions{})
	return provider.handleRemovalError(err)
}

func (provider *Provider) RemoveService(ctx context.Context, namespace string, serviceName string) error {
	err := provider.clientSet.CoreV1().Services(namespace).Delete(ctx, serviceName, metav1.DeleteOptions{})
	return provider.handleRemovalError(err)
}

func (provider *Provider) RemoveDaemonSet(ctx context.Context, namespace string, daemonSetName string) error {
	err := provider.clientSet.AppsV1().DaemonSets(namespace).Delete(ctx, daemonSetName, metav1.DeleteOptions{})
	return provider.handleRemovalError(err)
}

func (provider *Provider) handleRemovalError(err error) error {
	// Ignore NotFound - There is nothing to delete.
	// Ignore Forbidden - Assume that a user could not have created the resource in the first place.
	if k8serrors.IsNotFound(err) || k8serrors.IsForbidden(err) {
		return nil
	}

	return err
}

func (provider *Provider) CreateConfigMap(ctx context.Context, namespace string, configMapName string, serializedValidationRules string, serializedContract string, serializedMizuConfig string) error {
	configMapData := make(map[string]string)
	if serializedValidationRules != "" {
		configMapData[shared.ValidationRulesFileName] = serializedValidationRules
	}
	if serializedContract != "" {
		configMapData[shared.ContractFileName] = serializedContract
	}
	configMapData[shared.ConfigFileName] = serializedMizuConfig

	configMap := &core.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: configMapName,
			Labels: map[string]string{
				LabelManagedBy: provider.managedBy,
				LabelCreatedBy: provider.createdBy,
			},
		},
		Data: configMapData,
	}
	if _, err := provider.clientSet.CoreV1().ConfigMaps(namespace).Create(ctx, configMap, metav1.CreateOptions{}); err != nil {
		return err
	}
	return nil
}

func (provider *Provider) ApplyMizuTapperDaemonSet(ctx context.Context, namespace string, daemonSetName string, podImage string, tapperPodName string, apiServerPodIp string, nodeNames []string, serviceAccountName string, resources shared.Resources, imagePullPolicy core.PullPolicy, mizuApiFilteringOptions api.TrafficFilteringOptions, logLevel logging.Level, serviceMesh bool, tls bool) error {
	logger.Log.Debugf("Applying %d tapper daemon sets, ns: %s, daemonSetName: %s, podImage: %s, tapperPodName: %s", len(nodeNames), namespace, daemonSetName, podImage, tapperPodName)

	if len(nodeNames) == 0 {
		return fmt.Errorf("daemon set %s must tap at least 1 pod", daemonSetName)
	}

	mizuApiFilteringOptionsJsonStr, err := json.Marshal(mizuApiFilteringOptions)
	if err != nil {
		return err
	}

	mizuCmd := []string{
		"./mizuagent",
		"-i", "any",
		"--tap",
		"--api-server-address", fmt.Sprintf("ws://%s/wsTapper", apiServerPodIp),
		"--nodefrag",
	}

	if serviceMesh {
		mizuCmd = append(mizuCmd, "--servicemesh")
	}

	if tls {
		mizuCmd = append(mizuCmd, "--tls")
	}

	if serviceMesh || tls {
		mizuCmd = append(mizuCmd, "--procfs", procfsMountPath)
	}

	agentContainer := applyconfcore.Container()
	agentContainer.WithName(tapperPodName)
	agentContainer.WithImage(podImage)
	agentContainer.WithImagePullPolicy(imagePullPolicy)

	caps := applyconfcore.Capabilities().WithDrop("ALL")

	caps = caps.WithAdd("NET_RAW").WithAdd("NET_ADMIN") // to listen to traffic using libpcap

	if serviceMesh || tls {
		caps = caps.WithAdd("SYS_ADMIN")  // to read /proc/PID/net/ns + to install eBPF programs (kernel < 5.8)
		caps = caps.WithAdd("SYS_PTRACE") // to set netns to other process + to open libssl.so of other process

		if serviceMesh {
			caps = caps.WithAdd("DAC_OVERRIDE") // to read /proc/PID/environ
		}

		if tls {
			caps = caps.WithAdd("SYS_RESOURCE") // to change rlimits for eBPF
		}
	}

	agentContainer.WithSecurityContext(applyconfcore.SecurityContext().WithCapabilities(caps))

	agentContainer.WithCommand(mizuCmd...)
	agentContainer.WithEnv(
		applyconfcore.EnvVar().WithName(shared.LogLevelEnvVar).WithValue(logLevel.String()),
		applyconfcore.EnvVar().WithName(shared.HostModeEnvVar).WithValue("1"),
		applyconfcore.EnvVar().WithName(shared.MizuFilteringOptionsEnvVar).WithValue(string(mizuApiFilteringOptionsJsonStr)),
	)
	agentContainer.WithEnv(
		applyconfcore.EnvVar().WithName(shared.NodeNameEnvVar).WithValueFrom(
			applyconfcore.EnvVarSource().WithFieldRef(
				applyconfcore.ObjectFieldSelector().WithAPIVersion("v1").WithFieldPath("spec.nodeName"),
			),
		),
	)
	cpuLimit, err := resource.ParseQuantity(resources.CpuLimit)
	if err != nil {
		return fmt.Errorf("invalid cpu limit for %s container", tapperPodName)
	}
	memLimit, err := resource.ParseQuantity(resources.MemoryLimit)
	if err != nil {
		return fmt.Errorf("invalid memory limit for %s container", tapperPodName)
	}
	cpuRequests, err := resource.ParseQuantity(resources.CpuRequests)
	if err != nil {
		return fmt.Errorf("invalid cpu request for %s container", tapperPodName)
	}
	memRequests, err := resource.ParseQuantity(resources.MemoryRequests)
	if err != nil {
		return fmt.Errorf("invalid memory request for %s container", tapperPodName)
	}
	agentResourceLimits := core.ResourceList{
		"cpu":    cpuLimit,
		"memory": memLimit,
	}
	agentResourceRequests := core.ResourceList{
		"cpu":    cpuRequests,
		"memory": memRequests,
	}
	agentResources := applyconfcore.ResourceRequirements().WithRequests(agentResourceRequests).WithLimits(agentResourceLimits)
	agentContainer.WithResources(agentResources)

	matchFields := make([]*applyconfcore.NodeSelectorTermApplyConfiguration, 0)
	for _, nodeName := range nodeNames {
		nodeSelectorRequirement := applyconfcore.NodeSelectorRequirement()
		nodeSelectorRequirement.WithKey("metadata.name")
		nodeSelectorRequirement.WithOperator(core.NodeSelectorOpIn)
		nodeSelectorRequirement.WithValues(nodeName)

		nodeSelectorTerm := applyconfcore.NodeSelectorTerm()
		nodeSelectorTerm.WithMatchFields(nodeSelectorRequirement)
		matchFields = append(matchFields, nodeSelectorTerm)
	}

	nodeSelector := applyconfcore.NodeSelector()
	nodeSelector.WithNodeSelectorTerms(matchFields...)
	nodeAffinity := applyconfcore.NodeAffinity()
	nodeAffinity.WithRequiredDuringSchedulingIgnoredDuringExecution(nodeSelector)
	affinity := applyconfcore.Affinity()
	affinity.WithNodeAffinity(nodeAffinity)

	noExecuteToleration := applyconfcore.Toleration()
	noExecuteToleration.WithOperator(core.TolerationOpExists)
	noExecuteToleration.WithEffect(core.TaintEffectNoExecute)
	noScheduleToleration := applyconfcore.Toleration()
	noScheduleToleration.WithOperator(core.TolerationOpExists)
	noScheduleToleration.WithEffect(core.TaintEffectNoSchedule)

	// Host procfs is needed inside the container because we need access to
	//	the network namespaces of processes on the machine.
	//
	procfsVolume := applyconfcore.Volume()
	procfsVolume.WithName(procfsVolumeName).WithHostPath(applyconfcore.HostPathVolumeSource().WithPath("/proc"))
	procfsVolumeMount := applyconfcore.VolumeMount().WithName(procfsVolumeName).WithMountPath(procfsMountPath).WithReadOnly(true)
	agentContainer.WithVolumeMounts(procfsVolumeMount)

	// We need access to /sys in order to install certain eBPF tracepoints
	//
	sysfsVolume := applyconfcore.Volume()
	sysfsVolume.WithName(sysfsVolumeName).WithHostPath(applyconfcore.HostPathVolumeSource().WithPath("/sys"))
	sysfsVolumeMount := applyconfcore.VolumeMount().WithName(sysfsVolumeName).WithMountPath(sysfsMountPath).WithReadOnly(true)
	agentContainer.WithVolumeMounts(sysfsVolumeMount)

	podSpec := applyconfcore.PodSpec()
	podSpec.WithHostNetwork(true)
	podSpec.WithDNSPolicy(core.DNSClusterFirstWithHostNet)
	podSpec.WithTerminationGracePeriodSeconds(0)
	if serviceAccountName != "" {
		podSpec.WithServiceAccountName(serviceAccountName)
	}
	podSpec.WithContainers(agentContainer)
	podSpec.WithAffinity(affinity)
	podSpec.WithTolerations(noExecuteToleration, noScheduleToleration)
	podSpec.WithVolumes(procfsVolume, sysfsVolume)

	podTemplate := applyconfcore.PodTemplateSpec()
	podTemplate.WithLabels(map[string]string{
		"app":          tapperPodName,
		LabelManagedBy: provider.managedBy,
		LabelCreatedBy: provider.createdBy,
	})
	podTemplate.WithSpec(podSpec)

	labelSelector := applyconfmeta.LabelSelector()
	labelSelector.WithMatchLabels(map[string]string{"app": tapperPodName})

	applyOptions := metav1.ApplyOptions{
		Force:        true,
		FieldManager: fieldManagerName,
	}

	daemonSet := applyconfapp.DaemonSet(daemonSetName, namespace)
	daemonSet.
		WithLabels(map[string]string{
			LabelManagedBy: provider.managedBy,
			LabelCreatedBy: provider.createdBy,
		}).
		WithSpec(applyconfapp.DaemonSetSpec().WithSelector(labelSelector).WithTemplate(podTemplate))

	_, err = provider.clientSet.AppsV1().DaemonSets(namespace).Apply(ctx, daemonSet, applyOptions)
	return err
}

func (provider *Provider) ResetMizuTapperDaemonSet(ctx context.Context, namespace string, daemonSetName string, podImage string, tapperPodName string) error {
	agentContainer := applyconfcore.Container()
	agentContainer.WithName(tapperPodName)
	agentContainer.WithImage(podImage)

	nodeSelectorRequirement := applyconfcore.NodeSelectorRequirement()
	nodeSelectorRequirement.WithKey("mizu-non-existing-label")
	nodeSelectorRequirement.WithOperator(core.NodeSelectorOpExists)
	nodeSelectorTerm := applyconfcore.NodeSelectorTerm()
	nodeSelectorTerm.WithMatchExpressions(nodeSelectorRequirement)
	nodeSelector := applyconfcore.NodeSelector()
	nodeSelector.WithNodeSelectorTerms(nodeSelectorTerm)
	nodeAffinity := applyconfcore.NodeAffinity()
	nodeAffinity.WithRequiredDuringSchedulingIgnoredDuringExecution(nodeSelector)
	affinity := applyconfcore.Affinity()
	affinity.WithNodeAffinity(nodeAffinity)

	podSpec := applyconfcore.PodSpec()
	podSpec.WithContainers(agentContainer)
	podSpec.WithAffinity(affinity)

	podTemplate := applyconfcore.PodTemplateSpec()
	podTemplate.WithLabels(map[string]string{
		"app":          tapperPodName,
		LabelManagedBy: provider.managedBy,
		LabelCreatedBy: provider.createdBy,
	})
	podTemplate.WithSpec(podSpec)

	labelSelector := applyconfmeta.LabelSelector()
	labelSelector.WithMatchLabels(map[string]string{"app": tapperPodName})

	applyOptions := metav1.ApplyOptions{
		Force:        true,
		FieldManager: fieldManagerName,
	}

	daemonSet := applyconfapp.DaemonSet(daemonSetName, namespace)
	daemonSet.
		WithLabels(map[string]string{
			LabelManagedBy: provider.managedBy,
			LabelCreatedBy: provider.createdBy,
		}).
		WithSpec(applyconfapp.DaemonSetSpec().WithSelector(labelSelector).WithTemplate(podTemplate))

	_, err := provider.clientSet.AppsV1().DaemonSets(namespace).Apply(ctx, daemonSet, applyOptions)
	return err
}

func (provider *Provider) listPodsImpl(ctx context.Context, regex *regexp.Regexp, namespaces []string, listOptions metav1.ListOptions) ([]core.Pod, error) {
	var pods []core.Pod
	for _, namespace := range namespaces {
		namespacePods, err := provider.clientSet.CoreV1().Pods(namespace).List(ctx, listOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to get pods in ns: [%s], %w", namespace, err)
		}

		pods = append(pods, namespacePods.Items...)
	}

	matchingPods := make([]core.Pod, 0)
	for _, pod := range pods {
		if regex.MatchString(pod.Name) {
			matchingPods = append(matchingPods, pod)
		}
	}
	return matchingPods, nil
}

func (provider *Provider) ListAllPodsMatchingRegex(ctx context.Context, regex *regexp.Regexp, namespaces []string) ([]core.Pod, error) {
	return provider.listPodsImpl(ctx, regex, namespaces, metav1.ListOptions{})
}

func (provider *Provider) GetPod(ctx context.Context, namespaces string, podName string) (*core.Pod, error) {
	return provider.clientSet.CoreV1().Pods(namespaces).Get(ctx, podName, metav1.GetOptions{})
}

func (provider *Provider) ListAllRunningPodsMatchingRegex(ctx context.Context, regex *regexp.Regexp, namespaces []string) ([]core.Pod, error) {
	pods, err := provider.ListAllPodsMatchingRegex(ctx, regex, namespaces)
	if err != nil {
		return nil, err
	}

	matchingPods := make([]core.Pod, 0)
	for _, pod := range pods {
		if IsPodRunning(&pod) {
			matchingPods = append(matchingPods, pod)
		}
	}
	return matchingPods, nil
}

func (provider *Provider) ListPodsByAppLabel(ctx context.Context, namespaces string, labelName string) ([]core.Pod, error) {
	pods, err := provider.clientSet.CoreV1().Pods(namespaces).List(ctx, metav1.ListOptions{LabelSelector: fmt.Sprintf("app=%s", labelName)})
	if err != nil {
		return nil, err
	}

	return pods.Items, err
}

func (provider *Provider) ListAllNamespaces(ctx context.Context) ([]core.Namespace, error) {
	namespaces, err := provider.clientSet.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return namespaces.Items, err
}

func (provider *Provider) GetPodLogs(ctx context.Context, namespace string, podName string, containerName string) (string, error) {
	podLogOpts := core.PodLogOptions{Container: containerName}
	req := provider.clientSet.CoreV1().Pods(namespace).GetLogs(podName, &podLogOpts)
	podLogs, err := req.Stream(ctx)
	if err != nil {
		return "", fmt.Errorf("error opening log stream on ns: %s, pod: %s, %w", namespace, podName, err)
	}
	defer podLogs.Close()
	buf := new(bytes.Buffer)
	if _, err = io.Copy(buf, podLogs); err != nil {
		return "", fmt.Errorf("error copy information from podLogs to buf, ns: %s, pod: %s, %w", namespace, podName, err)
	}
	str := buf.String()
	return str, nil
}

func (provider *Provider) GetNamespaceEvents(ctx context.Context, namespace string) (string, error) {
	eventList, err := provider.clientSet.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("error getting events on ns: %s, %w", namespace, err)
	}

	return eventList.String(), nil
}

func (provider *Provider) ListManagedServiceAccounts(ctx context.Context, namespace string) (*core.ServiceAccountList, error) {
	listOptions := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", LabelManagedBy, provider.managedBy),
	}
	return provider.clientSet.CoreV1().ServiceAccounts(namespace).List(ctx, listOptions)
}

func (provider *Provider) ListManagedClusterRoles(ctx context.Context) (*rbac.ClusterRoleList, error) {
	listOptions := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", LabelManagedBy, provider.managedBy),
	}
	return provider.clientSet.RbacV1().ClusterRoles().List(ctx, listOptions)
}

func (provider *Provider) ListManagedClusterRoleBindings(ctx context.Context) (*rbac.ClusterRoleBindingList, error) {
	listOptions := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", LabelManagedBy, provider.managedBy),
	}
	return provider.clientSet.RbacV1().ClusterRoleBindings().List(ctx, listOptions)
}

func (provider *Provider) ListManagedRoles(ctx context.Context, namespace string) (*rbac.RoleList, error) {
	listOptions := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", LabelManagedBy, provider.managedBy),
	}
	return provider.clientSet.RbacV1().Roles(namespace).List(ctx, listOptions)
}

func (provider *Provider) ListManagedRoleBindings(ctx context.Context, namespace string) (*rbac.RoleBindingList, error) {
	listOptions := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", LabelManagedBy, provider.managedBy),
	}
	return provider.clientSet.RbacV1().RoleBindings(namespace).List(ctx, listOptions)
}

// ValidateNotProxy We added this after a customer tried to run mizu from lens, which used len's kube config, which have cluster server configuration, which points to len's local proxy.
// The workaround was to use the user's local default kube config.
// For now - we are blocking the option to run mizu through a proxy to k8s server
func (provider *Provider) ValidateNotProxy() error {
	kubernetesUrl, err := url.Parse(provider.clientConfig.Host)
	if err != nil {
		logger.Log.Debugf("ValidateNotProxy - error while parsing kubernetes host, err: %v", err)
		return nil
	}

	restProxyClientConfig, _ := provider.kubernetesConfig.ClientConfig()
	restProxyClientConfig.Host = kubernetesUrl.Host

	clientProxySet, err := getClientSet(restProxyClientConfig)
	if err == nil {
		proxyServerVersion, err := clientProxySet.ServerVersion()
		if err != nil {
			return nil
		}

		if *proxyServerVersion == (version.Info{}) {
			return &ClusterBehindProxyError{}
		}
	}

	return nil
}

func (provider *Provider) GetKubernetesVersion() (*semver.SemVersion, error) {
	serverVersion, err := provider.clientSet.ServerVersion()
	if err != nil {
		logger.Log.Debugf("error while getting kubernetes server version, err: %v", err)
		return nil, err
	}

	serverVersionSemVer := semver.SemVersion(serverVersion.GitVersion)
	return &serverVersionSemVer, nil
}

func getClientSet(config *rest.Config) (*kubernetes.Clientset, error) {
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientSet, nil
}

func ValidateKubernetesVersion(serverVersionSemVer *semver.SemVersion) error {
	minKubernetesServerVersionSemVer := semver.SemVersion(MinKubernetesServerVersion)
	if minKubernetesServerVersionSemVer.GreaterThan(*serverVersionSemVer) {
		return fmt.Errorf("kubernetes server version %v is not supported, supporting only kubernetes server version of %v or higher", serverVersionSemVer, MinKubernetesServerVersion)
	}

	return nil
}

func loadKubernetesConfiguration(kubeConfigPath string, context string) clientcmd.ClientConfig {
	logger.Log.Debugf("Using kube config %s", kubeConfigPath)
	configPathList := filepath.SplitList(kubeConfigPath)
	configLoadingRules := &clientcmd.ClientConfigLoadingRules{}
	if len(configPathList) <= 1 {
		configLoadingRules.ExplicitPath = kubeConfigPath
	} else {
		configLoadingRules.Precedence = configPathList
	}
	contextName := context
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		configLoadingRules,
		&clientcmd.ConfigOverrides{
			CurrentContext: contextName,
		},
	)
}

func IsPodRunning(pod *core.Pod) bool {
	return pod.Status.Phase == core.PodRunning
}
