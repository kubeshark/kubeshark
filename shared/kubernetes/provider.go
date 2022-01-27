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
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/logger"
	"github.com/up9inc/mizu/shared/semver"
	"github.com/up9inc/mizu/tap/api"
	v1 "k8s.io/api/apps/v1"
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
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	watchtools "k8s.io/client-go/tools/watch"
)

type Provider struct {
	clientSet        *kubernetes.Clientset
	kubernetesConfig clientcmd.ClientConfig
	clientConfig     restclient.Config
	Namespace        string
	managedBy        string
	createdBy        string
}

const (
	fieldManagerName = "mizu-manager"
	procfsVolumeName = "proc"
	procfsMountPath  = "/hostproc"
)

func NewProvider(kubeConfigPath string) (*Provider, error) {
	kubernetesConfig := loadKubernetesConfiguration(kubeConfigPath)
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

	return &Provider{
		clientSet:        clientSet,
		kubernetesConfig: kubernetesConfig,
		clientConfig:     *restClientConfig,
		managedBy:        LabelValueMizu,
		createdBy:        LabelValueMizuCLI,
	}, nil
}

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
	SyncEntriesConfig     *shared.SyncEntriesConfig
	MaxEntriesDBSizeBytes int64
	Resources             shared.Resources
	ImagePullPolicy       core.PullPolicy
	LogLevel              logging.Level
}

func (provider *Provider) GetMizuApiServerPodObject(opts *ApiServerOptions, mountVolumeClaim bool, volumeClaimName string, createAuthContainer bool) (*core.Pod, error) {
	var marshaledSyncEntriesConfig []byte
	if opts.SyncEntriesConfig != nil {
		var err error
		if marshaledSyncEntriesConfig, err = json.Marshal(opts.SyncEntriesConfig); err != nil {
			return nil, err
		}
	}

	configMapVolume := &core.ConfigMapVolumeSource{}
	configMapVolume.Name = ConfigMapName

	cpuLimit, err := resource.ParseQuantity(opts.Resources.CpuLimit)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("invalid cpu limit for %s container", opts.PodName))
	}
	memLimit, err := resource.ParseQuantity(opts.Resources.MemoryLimit)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("invalid memory limit for %s container", opts.PodName))
	}
	cpuRequests, err := resource.ParseQuantity(opts.Resources.CpuRequests)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("invalid cpu request for %s container", opts.PodName))
	}
	memRequests, err := resource.ParseQuantity(opts.Resources.MemoryRequests)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("invalid memory request for %s container", opts.PodName))
	}

	command := []string{"./mizuagent", "--api-server"}
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
					Name:  shared.SyncEntriesConfigEnvVar,
					Value: string(marshaledSyncEntriesConfig),
				},
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
				Handler: core.Handler{
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
				Handler: core.Handler{
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
				Handler: core.Handler{
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

func (provider *Provider) CreateDeployment(ctx context.Context, namespace string, deploymentName string, podSpec *core.Pod) (*v1.Deployment, error) {
	if _, keyExists := podSpec.ObjectMeta.Labels["app"]; keyExists == false {
		return nil, errors.New("pod spec must contain 'app' label")
	}
	podTemplate := &core.PodTemplateSpec{
		ObjectMeta: podSpec.ObjectMeta,
		Spec:       podSpec.Spec,
	}
	deployment := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: deploymentName,
			Labels: map[string]string{
				LabelManagedBy: provider.managedBy,
				LabelCreatedBy: provider.createdBy,
			},
		},
		Spec: v1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": podSpec.ObjectMeta.Labels["app"]},
			},
			Template: *podTemplate,
			Strategy: v1.DeploymentStrategy{},
		},
	}
	return provider.clientSet.AppsV1().Deployments(namespace).Create(ctx, deployment, metav1.CreateOptions{})
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
				Resource: resource,
				Verb: verb,
				Group: group,
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

func (provider *Provider) DoesPersistentVolumeClaimExist(ctx context.Context, namespace string, name string) (bool, error) {
	persistentVolumeClaimResource, err := provider.clientSet.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, name, metav1.GetOptions{})
	return provider.doesResourceExist(persistentVolumeClaimResource, err)
}

func (provider *Provider) DoesDeploymentExist(ctx context.Context, namespace string, name string) (bool, error) {
	deploymentResource, err := provider.clientSet.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	return provider.doesResourceExist(deploymentResource, err)
}

func (provider *Provider) DoesPodExist(ctx context.Context, namespace string, name string) (bool, error) {
	podResource, err := provider.clientSet.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	return provider.doesResourceExist(podResource, err)
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

func (provider *Provider) CreateDaemonsetRBAC(ctx context.Context, namespace string, serviceAccountName string, roleName string, roleBindingName string, version string) error {
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
				APIGroups: []string{"apps"},
				Resources: []string{"daemonsets"},
				Verbs:     []string{"patch", "get", "list", "create", "delete"},
			},
			{
				APIGroups: []string{"events.k8s.io"},
				Resources: []string{"events"},
				Verbs:     []string{"list", "watch"},
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
	_, err := provider.clientSet.RbacV1().Roles(namespace).Create(ctx, role, metav1.CreateOptions{})
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

func (provider *Provider) RemoveServicAccount(ctx context.Context, namespace string, name string) error {
	err := provider.clientSet.CoreV1().ServiceAccounts(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	return provider.handleRemovalError(err)
}

func (provider *Provider) RemovePod(ctx context.Context, namespace string, podName string) error {
	err := provider.clientSet.CoreV1().Pods(namespace).Delete(ctx, podName, metav1.DeleteOptions{})
	return provider.handleRemovalError(err)
}

func (provider *Provider) RemoveDeployment(ctx context.Context, namespace string, deploymentName string) error {
	err := provider.clientSet.AppsV1().Deployments(namespace).Delete(ctx, deploymentName, metav1.DeleteOptions{})
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

func (provider *Provider) RemovePersistentVolumeClaim(ctx context.Context, namespace string, volumeClaimName string) error {
	err := provider.clientSet.CoreV1().PersistentVolumeClaims(namespace).Delete(ctx, volumeClaimName, metav1.DeleteOptions{})
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
	configMapData := make(map[string]string, 0)
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

func (provider *Provider) ApplyMizuTapperDaemonSet(ctx context.Context, namespace string, daemonSetName string, podImage string, tapperPodName string, apiServerPodIp string, nodeToTappedPodMap map[string][]core.Pod, serviceAccountName string, resources shared.Resources, imagePullPolicy core.PullPolicy, mizuApiFilteringOptions api.TrafficFilteringOptions, logLevel logging.Level, serviceMesh bool) error {
	logger.Log.Debugf("Applying %d tapper daemon sets, ns: %s, daemonSetName: %s, podImage: %s, tapperPodName: %s", len(nodeToTappedPodMap), namespace, daemonSetName, podImage, tapperPodName)

	if len(nodeToTappedPodMap) == 0 {
		return fmt.Errorf("daemon set %s must tap at least 1 pod", daemonSetName)
	}

	nodeToTappedPodMapJsonStr, err := json.Marshal(nodeToTappedPodMap)
	if err != nil {
		return err
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
		mizuCmd = append(mizuCmd, "--procfs", procfsMountPath, "--servicemesh")
	}

	agentContainer := applyconfcore.Container()
	agentContainer.WithName(tapperPodName)
	agentContainer.WithImage(podImage)
	agentContainer.WithImagePullPolicy(imagePullPolicy)

	caps := applyconfcore.Capabilities().WithDrop("ALL").WithAdd("NET_RAW").WithAdd("NET_ADMIN")

	if serviceMesh {
		caps = caps.WithAdd("SYS_ADMIN")    // for reading /proc/PID/net/ns
		caps = caps.WithAdd("SYS_PTRACE")   // for setting netns to other process
		caps = caps.WithAdd("DAC_OVERRIDE") // for reading /proc/PID/environ
	}

	agentContainer.WithSecurityContext(applyconfcore.SecurityContext().WithCapabilities(caps))

	agentContainer.WithCommand(mizuCmd...)
	agentContainer.WithEnv(
		applyconfcore.EnvVar().WithName(shared.LogLevelEnvVar).WithValue(logLevel.String()),
		applyconfcore.EnvVar().WithName(shared.HostModeEnvVar).WithValue("1"),
		applyconfcore.EnvVar().WithName(shared.TappedAddressesPerNodeDictEnvVar).WithValue(string(nodeToTappedPodMapJsonStr)),
		applyconfcore.EnvVar().WithName(shared.GoGCEnvVar).WithValue("12800"),
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
		return errors.New(fmt.Sprintf("invalid cpu limit for %s container", tapperPodName))
	}
	memLimit, err := resource.ParseQuantity(resources.MemoryLimit)
	if err != nil {
		return errors.New(fmt.Sprintf("invalid memory limit for %s container", tapperPodName))
	}
	cpuRequests, err := resource.ParseQuantity(resources.CpuRequests)
	if err != nil {
		return errors.New(fmt.Sprintf("invalid cpu request for %s container", tapperPodName))
	}
	memRequests, err := resource.ParseQuantity(resources.MemoryRequests)
	if err != nil {
		return errors.New(fmt.Sprintf("invalid memory request for %s container", tapperPodName))
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

	nodeNames := make([]string, 0, len(nodeToTappedPodMap))
	for nodeName := range nodeToTappedPodMap {
		nodeNames = append(nodeNames, nodeName)
	}
	nodeSelectorRequirement := applyconfcore.NodeSelectorRequirement()
	nodeSelectorRequirement.WithKey("kubernetes.io/hostname")
	nodeSelectorRequirement.WithOperator(core.NodeSelectorOpIn)
	nodeSelectorRequirement.WithValues(nodeNames...)
	nodeSelectorTerm := applyconfcore.NodeSelectorTerm()
	nodeSelectorTerm.WithMatchExpressions(nodeSelectorRequirement)
	nodeSelector := applyconfcore.NodeSelector()
	nodeSelector.WithNodeSelectorTerms(nodeSelectorTerm)
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
	volumeMount := applyconfcore.VolumeMount().WithName(procfsVolumeName).WithMountPath(procfsMountPath).WithReadOnly(true)
	agentContainer.WithVolumeMounts(volumeMount)

	volumeName := ConfigMapName
	configMapVolume := applyconfcore.VolumeApplyConfiguration{
		Name: &volumeName,
		VolumeSourceApplyConfiguration: applyconfcore.VolumeSourceApplyConfiguration{
			ConfigMap: &applyconfcore.ConfigMapVolumeSourceApplyConfiguration{
				LocalObjectReferenceApplyConfiguration: applyconfcore.LocalObjectReferenceApplyConfiguration{
					Name: &volumeName,
				},
			},
		},
	}
	mountPath := shared.ConfigDirPath
	configMapVolumeMount := applyconfcore.VolumeMountApplyConfiguration{
		Name:      &volumeName,
		MountPath: &mountPath,
	}
	agentContainer.WithVolumeMounts(&configMapVolumeMount)

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
	podSpec.WithVolumes(&configMapVolume, procfsVolume)

	podTemplate := applyconfcore.PodTemplateSpec()
	podTemplate.WithLabels(map[string]string{
		"app":          tapperPodName,
		LabelManagedBy: provider.managedBy,
		LabelCreatedBy: provider.createdBy,
	})
	podTemplate.WithSpec(podSpec)

	labelSelector := applyconfmeta.LabelSelector()
	labelSelector.WithMatchLabels(map[string]string{"app": tapperPodName})

	daemonSet := applyconfapp.DaemonSet(daemonSetName, namespace)
	daemonSet.
		WithLabels(map[string]string{
			LabelManagedBy: provider.managedBy,
			LabelCreatedBy: provider.createdBy,
		}).
		WithSpec(applyconfapp.DaemonSetSpec().WithSelector(labelSelector).WithTemplate(podTemplate))

	_, err = provider.clientSet.AppsV1().DaemonSets(namespace).Apply(ctx, daemonSet, metav1.ApplyOptions{FieldManager: fieldManagerName})
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
		if isPodRunning(&pod) {
			matchingPods = append(matchingPods, pod)
		}
	}
	return matchingPods, nil
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

func (provider *Provider) IsDefaultStorageProviderAvailable(ctx context.Context) (bool, error) {
	storageClassList, err := provider.clientSet.StorageV1().StorageClasses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return false, err
	}
	for _, storageClass := range storageClassList.Items {
		if storageClass.Annotations["storageclass.kubernetes.io/is-default-class"] == "true" {
			return true, nil
		}
	}
	return false, nil
}

func (provider *Provider) CreatePersistentVolumeClaim(ctx context.Context, namespace string, volumeClaimName string, sizeLimitBytes int64) (*core.PersistentVolumeClaim, error) {
	sizeLimitQuantity := resource.NewQuantity(sizeLimitBytes, resource.DecimalSI)
	volumeClaim := &core.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: volumeClaimName,
			Labels: map[string]string{
				LabelManagedBy: provider.managedBy,
				LabelCreatedBy: provider.createdBy,
			},
		},
		Spec: core.PersistentVolumeClaimSpec{
			AccessModes: []core.PersistentVolumeAccessMode{core.ReadWriteOnce},
			Resources: core.ResourceRequirements{
				Limits: core.ResourceList{
					core.ResourceStorage: *sizeLimitQuantity,
				},
				Requests: core.ResourceList{
					core.ResourceStorage: *sizeLimitQuantity,
				},
			},
		},
	}

	return provider.clientSet.CoreV1().PersistentVolumeClaims(namespace).Create(ctx, volumeClaim, metav1.CreateOptions{})
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

func getClientSet(config *restclient.Config) (*kubernetes.Clientset, error) {
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

func loadKubernetesConfiguration(kubeConfigPath string) clientcmd.ClientConfig {
	logger.Log.Debugf("Using kube config %s", kubeConfigPath)
	configPathList := filepath.SplitList(kubeConfigPath)
	configLoadingRules := &clientcmd.ClientConfigLoadingRules{}
	if len(configPathList) <= 1 {
		configLoadingRules.ExplicitPath = kubeConfigPath
	} else {
		configLoadingRules.Precedence = configPathList
	}
	contextName := ""
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		configLoadingRules,
		&clientcmd.ConfigOverrides{
			CurrentContext: contextName,
		},
	)
}

func isPodRunning(pod *core.Pod) bool {
	return pod.Status.Phase == core.PodRunning
}
