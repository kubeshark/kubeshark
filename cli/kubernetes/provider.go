package kubernetes

import (
	_ "bytes"
	"context"
	"errors"
	"fmt"
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	_ "k8s.io/client-go/plugin/pkg/client/auth/openstack"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	_ "k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
	"strings"
)

type Provider struct {
	clientSet        *kubernetes.Clientset
	kubernetesConfig clientcmd.ClientConfig
	clientConfig     restclient.Config
	Namespace        string
}

const (
	serviceAccountName     = "mizu-service-account"
)

func NewProvider(kubeConfigPath string, overrideNamespace string) *Provider {
	kubernetesConfig := loadKubernetesConfiguration(kubeConfigPath)
	restClientConfig, err := kubernetesConfig.ClientConfig()
	if err != nil {
		panic(err.Error())
	}
	clientSet := getClientSet(restClientConfig)

	var namespace string
	if len(overrideNamespace) > 0 {
		namespace = overrideNamespace
	} else {
		configuredNamespace, _, err := kubernetesConfig.Namespace()
		if err != nil {
			panic(err)
		}
		namespace = configuredNamespace
	}

	return &Provider{
		clientSet:        clientSet,
		kubernetesConfig: kubernetesConfig,
		clientConfig:     *restClientConfig,
		Namespace:        namespace,
	}
}

func (provider *Provider) GetPodWatcher(ctx context.Context, namespace string) watch.Interface {
	watcher, err := provider.clientSet.CoreV1().Pods(namespace).Watch(ctx, metav1.ListOptions{Watch: true})
	if err != nil {
		panic(err.Error())
	}
	return watcher
}

func (provider *Provider) GetPods(ctx context.Context, namespace string) {
	pods, err := provider.clientSet.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("There are %d pods in Namespace %s\n", len(pods.Items), namespace)
}

func (provider *Provider) CreateMizuPod(ctx context.Context, namespace string, podName string, podImage string, tappedPodNamespace string, tappedPodName string) (*core.Pod, error) {
	tappedPod, err := provider.clientSet.CoreV1().Pods(tappedPodNamespace).Get(ctx, tappedPodName, metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}

	podIps := make([]string, len(tappedPod.Status.PodIPs))
	for ii, podIp := range tappedPod.Status.PodIPs {
		podIps[ii] = podIp.IP
	}
	podIpsString := strings.Join(podIps, ",")

	privileged := true
	pod := &core.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: namespace,
		},
		Spec: core.PodSpec{
			HostNetwork: true,  // very important to make passive tapper see traffic
			Containers: []core.Container{
				{
					Name:            podName,
					Image:           podImage,
					ImagePullPolicy: core.PullAlways,
					SecurityContext: &core.SecurityContext{
						Privileged: &privileged, // must be privileged to get node level traffic
					},
					Env: []core.EnvVar{
						{
							Name: "HOST_MODE",
							Value: "1",
						},
						{
							Name: "TAPPED_ADDRESSES",
							Value: podIpsString,
						},
					},
				},
			},
			ServiceAccountName: serviceAccountName,
			TerminationGracePeriodSeconds: new(int64),
			NodeSelector: map[string]string{"kubernetes.io/hostname": tappedPod.Spec.NodeName},
		},
	}
	return provider.clientSet.CoreV1().Pods(namespace).Create(ctx, pod, metav1.CreateOptions{})
}

func (provider *Provider) DoesMizuRBACExist(ctx context.Context, namespace string) (bool, error){
	serviceAccount, err := provider.clientSet.CoreV1().ServiceAccounts(namespace).Get(ctx, serviceAccountName, metav1.GetOptions{})

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
	return serviceAccount != nil, nil
}

func (provider *Provider) CreateMizuRBAC(ctx context.Context, namespace string ,version string) error {
	clusterRoleName := "mizu-cluster-role"

	serviceAccount := &core.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceAccountName,
			Namespace: namespace,
			Labels:    map[string]string{"mizu-cli-version": version},
		},
	}
	clusterRole := &rbac.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterRoleName,
			Labels: map[string]string{"mizu-cli-version": version},
		},
		Rules: []rbac.PolicyRule{
			{
				APIGroups: []string {"", "extensions", "apps"},
				Resources: []string {"pods", "services", "endpoints"},
				Verbs: []string {"list", "get", "watch"},
			},
		},
	}
	clusterRoleBinding := &rbac.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "mizu-cluster-role-binding",
			Labels: map[string]string{"mizu-cli-version": version},
		},
		RoleRef: rbac.RoleRef{
			Name: clusterRoleName,
			Kind: "ClusterRole",
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
	if err != nil {
		return err
	}
	_, err = provider.clientSet.RbacV1().ClusterRoles().Create(ctx, clusterRole, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	_, err = provider.clientSet.RbacV1().ClusterRoleBindings().Create(ctx, clusterRoleBinding, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (provider *Provider) RemovePod(ctx context.Context, namespace string, podName string) {
	provider.clientSet.CoreV1().Pods(namespace).Delete(ctx, podName, metav1.DeleteOptions{})
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
		home := homedir.HomeDir()
		kubeConfigPath = filepath.Join(home, ".kube", "config")
	}

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
