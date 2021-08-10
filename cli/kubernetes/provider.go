package kubernetes

import (
	"bytes"
	_ "bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/up9inc/mizu/cli/logger"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/up9inc/mizu/cli/mizu"
	"github.com/up9inc/mizu/shared"
	"io"
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/watch"
	applyconfapp "k8s.io/client-go/applyconfigurations/apps/v1"
	applyconfcore "k8s.io/client-go/applyconfigurations/core/v1"
	applyconfmeta "k8s.io/client-go/applyconfigurations/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	_ "k8s.io/client-go/plugin/pkg/client/auth/openstack"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	_ "k8s.io/client-go/tools/portforward"
	watchtools "k8s.io/client-go/tools/watch"
	"k8s.io/client-go/util/homedir"
)

type Provider struct {
	clientSet        *kubernetes.Clientset
	kubernetesConfig clientcmd.ClientConfig
	clientConfig     restclient.Config
	Namespace        string
}

const (
	fieldManagerName = "mizu-manager"
)

func NewProvider(kubeConfigPath string) (*Provider, error) {
	kubernetesConfig := loadKubernetesConfiguration(kubeConfigPath)
	restClientConfig, err := kubernetesConfig.ClientConfig()
	if err != nil {
		if clientcmd.IsEmptyConfig(err) {
			return nil, fmt.Errorf("Couldn't find the kube config file, or file is empty. Try adding '--kube-config=<path to kube config file>'\n")
		}
		if clientcmd.IsConfigurationInvalid(err) {
			return nil, fmt.Errorf("Invalid kube config file. Try using a different config with '--kube-config=<path to kube config file>'\n")
		}
	}
	clientSet := getClientSet(restClientConfig)

	return &Provider{
		clientSet:        clientSet,
		kubernetesConfig: kubernetesConfig,
		clientConfig:     *restClientConfig,
	}, nil
}

func (provider *Provider) CurrentNamespace() string {
	ns, _, _ := provider.kubernetesConfig.Namespace()
	return ns
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

func (provider *Provider) GetPodWatcher(ctx context.Context, namespace string) watch.Interface {
	watcher, err := provider.clientSet.CoreV1().Pods(namespace).Watch(ctx, metav1.ListOptions{Watch: true})
	if err != nil {
		panic(err.Error())
	}
	return watcher
}

func (provider *Provider) CreateNamespace(ctx context.Context, name string) (*core.Namespace, error) {
	namespaceSpec := &core.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	return provider.clientSet.CoreV1().Namespaces().Create(ctx, namespaceSpec, metav1.CreateOptions{})
}

type ApiServerOptions struct {
	Namespace               string
	PodName                 string
	PodImage                string
	ServiceAccountName      string
	IsNamespaceRestricted   bool
	MizuApiFilteringOptions *shared.TrafficFilteringOptions
	MaxEntriesDBSizeBytes   int64
}

func (provider *Provider) CreateMizuApiServerPod(ctx context.Context, opts *ApiServerOptions) (*core.Pod, error) {
	marshaledFilteringOptions, err := json.Marshal(opts.MizuApiFilteringOptions)
	if err != nil {
		return nil, err
	}
	configMapVolumeName := &core.ConfigMapVolumeSource{}
	configMapVolumeName.Name = mizu.ConfigMapName
	configMapOptional := true
	configMapVolumeName.Optional = &configMapOptional

	cpuLimit, err := resource.ParseQuantity("750m")
	if err != nil {
		return nil, errors.New(fmt.Sprintf("invalid cpu limit for %s container", opts.PodName))
	}
	memLimit, err := resource.ParseQuantity("512Mi")
	if err != nil {
		return nil, errors.New(fmt.Sprintf("invalid memory limit for %s container", opts.PodName))
	}
	cpuRequests, err := resource.ParseQuantity("50m")
	if err != nil {
		return nil, errors.New(fmt.Sprintf("invalid cpu request for %s container", opts.PodName))
	}
	memRequests, err := resource.ParseQuantity("50Mi")
	if err != nil {
		return nil, errors.New(fmt.Sprintf("invalid memory request for %s container", opts.PodName))
	}

	command := []string{"./mizuagent", "--api-server"}
	if opts.IsNamespaceRestricted {
		command = append(command, "--namespace", opts.Namespace)
	}

	pod := &core.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      opts.PodName,
			Namespace: opts.Namespace,
			Labels:    map[string]string{"app": opts.PodName},
		},
		Spec: core.PodSpec{
			Containers: []core.Container{
				{
					Name:            opts.PodName,
					Image:           opts.PodImage,
					ImagePullPolicy: core.PullAlways,
					VolumeMounts: []core.VolumeMount{
						{
							Name:      mizu.ConfigMapName,
							MountPath: shared.RulePolicyPath,
						},
					},
					Command: command,
					Env: []core.EnvVar{
						{
							Name:  shared.HostModeEnvVar,
							Value: "1",
						},
						{
							Name:  shared.MizuFilteringOptionsEnvVar,
							Value: string(marshaledFilteringOptions),
						},
						{
							Name:  shared.MaxEntriesDBSizeBytesEnvVar,
							Value: strconv.FormatInt(opts.MaxEntriesDBSizeBytes, 10),
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
			},
			Volumes: []core.Volume{
				{
					Name: mizu.ConfigMapName,
					VolumeSource: core.VolumeSource{
						ConfigMap: configMapVolumeName,
					},
				},
			},
			DNSPolicy:                     core.DNSClusterFirstWithHostNet,
			TerminationGracePeriodSeconds: new(int64),
		},
	}
	//define the service account only when it exists to prevent pod crash
	if opts.ServiceAccountName != "" {
		pod.Spec.ServiceAccountName = opts.ServiceAccountName
	}
	return provider.clientSet.CoreV1().Pods(opts.Namespace).Create(ctx, pod, metav1.CreateOptions{})
}

func (provider *Provider) CreateService(ctx context.Context, namespace string, serviceName string, appLabelValue string) (*core.Service, error) {
	service := core.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: namespace,
		},
		Spec: core.ServiceSpec{
			Ports:    []core.ServicePort{{TargetPort: intstr.FromInt(8899), Port: 80}},
			Type:     core.ServiceTypeClusterIP,
			Selector: map[string]string{"app": appLabelValue},
		},
	}
	return provider.clientSet.CoreV1().Services(namespace).Create(ctx, &service, metav1.CreateOptions{})
}

func (provider *Provider) DoesServiceAccountExist(ctx context.Context, namespace string, serviceAccountName string) (bool, error) {
	serviceAccount, err := provider.clientSet.CoreV1().ServiceAccounts(namespace).Get(ctx, serviceAccountName, metav1.GetOptions{})
	return provider.doesResourceExist(serviceAccount, err)
}

func (provider *Provider) DoesConfigMapExist(ctx context.Context, namespace string, name string) (bool, error) {
	resource, err := provider.clientSet.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
	return provider.doesResourceExist(resource, err)
}

func (provider *Provider) DoesServicesExist(ctx context.Context, namespace string, name string) (bool, error) {
	resource, err := provider.clientSet.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	return provider.doesResourceExist(resource, err)
}

func (provider *Provider) DoesNamespaceExist(ctx context.Context, name string) (bool, error) {
	resource, err := provider.clientSet.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	return provider.doesResourceExist(resource, err)
}

func (provider *Provider) DoesClusterRoleExist(ctx context.Context, name string) (bool, error) {
	resource, err := provider.clientSet.RbacV1().ClusterRoles().Get(ctx, name, metav1.GetOptions{})
	return provider.doesResourceExist(resource, err)
}

func (provider *Provider) DoesClusterRoleBindingExist(ctx context.Context, name string) (bool, error) {
	resource, err := provider.clientSet.RbacV1().ClusterRoleBindings().Get(ctx, name, metav1.GetOptions{})
	return provider.doesResourceExist(resource, err)
}

func (provider *Provider) DoesRoleExist(ctx context.Context, namespace string, name string) (bool, error) {
	resource, err := provider.clientSet.RbacV1().Roles(namespace).Get(ctx, name, metav1.GetOptions{})
	return provider.doesResourceExist(resource, err)
}

func (provider *Provider) DoesRoleBindingExist(ctx context.Context, namespace string, name string) (bool, error) {
	resource, err := provider.clientSet.RbacV1().RoleBindings(namespace).Get(ctx, name, metav1.GetOptions{})
	return provider.doesResourceExist(resource, err)
}

func (provider *Provider) DoesPodExist(ctx context.Context, namespace string, name string) (bool, error) {
	resource, err := provider.clientSet.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	return provider.doesResourceExist(resource, err)
}

func (provider *Provider) DoesDaemonSetExist(ctx context.Context, namespace string, name string) (bool, error) {
	resource, err := provider.clientSet.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
	return provider.doesResourceExist(resource, err)
}

func (provider *Provider) doesResourceExist(resource interface{}, err error) (bool, error) {
	var statusError *k8serrors.StatusError
	if errors.As(err, &statusError) {
		// expected behavior when resource does not exist
		if statusError.ErrStatus.Reason == metav1.StatusReasonNotFound {
			return false, nil
		}
	}
	if err != nil {
		return false, err
	}
	return resource != nil, nil
}

func (provider *Provider) CreateMizuRBAC(ctx context.Context, namespace string, serviceAccountName string, clusterRoleName string, clusterRoleBindingName string, version string) error {
	serviceAccount := &core.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceAccountName,
			Namespace: namespace,
			Labels:    map[string]string{"mizu-cli-version": version},
		},
	}
	clusterRole := &rbac.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:   clusterRoleName,
			Labels: map[string]string{"mizu-cli-version": version},
		},
		Rules: []rbac.PolicyRule{
			{
				APIGroups: []string{"", "extensions", "apps"},
				Resources: []string{"pods", "services", "endpoints"},
				Verbs:     []string{"list", "get", "watch"},
			},
		},
	}
	clusterRoleBinding := &rbac.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   clusterRoleBindingName,
			Labels: map[string]string{"mizu-cli-version": version},
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
			Name:      serviceAccountName,
			Namespace: namespace,
			Labels:    map[string]string{"mizu-cli-version": version},
		},
	}
	role := &rbac.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:   roleName,
			Labels: map[string]string{"mizu-cli-version": version},
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
			Name:   roleBindingName,
			Labels: map[string]string{"mizu-cli-version": version},
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
	if isFound, err := provider.DoesNamespaceExist(ctx, name); err != nil {
		return err
	} else if !isFound {
		return nil
	}

	return provider.clientSet.CoreV1().Namespaces().Delete(ctx, name, metav1.DeleteOptions{})
}

func (provider *Provider) RemoveNonNamespacedResources(ctx context.Context, clusterRoleName string, clusterRoleBindingName string) error {
	if err := provider.RemoveClusterRole(ctx, clusterRoleName); err != nil {
		return err
	}

	if err := provider.RemoveClusterRoleBinding(ctx, clusterRoleBindingName); err != nil {
		return err
	}

	return nil
}

func (provider *Provider) RemoveClusterRole(ctx context.Context, name string) error {
	if isFound, err := provider.DoesClusterRoleExist(ctx, name); err != nil {
		return err
	} else if !isFound {
		return nil
	}

	return provider.clientSet.RbacV1().ClusterRoles().Delete(ctx, name, metav1.DeleteOptions{})
}

func (provider *Provider) RemoveClusterRoleBinding(ctx context.Context, name string) error {
	if isFound, err := provider.DoesClusterRoleBindingExist(ctx, name); err != nil {
		return err
	} else if !isFound {
		return nil
	}

	return provider.clientSet.RbacV1().ClusterRoleBindings().Delete(ctx, name, metav1.DeleteOptions{})
}

func (provider *Provider) RemoveRoleBinding(ctx context.Context, namespace string, name string) error {
	if isFound, err := provider.DoesRoleBindingExist(ctx, namespace, name); err != nil {
		return err
	} else if !isFound {
		return nil
	}

	return provider.clientSet.RbacV1().RoleBindings(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (provider *Provider) RemoveRole(ctx context.Context, namespace string, name string) error {
	if isFound, err := provider.DoesRoleExist(ctx, namespace, name); err != nil {
		return err
	} else if !isFound {
		return nil
	}

	return provider.clientSet.RbacV1().Roles(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (provider *Provider) RemoveServicAccount(ctx context.Context, namespace string, name string) error {
	if isFound, err := provider.DoesServiceAccountExist(ctx, namespace, name); err != nil {
		return err
	} else if !isFound {
		return nil
	}

	return provider.clientSet.CoreV1().ServiceAccounts(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (provider *Provider) RemovePod(ctx context.Context, namespace string, podName string) error {
	if isFound, err := provider.DoesPodExist(ctx, namespace, podName); err != nil {
		return err
	} else if !isFound {
		return nil
	}

	return provider.clientSet.CoreV1().Pods(namespace).Delete(ctx, podName, metav1.DeleteOptions{})
}

func (provider *Provider) RemoveConfigMap(ctx context.Context, namespace string, configMapName string) error {
	if isFound, err := provider.DoesConfigMapExist(ctx, namespace, configMapName); err != nil {
		return err
	} else if !isFound {
		return nil
	}

	return provider.clientSet.CoreV1().ConfigMaps(namespace).Delete(ctx, configMapName, metav1.DeleteOptions{})
}

func (provider *Provider) RemoveService(ctx context.Context, namespace string, serviceName string) error {
	if isFound, err := provider.DoesServicesExist(ctx, namespace, serviceName); err != nil {
		return err
	} else if !isFound {
		return nil
	}

	return provider.clientSet.CoreV1().Services(namespace).Delete(ctx, serviceName, metav1.DeleteOptions{})
}

func (provider *Provider) RemoveDaemonSet(ctx context.Context, namespace string, daemonSetName string) error {
	if isFound, err := provider.DoesDaemonSetExist(ctx, namespace, daemonSetName); err != nil {
		return err
	} else if !isFound {
		return nil
	}

	return provider.clientSet.AppsV1().DaemonSets(namespace).Delete(ctx, daemonSetName, metav1.DeleteOptions{})
}

func (provider *Provider) CreateConfigMap(ctx context.Context, namespace string, configMapName string, data string) error {
	if data == "" {
		return nil
	}

	configMapData := make(map[string]string, 0)
	configMapData[shared.RulePolicyFileName] = data
	configMap := &core.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: namespace,
		},
		Data: configMapData,
	}
	if _, err := provider.clientSet.CoreV1().ConfigMaps(namespace).Create(ctx, configMap, metav1.CreateOptions{}); err != nil {
		return err
	}
	return nil
}

func (provider *Provider) ApplyMizuTapperDaemonSet(ctx context.Context, namespace string, daemonSetName string, podImage string, tapperPodName string, apiServerPodIp string, nodeToTappedPodIPMap map[string][]string, serviceAccountName string, tapOutgoing bool) error {
	logger.Log.Debugf("Applying %d tapper deamonsets, ns: %s, daemonSetName: %s, podImage: %s, tapperPodName: %s", len(nodeToTappedPodIPMap), namespace, daemonSetName, podImage, tapperPodName)

	if len(nodeToTappedPodIPMap) == 0 {
		return fmt.Errorf("Daemon set %s must tap at least 1 pod", daemonSetName)
	}

	nodeToTappedPodIPMapJsonStr, err := json.Marshal(nodeToTappedPodIPMap)
	if err != nil {
		return err
	}

	mizuCmd := []string{
		"./mizuagent",
		"-i", "any",
		"--tap",
		"--api-server-address", fmt.Sprintf("ws://%s/wsTapper", apiServerPodIp),
	}
	if tapOutgoing {
		mizuCmd = append(mizuCmd, "--anydirection")
	}

	agentContainer := applyconfcore.Container()
	agentContainer.WithName(tapperPodName)
	agentContainer.WithImage(podImage)
	agentContainer.WithImagePullPolicy(core.PullAlways)
	agentContainer.WithSecurityContext(applyconfcore.SecurityContext().WithPrivileged(true))
	agentContainer.WithCommand(mizuCmd...)
	agentContainer.WithEnv(
		applyconfcore.EnvVar().WithName(shared.HostModeEnvVar).WithValue("1"),
		applyconfcore.EnvVar().WithName(shared.TappedAddressesPerNodeDictEnvVar).WithValue(string(nodeToTappedPodIPMapJsonStr)),
	)
	agentContainer.WithEnv(
		applyconfcore.EnvVar().WithName(shared.NodeNameEnvVar).WithValueFrom(
			applyconfcore.EnvVarSource().WithFieldRef(
				applyconfcore.ObjectFieldSelector().WithAPIVersion("v1").WithFieldPath("spec.nodeName"),
			),
		),
	)
	cpuLimit, err := resource.ParseQuantity("500m")
	if err != nil {
		return errors.New(fmt.Sprintf("invalid cpu limit for %s container", tapperPodName))
	}
	memLimit, err := resource.ParseQuantity("1Gi")
	if err != nil {
		return errors.New(fmt.Sprintf("invalid memory limit for %s container", tapperPodName))
	}
	cpuRequests, err := resource.ParseQuantity("50m")
	if err != nil {
		return errors.New(fmt.Sprintf("invalid cpu request for %s container", tapperPodName))
	}
	memRequests, err := resource.ParseQuantity("50Mi")
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

	nodeNames := make([]string, 0, len(nodeToTappedPodIPMap))
	for nodeName := range nodeToTappedPodIPMap {
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

	podTemplate := applyconfcore.PodTemplateSpec()
	podTemplate.WithLabels(map[string]string{"app": tapperPodName})
	podTemplate.WithSpec(podSpec)

	labelSelector := applyconfmeta.LabelSelector()
	labelSelector.WithMatchLabels(map[string]string{"app": tapperPodName})

	daemonSet := applyconfapp.DaemonSet(daemonSetName, namespace)
	daemonSet.WithSpec(applyconfapp.DaemonSetSpec().WithSelector(labelSelector).WithTemplate(podTemplate))

	_, err = provider.clientSet.AppsV1().DaemonSets(namespace).Apply(ctx, daemonSet, metav1.ApplyOptions{FieldManager: fieldManagerName})
	return err
}

func (provider *Provider) ListAllPodsMatchingRegex(ctx context.Context, regex *regexp.Regexp, namespaces []string) ([]core.Pod, error) {
	var pods []core.Pod
	for _, namespace := range namespaces {
		namespacePods, err := provider.clientSet.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
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

func (provider *Provider) GetPodLogs(namespace string, podName string, ctx context.Context) (string, error) {
	podLogOpts := core.PodLogOptions{}
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

func getClientSet(config *restclient.Config) *kubernetes.Clientset {
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return clientSet
}

func loadKubernetesConfiguration(kubeConfigPath string) clientcmd.ClientConfig {
	if kubeConfigPath == "" {
		kubeConfigPath = os.Getenv("KUBECONFIG")
	}

	if kubeConfigPath == "" {
		home := homedir.HomeDir()
		kubeConfigPath = filepath.Join(home, ".kube", "config")
	}

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
