package resources

import (
	"context"
	"fmt"

	"github.com/op/go-logging"
	"github.com/up9inc/mizu/cli/errormessage"
	"github.com/up9inc/mizu/cli/mizu"
	"github.com/up9inc/mizu/cli/uiUtils"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/kubernetes"
	"github.com/up9inc/mizu/shared/logger"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func CreateTapMizuResources(ctx context.Context, kubernetesProvider *kubernetes.Provider, serializedValidationRules string, serializedContract string, serializedMizuConfig string, isNsRestrictedMode bool, mizuResourcesNamespace string, agentImage string, syncEntriesConfig *shared.SyncEntriesConfig, maxEntriesDBSizeBytes int64, apiServerResources shared.Resources, imagePullPolicy core.PullPolicy, logLevel logging.Level) (bool, error) {
	if !isNsRestrictedMode {
		if err := createMizuNamespace(ctx, kubernetesProvider, mizuResourcesNamespace); err != nil {
			return false, err
		}
	}

	if err := createMizuConfigmap(ctx, kubernetesProvider, serializedValidationRules, serializedContract, serializedMizuConfig, mizuResourcesNamespace); err != nil {
		logger.Log.Warningf(uiUtils.Warning, fmt.Sprintf("Failed to create resources required for policy validation. Mizu will not validate policy rules. error: %v", errormessage.FormatError(err)))
	}

	mizuServiceAccountExists, err := createRBACIfNecessary(ctx, kubernetesProvider, isNsRestrictedMode, mizuResourcesNamespace, []string{"pods", "services", "endpoints"})
	if err != nil {
		logger.Log.Warningf(uiUtils.Warning, fmt.Sprintf("Failed to ensure the resources required for IP resolving. Mizu will not resolve target IPs to names. error: %v", errormessage.FormatError(err)))
	}

	var serviceAccountName string
	if mizuServiceAccountExists {
		serviceAccountName = kubernetes.ServiceAccountName
	} else {
		serviceAccountName = ""
	}

	opts := &kubernetes.ApiServerOptions{
		Namespace:             mizuResourcesNamespace,
		PodName:               kubernetes.ApiServerPodName,
		PodImage:              agentImage,
		KratosImage:           "",
		KetoImage:             "",
		ServiceAccountName:    serviceAccountName,
		IsNamespaceRestricted: isNsRestrictedMode,
		SyncEntriesConfig:     syncEntriesConfig,
		MaxEntriesDBSizeBytes: maxEntriesDBSizeBytes,
		Resources:             apiServerResources,
		ImagePullPolicy:       imagePullPolicy,
		LogLevel:              logLevel,
	}

	if err := createMizuApiServerPod(ctx, kubernetesProvider, opts); err != nil {
		return mizuServiceAccountExists, err
	}

	_, err = kubernetesProvider.CreateService(ctx, mizuResourcesNamespace, kubernetes.ApiServerPodName, kubernetes.ApiServerPodName)
	if err != nil {
		return mizuServiceAccountExists, err
	}

	logger.Log.Debugf("Successfully created service: %s", kubernetes.ApiServerPodName)

	return mizuServiceAccountExists, nil
}

func CreateInstallMizuResources(ctx context.Context, kubernetesProvider *kubernetes.Provider, serializedValidationRules string, serializedContract string, serializedMizuConfig string, isNsRestrictedMode bool, mizuResourcesNamespace string, agentImage string, kratosImage string, ketoImage string, syncEntriesConfig *shared.SyncEntriesConfig, maxEntriesDBSizeBytes int64, apiServerResources shared.Resources, imagePullPolicy core.PullPolicy, logLevel logging.Level, noPersistentVolumeClaim bool) error {
	if err := createMizuNamespace(ctx, kubernetesProvider, mizuResourcesNamespace); err != nil {
		return err
	}
	logger.Log.Infof("namespace/%v created", mizuResourcesNamespace)

	if err := createMizuConfigmap(ctx, kubernetesProvider, serializedValidationRules, serializedContract, serializedMizuConfig, mizuResourcesNamespace); err != nil {
		return err
	}
	logger.Log.Infof("configmap/%v created", kubernetes.ConfigMapName)

	_, err := createRBACIfNecessary(ctx, kubernetesProvider, isNsRestrictedMode, mizuResourcesNamespace, []string{"pods", "services", "endpoints", "namespaces"})
	if err != nil {
		return err
	}
	logger.Log.Infof("serviceaccount/%v created", kubernetes.ServiceAccountName)
	logger.Log.Infof("clusterrole.rbac.authorization.k8s.io/%v created", kubernetes.ClusterRoleName)
	logger.Log.Infof("clusterrolebinding.rbac.authorization.k8s.io/%v created", kubernetes.ClusterRoleBindingName)

	if err := kubernetesProvider.CreateDaemonsetRBAC(ctx, mizuResourcesNamespace, kubernetes.ServiceAccountName, kubernetes.DaemonRoleName, kubernetes.DaemonRoleBindingName, mizu.RBACVersion); err != nil {
		return err
	}
	logger.Log.Infof("role.rbac.authorization.k8s.io/%v created", kubernetes.DaemonRoleName)
	logger.Log.Infof("rolebinding.rbac.authorization.k8s.io/%v created", kubernetes.DaemonRoleBindingName)

	serviceAccountName := kubernetes.ServiceAccountName
	opts := &kubernetes.ApiServerOptions{
		Namespace:             mizuResourcesNamespace,
		PodName:               kubernetes.ApiServerPodName,
		PodImage:              agentImage,
		KratosImage:           kratosImage,
		KetoImage:             ketoImage,
		ServiceAccountName:    serviceAccountName,
		IsNamespaceRestricted: isNsRestrictedMode,
		SyncEntriesConfig:     syncEntriesConfig,
		MaxEntriesDBSizeBytes: maxEntriesDBSizeBytes,
		Resources:             apiServerResources,
		ImagePullPolicy:       imagePullPolicy,
		LogLevel:              logLevel,
	}

	if err := createMizuApiServerDeployment(ctx, kubernetesProvider, opts, noPersistentVolumeClaim); err != nil {
		return err
	}
	logger.Log.Infof("deployment.apps/%v created", kubernetes.ApiServerPodName)

	_, err = kubernetesProvider.CreateService(ctx, mizuResourcesNamespace, kubernetes.ApiServerPodName, kubernetes.ApiServerPodName)
	if err != nil {
		return err
	}
	logger.Log.Infof("service/%v created", kubernetes.ApiServerPodName)

	return nil
}

func createMizuNamespace(ctx context.Context, kubernetesProvider *kubernetes.Provider, mizuResourcesNamespace string) error {
	_, err := kubernetesProvider.CreateNamespace(ctx, mizuResourcesNamespace)
	return err
}

func createMizuConfigmap(ctx context.Context, kubernetesProvider *kubernetes.Provider, serializedValidationRules string, serializedContract string, serializedMizuConfig string, mizuResourcesNamespace string) error {
	err := kubernetesProvider.CreateConfigMap(ctx, mizuResourcesNamespace, kubernetes.ConfigMapName, serializedValidationRules, serializedContract, serializedMizuConfig)
	return err
}

func createRBACIfNecessary(ctx context.Context, kubernetesProvider *kubernetes.Provider, isNsRestrictedMode bool, mizuResourcesNamespace string, resources []string) (bool, error) {
	if !isNsRestrictedMode {
		if err := kubernetesProvider.CreateMizuRBAC(ctx, mizuResourcesNamespace, kubernetes.ServiceAccountName, kubernetes.ClusterRoleName, kubernetes.ClusterRoleBindingName, mizu.RBACVersion, resources); err != nil {
			return false, err
		}
	} else {
		if err := kubernetesProvider.CreateMizuRBACNamespaceRestricted(ctx, mizuResourcesNamespace, kubernetes.ServiceAccountName, kubernetes.RoleName, kubernetes.RoleBindingName, mizu.RBACVersion); err != nil {
			return false, err
		}
	}

	return true, nil
}

func createMizuApiServerDeployment(ctx context.Context, kubernetesProvider *kubernetes.Provider, opts *kubernetes.ApiServerOptions, noPersistentVolumeClaim bool) error {
	volumeClaimCreated := false
	if !noPersistentVolumeClaim {
		volumeClaimCreated = tryToCreatePersistentVolumeClaim(ctx, kubernetesProvider, opts)
	}

	pod, err := kubernetesProvider.GetMizuApiServerPodObject(opts, volumeClaimCreated, kubernetes.PersistentVolumeClaimName, true)
	if err != nil {
		return err
	}
	pod.Spec.Containers[0].LivenessProbe = &core.Probe{
		ProbeHandler: core.ProbeHandler{
			HTTPGet: &core.HTTPGetAction{
				Path: "/echo",
				Port: intstr.FromInt(shared.DefaultApiServerPort),
			},
		},
		InitialDelaySeconds: 1,
		PeriodSeconds:       10,
	}
	if _, err = kubernetesProvider.CreateDeployment(ctx, opts.Namespace, opts.PodName, pod); err != nil {
		return err
	}
	logger.Log.Debugf("Successfully created API server deployment: %s", kubernetes.ApiServerPodName)
	return nil
}

func tryToCreatePersistentVolumeClaim(ctx context.Context, kubernetesProvider *kubernetes.Provider, opts *kubernetes.ApiServerOptions) bool {
	isDefaultStorageClassAvailable, err := kubernetesProvider.IsDefaultStorageProviderAvailable(ctx)
	if err != nil {
		logger.Log.Warningf(uiUtils.Yellow, "An error occured when checking if a default storage provider exists in this cluster, this means mizu data will be lost on mizu-api-server pod restart")
		logger.Log.Debugf("error checking if default storage class exists: %v", err)
		return false
	} else if !isDefaultStorageClassAvailable {
		logger.Log.Warningf(uiUtils.Yellow, "Could not find default storage provider in this cluster, this means mizu data will be lost on mizu-api-server pod restart")
		return false
	}

	if _, err = kubernetesProvider.CreatePersistentVolumeClaim(ctx, opts.Namespace, kubernetes.PersistentVolumeClaimName, opts.MaxEntriesDBSizeBytes+mizu.InstallModePersistentVolumeSizeBufferBytes); err != nil {
		logger.Log.Warningf(uiUtils.Yellow, "An error has occured while creating a persistent volume claim for mizu, this means mizu data will be lost on mizu-api-server pod restart")
		logger.Log.Debugf("error creating persistent volume claim: %v", err)
		return false
	}

	return true
}

func createMizuApiServerPod(ctx context.Context, kubernetesProvider *kubernetes.Provider, opts *kubernetes.ApiServerOptions) error {
	pod, err := kubernetesProvider.GetMizuApiServerPodObject(opts, false, "", false)
	if err != nil {
		return err
	}
	if _, err = kubernetesProvider.CreatePod(ctx, opts.Namespace, pod); err != nil {
		return err
	}
	logger.Log.Debugf("Successfully created API server pod: %s", kubernetes.ApiServerPodName)
	return nil
}
