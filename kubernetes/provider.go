package kubernetes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/config/configStructs"
	"github.com/kubeshark/kubeshark/docker"
	"github.com/kubeshark/kubeshark/misc"
	"github.com/kubeshark/kubeshark/semver"
	"github.com/kubeshark/kubeshark/utils"
	"github.com/rs/zerolog/log"
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
	fieldManagerName = "kubeshark-manager"
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
				"you can set alternative kube config file path by adding the kube-config-path field to the %s config file, err:  %w", kubeConfigPath, misc.Program, err)
		}
		if clientcmd.IsConfigurationInvalid(err) {
			return nil, fmt.Errorf("invalid kube config file (%s)\n"+
				"you can set alternative kube config file path by adding the kube-config-path field to the %s config file, err:  %w", kubeConfigPath, misc.Program, err)
		}

		return nil, fmt.Errorf("error while using kube config (%s)\n"+
			"you can set alternative kube config file path by adding the kube-config-path field to the %s config file, err:  %w", kubeConfigPath, misc.Program, err)
	}

	clientSet, err := getClientSet(restClientConfig)
	if err != nil {
		return nil, fmt.Errorf("error while using kube config (%s)\n"+
			"you can set alternative kube config file path by adding the kube-config-path field to the %s config file, err:  %w", kubeConfigPath, misc.Program, err)
	}

	log.Debug().
		Str("host", restClientConfig.Host).
		Str("api-path", restClientConfig.APIPath).
		Str("user-agent", restClientConfig.UserAgent).
		Msg("K8s client config.")

	return &Provider{
		clientSet:        clientSet,
		kubernetesConfig: kubernetesConfig,
		clientConfig:     *restClientConfig,
		managedBy:        misc.Program,
		createdBy:        misc.Program,
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
		managedBy:        misc.Program,
		createdBy:        misc.Program,
	}, nil
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

func (provider *Provider) BuildNamespace(name string) *core.Namespace {
	return &core.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: buildWithDefaultLabels(map[string]string{}, provider),
		},
	}
}

func (provider *Provider) CreateNamespace(ctx context.Context, namespace *core.Namespace) (*core.Namespace, error) {
	return provider.clientSet.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
}

type PodOptions struct {
	Namespace          string
	PodName            string
	PodImage           string
	ServiceAccountName string
	Resources          configStructs.ResourceRequirements
	ImagePullPolicy    core.PullPolicy
	ImagePullSecrets   []core.LocalObjectReference
	Debug              bool
}

func (provider *Provider) BuildHubPod(opts *PodOptions) (*core.Pod, error) {
	cpuLimit, err := resource.ParseQuantity(opts.Resources.Limits.CPU)
	if err != nil {
		return nil, fmt.Errorf("invalid cpu limit for %s pod", opts.PodName)
	}
	memLimit, err := resource.ParseQuantity(opts.Resources.Limits.Memory)
	if err != nil {
		return nil, fmt.Errorf("invalid memory limit for %s pod", opts.PodName)
	}
	cpuRequests, err := resource.ParseQuantity(opts.Resources.Requests.CPU)
	if err != nil {
		return nil, fmt.Errorf("invalid cpu request for %s pod", opts.PodName)
	}
	memRequests, err := resource.ParseQuantity(opts.Resources.Requests.Memory)
	if err != nil {
		return nil, fmt.Errorf("invalid memory request for %s pod", opts.PodName)
	}

	command := []string{
		"./hub",
	}

	if opts.Debug {
		command = append(command, "-debug")
	}

	// Scripting environment variables
	scriptingEnvMarshalled, err := json.Marshal(config.Config.Scripting.Env)
	if err != nil {
		return nil, err
	}

	// Scripting scripts
	scripts, err := config.Config.Scripting.GetScripts()
	if err != nil {
		return nil, err
	}
	if scripts == nil {
		scripts = []*misc.Script{}
	}
	scriptsMarshalled, err := json.Marshal(scripts)
	if err != nil {
		return nil, err
	}

	containers := []core.Container{
		{
			Name:            opts.PodName,
			Image:           opts.PodImage,
			ImagePullPolicy: opts.ImagePullPolicy,
			Command:         command,
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
			Env: []core.EnvVar{
				{
					Name:  "POD_REGEX",
					Value: config.Config.Tap.PodRegexStr,
				},
				{
					Name:  "NAMESPACES",
					Value: strings.Join(provider.GetNamespaces(), ","),
				},
				{
					Name:  "LICENSE",
					Value: "",
				},
				{
					Name:  "SCRIPTING_ENV",
					Value: string(scriptingEnvMarshalled),
				},
				{
					Name:  "SCRIPTING_SCRIPTS",
					Value: string(scriptsMarshalled),
				},
			},
		},
	}

	pod := &core.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      opts.PodName,
			Namespace: config.Config.Tap.SelfNamespace,
			Labels: buildWithDefaultLabels(map[string]string{
				"app": opts.PodName,
			}, provider),
		},
		Spec: core.PodSpec{
			ServiceAccountName:            ServiceAccountName,
			Containers:                    containers,
			DNSPolicy:                     core.DNSClusterFirstWithHostNet,
			TerminationGracePeriodSeconds: new(int64),
			Tolerations:                   provider.BuildTolerations(),
			ImagePullSecrets:              opts.ImagePullSecrets,
		},
	}

	if len(config.Config.Tap.NodeSelectorTerms) > 0 {
		pod.Spec.Affinity = &core.Affinity{
			NodeAffinity: &core.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &core.NodeSelector{
					NodeSelectorTerms: config.Config.Tap.NodeSelectorTerms,
				},
			},
		}
	}

	return pod, nil
}

func (provider *Provider) BuildFrontPod(opts *PodOptions, hubHost string, hubPort string) (*core.Pod, error) {
	cpuLimit, err := resource.ParseQuantity(opts.Resources.Limits.CPU)
	if err != nil {
		return nil, fmt.Errorf("invalid cpu limit for %s pod", opts.PodName)
	}
	memLimit, err := resource.ParseQuantity(opts.Resources.Limits.Memory)
	if err != nil {
		return nil, fmt.Errorf("invalid memory limit for %s pod", opts.PodName)
	}
	cpuRequests, err := resource.ParseQuantity(opts.Resources.Requests.CPU)
	if err != nil {
		return nil, fmt.Errorf("invalid cpu request for %s pod", opts.PodName)
	}
	memRequests, err := resource.ParseQuantity(opts.Resources.Requests.Memory)
	if err != nil {
		return nil, fmt.Errorf("invalid memory request for %s pod", opts.PodName)
	}

	volumeMounts := []core.VolumeMount{}
	volumes := []core.Volume{}

	containers := []core.Container{
		{
			Name:            opts.PodName,
			Image:           docker.GetFrontImage(),
			ImagePullPolicy: opts.ImagePullPolicy,
			VolumeMounts:    volumeMounts,
			ReadinessProbe: &core.Probe{
				FailureThreshold: 3,
				ProbeHandler: core.ProbeHandler{
					TCPSocket: &core.TCPSocketAction{
						Port: intstr.Parse("80"),
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
			Env: []core.EnvVar{
				{
					Name:  "REACT_APP_DEFAULT_FILTER",
					Value: " ",
				},
				{
					Name:  "REACT_APP_HUB_HOST",
					Value: " ",
				},
				{
					Name:  "REACT_APP_HUB_PORT",
					Value: hubPort,
				},
			},
		},
	}

	pod := &core.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      opts.PodName,
			Namespace: config.Config.Tap.SelfNamespace,
			Labels: buildWithDefaultLabels(map[string]string{
				"app": opts.PodName,
			}, provider),
		},
		Spec: core.PodSpec{
			ServiceAccountName:            ServiceAccountName,
			Containers:                    containers,
			Volumes:                       volumes,
			DNSPolicy:                     core.DNSClusterFirstWithHostNet,
			TerminationGracePeriodSeconds: new(int64),
			Tolerations:                   provider.BuildTolerations(),
			ImagePullSecrets:              opts.ImagePullSecrets,
		},
	}

	if len(config.Config.Tap.NodeSelectorTerms) > 0 {
		pod.Spec.Affinity = &core.Affinity{
			NodeAffinity: &core.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &core.NodeSelector{
					NodeSelectorTerms: config.Config.Tap.NodeSelectorTerms,
				},
			},
		}
	}

	return pod, nil
}

func (provider *Provider) CreatePod(ctx context.Context, namespace string, podSpec *core.Pod) (*core.Pod, error) {
	return provider.clientSet.CoreV1().Pods(namespace).Create(ctx, podSpec, metav1.CreateOptions{})
}

func (provider *Provider) BuildHubService(namespace string) *core.Service {
	return &core.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      HubServiceName,
			Namespace: config.Config.Tap.SelfNamespace,
			Labels:    buildWithDefaultLabels(map[string]string{}, provider),
		},
		Spec: core.ServiceSpec{
			Ports: []core.ServicePort{
				{
					Name:       HubServiceName,
					TargetPort: intstr.FromInt(80),
					Port:       80,
				},
			},
			Type:     core.ServiceTypeClusterIP,
			Selector: map[string]string{"app": HubServiceName},
		},
	}
}

func (provider *Provider) BuildFrontService(namespace string) *core.Service {
	return &core.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      FrontServiceName,
			Namespace: config.Config.Tap.SelfNamespace,
			Labels:    buildWithDefaultLabels(map[string]string{}, provider),
		},
		Spec: core.ServiceSpec{
			Ports: []core.ServicePort{
				{
					Name:       FrontServiceName,
					TargetPort: intstr.FromInt(80),
					Port:       80,
				},
			},
			Type:     core.ServiceTypeClusterIP,
			Selector: map[string]string{"app": FrontServiceName},
		},
	}
}

func (provider *Provider) CreateService(ctx context.Context, namespace string, service *core.Service) (*core.Service, error) {
	return provider.clientSet.CoreV1().Services(namespace).Create(ctx, service, metav1.CreateOptions{})
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

func (provider *Provider) BuildServiceAccount() *core.ServiceAccount {
	return &core.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ServiceAccountName,
			Namespace: config.Config.Tap.SelfNamespace,
			Labels: buildWithDefaultLabels(map[string]string{
				fmt.Sprintf("%s-cli-version", misc.Program): misc.RBACVersion,
			}, provider),
		},
	}
}

func (provider *Provider) BuildClusterRole() *rbac.ClusterRole {
	return &rbac.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ClusterRoleName,
			Namespace: config.Config.Tap.SelfNamespace,
			Labels: buildWithDefaultLabels(map[string]string{
				fmt.Sprintf("%s-cli-version", misc.Program): misc.RBACVersion,
			}, provider),
		},
		Rules: []rbac.PolicyRule{
			{
				APIGroups: []string{
					"",
					"extensions",
					"apps",
				},
				Resources: []string{
					"pods",
					"services",
					"endpoints",
					"persistentvolumeclaims",
				},
				Verbs: []string{
					"list",
					"get",
					"watch",
				},
			},
		},
	}
}

func (provider *Provider) BuildClusterRoleBinding() *rbac.ClusterRoleBinding {
	return &rbac.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ClusterRoleBindingName,
			Namespace: config.Config.Tap.SelfNamespace,
			Labels: buildWithDefaultLabels(map[string]string{
				fmt.Sprintf("%s-cli-version", misc.Program): misc.RBACVersion,
			}, provider),
		},
		RoleRef: rbac.RoleRef{
			Name:     ClusterRoleName,
			Kind:     "ClusterRole",
			APIGroup: "rbac.authorization.k8s.io",
		},
		Subjects: []rbac.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      ServiceAccountName,
				Namespace: config.Config.Tap.SelfNamespace,
			},
		},
	}
}

func (provider *Provider) CreateSelfRBAC(ctx context.Context, namespace string) error {
	serviceAccount := provider.BuildServiceAccount()
	clusterRole := provider.BuildClusterRole()
	clusterRoleBinding := provider.BuildClusterRoleBinding()

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

func (provider *Provider) RemovePersistentVolumeClaim(ctx context.Context, namespace string, persistentVolumeClaimName string) error {
	err := provider.clientSet.CoreV1().PersistentVolumeClaims(namespace).Delete(ctx, persistentVolumeClaimName, metav1.DeleteOptions{})
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

func (provider *Provider) BuildPersistentVolumeClaim() (*core.PersistentVolumeClaim, error) {
	capacity, err := resource.ParseQuantity(config.Config.Tap.StorageLimit)
	if err != nil {
		return nil, fmt.Errorf("invalid capacity for the workers: %s", config.Config.Tap.StorageLimit)
	}

	storageClassName := config.Config.Tap.StorageClass

	return &core.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      PersistentVolumeClaimName,
			Namespace: config.Config.Tap.SelfNamespace,
			Labels: buildWithDefaultLabels(map[string]string{
				fmt.Sprintf("%s-cli-version", misc.Program): misc.RBACVersion,
			}, provider),
		},
		Spec: core.PersistentVolumeClaimSpec{
			Resources: core.ResourceRequirements{
				Requests: core.ResourceList{
					core.ResourceStorage: capacity,
				},
			},
			AccessModes:      []core.PersistentVolumeAccessMode{core.ReadWriteMany},
			StorageClassName: &storageClassName,
		},
	}, nil
}

func (provider *Provider) BuildWorkerDaemonSet(
	podImage string,
	podName string,
	serviceAccountName string,
	resources configStructs.ResourceRequirements,
	imagePullPolicy core.PullPolicy,
	imagePullSecrets []core.LocalObjectReference,
	serviceMesh bool,
	tls bool,
	debug bool,
) (*DaemonSet, error) {
	// Resource limits
	cpuLimit, err := resource.ParseQuantity(resources.Limits.CPU)
	if err != nil {
		return nil, fmt.Errorf("invalid cpu limit for %s pod", podName)
	}
	memLimit, err := resource.ParseQuantity(resources.Limits.Memory)
	if err != nil {
		return nil, fmt.Errorf("invalid memory limit for %s pod", podName)
	}
	cpuRequests, err := resource.ParseQuantity(resources.Requests.CPU)
	if err != nil {
		return nil, fmt.Errorf("invalid cpu request for %s pod", podName)
	}
	memRequests, err := resource.ParseQuantity(resources.Requests.Memory)
	if err != nil {
		return nil, fmt.Errorf("invalid memory request for %s pod", podName)
	}

	// Command
	command := []string{
		"./worker",
		"-i",
		"any",
		"-port",
		"8897",
		"-packet-capture",
		config.Config.Tap.PacketCapture,
	}
	if debug {
		command = append(command, "-debug")
	}
	if serviceMesh {
		command = append(command, "-servicemesh")
	}
	if tls {
		command = append(command, "-tls")
	}
	if serviceMesh || tls {
		command = append(command, "-procfs", procfsMountPath)
	}

	// Linux capabilities
	dropCaps := []core.Capability{"ALL"}
	addCaps := []core.Capability{"NET_RAW", "NET_ADMIN"} // to listen to traffic using libpcap
	if serviceMesh || tls {
		addCaps = append(addCaps, "SYS_ADMIN")  // to read /proc/PID/net/ns + to install eBPF programs (kernel < 5.8)
		addCaps = append(addCaps, "SYS_PTRACE") // to set netns to other process + to open libssl.so of other process

		if serviceMesh {
			addCaps = append(addCaps, "DAC_OVERRIDE") // to read /proc/PID/environ
		}

		if tls {
			addCaps = append(addCaps, "SYS_RESOURCE") // to change rlimits for eBPF
		}
	}

	// Environment variables
	var env []core.EnvVar
	if debug {
		env = append(env, core.EnvVar{
			Name:  "MEMORY_PROFILING_ENABLED",
			Value: "true",
		})
		env = append(env, core.EnvVar{
			Name:  "MEMORY_PROFILING_INTERVAL_SECONDS",
			Value: "10",
		})
		env = append(env, core.EnvVar{
			Name:  "MEMORY_USAGE_INTERVAL_MILLISECONDS",
			Value: "500",
		})
	}

	// Volumes and volume mounts

	// Host procfs is needed inside the container because we need access to
	// the network namespaces of processes on the machine.
	procfsVolume := core.Volume{
		Name: procfsVolumeName,
		VolumeSource: core.VolumeSource{
			HostPath: &core.HostPathVolumeSource{
				Path: "/proc",
			},
		},
	}
	procfsVolumeMount := core.VolumeMount{
		Name:      procfsVolumeName,
		MountPath: procfsMountPath,
		ReadOnly:  true,
	}

	// We need access to /sys in order to install certain eBPF tracepoints
	sysfsVolume := core.Volume{
		Name: sysfsVolumeName,
		VolumeSource: core.VolumeSource{
			HostPath: &core.HostPathVolumeSource{
				Path: "/sys",
			},
		},
	}
	sysfsVolumeMount := core.VolumeMount{
		Name:      sysfsVolumeName,
		MountPath: sysfsMountPath,
		ReadOnly:  true,
	}

	// Persistent volume and its mount
	persistentVolume := core.Volume{
		Name: PersistentVolumeName,
		VolumeSource: core.VolumeSource{
			PersistentVolumeClaim: &core.PersistentVolumeClaimVolumeSource{
				ClaimName: PersistentVolumeClaimName,
			},
		},
	}
	persistentVolumeMount := core.VolumeMount{
		Name:      PersistentVolumeName,
		MountPath: PersistentVolumeHostPath,
	}

	// Containers
	containers := []core.Container{
		{
			Name:            podName,
			Image:           podImage,
			ImagePullPolicy: imagePullPolicy,
			VolumeMounts: []core.VolumeMount{
				procfsVolumeMount,
				sysfsVolumeMount,
				persistentVolumeMount,
			},
			Command: command,
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
			SecurityContext: &core.SecurityContext{
				Capabilities: &core.Capabilities{
					Add:  addCaps,
					Drop: dropCaps,
				},
			},
			Env: env,
		},
	}

	// Pod
	pod := DaemonSetPod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: config.Config.Tap.SelfNamespace,
			Labels: buildWithDefaultLabels(map[string]string{
				"app": podName,
			}, provider),
		},
		Spec: core.PodSpec{
			ServiceAccountName: ServiceAccountName,
			HostNetwork:        true,
			Containers:         containers,
			Volumes: []core.Volume{
				procfsVolume,
				sysfsVolume,
				persistentVolume,
			},
			DNSPolicy:                     core.DNSClusterFirstWithHostNet,
			TerminationGracePeriodSeconds: new(int64),
			Tolerations:                   provider.BuildTolerations(),
			ImagePullSecrets:              imagePullSecrets,
		},
	}

	if len(config.Config.Tap.NodeSelectorTerms) > 0 {
		pod.Spec.Affinity = &core.Affinity{
			NodeAffinity: &core.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &core.NodeSelector{
					NodeSelectorTerms: config.Config.Tap.NodeSelectorTerms,
				},
			},
		}
	}

	return &DaemonSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: config.Config.Tap.SelfNamespace,
			Labels: buildWithDefaultLabels(map[string]string{
				"app": podName,
			}, provider),
		},
		Spec: DaemonSetSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: buildWithDefaultLabels(map[string]string{
					"app": podName,
				}, provider),
			},
			Template: pod,
		},
	}, nil
}

func (provider *Provider) BuildTolerations() []core.Toleration {
	tolerations := []core.Toleration{
		{
			Operator: core.TolerationOpExists,
			Effect:   core.TaintEffectNoExecute,
		},
	}

	if !config.Config.Tap.IgnoreTainted {
		tolerations = append(tolerations, core.Toleration{
			Operator: core.TolerationOpExists,
			Effect:   core.TaintEffectNoSchedule,
		})
	}

	return tolerations
}

func (provider *Provider) CreatePersistentVolumeClaim(ctx context.Context, namespace string, persistentVolumeClaim *core.PersistentVolumeClaim) (*core.PersistentVolumeClaim, error) {
	return provider.clientSet.CoreV1().PersistentVolumeClaims(namespace).Create(ctx, persistentVolumeClaim, metav1.CreateOptions{})
}

func (provider *Provider) ApplyWorkerDaemonSet(
	ctx context.Context,
	namespace string,
	daemonSetName string,
	podImage string,
	podName string,
	serviceAccountName string,
	resources configStructs.ResourceRequirements,
	imagePullPolicy core.PullPolicy,
	imagePullSecrets []core.LocalObjectReference,
	serviceMesh bool,
	tls bool,
	debug bool,
) error {
	log.Debug().
		Str("namespace", namespace).
		Str("daemonset-name", daemonSetName).
		Str("image", podImage).
		Str("pod", podName).
		Msg("Applying worker DaemonSets.")

	daemonSet, err := provider.BuildWorkerDaemonSet(
		podImage,
		podName,
		serviceAccountName,
		resources,
		imagePullPolicy,
		imagePullSecrets,
		serviceMesh,
		tls,
		debug,
	)
	if err != nil {
		return err
	}

	applyOptions := metav1.ApplyOptions{
		Force:        true,
		FieldManager: fieldManagerName,
	}

	_, err = provider.clientSet.AppsV1().DaemonSets(namespace).Apply(
		ctx,
		daemonSet.GenerateApplyConfiguration(
			daemonSetName,
			namespace,
			podName,
			provider,
		),
		applyOptions,
	)
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

// ValidateNotProxy We added this after a customer tried to run kubeshark from lens, which used len's kube config, which have cluster server configuration, which points to len's local proxy.
// The workaround was to use the user's local default kube config.
// For now - we are blocking the option to run kubeshark through a proxy to k8s server
func (provider *Provider) ValidateNotProxy() error {
	kubernetesUrl, err := url.Parse(provider.clientConfig.Host)
	if err != nil {
		log.Debug().Err(err).Msg("While parsing Kubernetes host!")
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
		log.Debug().Err(err).Msg("While getting Kubernetes server version!")
		return nil, err
	}

	serverVersionSemVer := semver.SemVersion(serverVersion.GitVersion)
	return &serverVersionSemVer, nil
}

func (provider *Provider) GetNamespaces() []string {
	if len(config.Config.Tap.Namespaces) > 0 {
		return utils.Unique(config.Config.Tap.Namespaces)
	} else {
		return []string{K8sAllNamespaces}
	}
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
