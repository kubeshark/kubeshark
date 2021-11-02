package cmd

import (
	"context"
	"fmt"
	"github.com/up9inc/mizu/cli/config"
	"github.com/up9inc/mizu/cli/errormessage"
	"github.com/up9inc/mizu/cli/mizu"
	"github.com/up9inc/mizu/cli/mizu/fsUtils"
	"github.com/up9inc/mizu/cli/uiUtils"
	"github.com/up9inc/mizu/shared/kubernetes"
	"github.com/up9inc/mizu/shared/logger"
	"k8s.io/apimachinery/pkg/util/wait"
)

//TODO: refactor mizu resource creation

func createMizuResources(ctx context.Context, cancel context.CancelFunc, kubernetesProvider *kubernetes.Provider, serializedValidationRules string, serializedContract string, serializedMizuConfig string) error {
	var err error
	if !config.Config.IsNsRestrictedMode() {
		if err = createMizuNamespace(ctx, kubernetesProvider); err != nil {
			return err
		}
	}

	if err := createMizuConfigmap(ctx, kubernetesProvider, serializedValidationRules, serializedContract, serializedMizuConfig); err != nil {
		logger.Log.Warningf(uiUtils.Warning, fmt.Sprintf("Failed to create resources required for policy validation. Mizu will not validate policy rules. error: %v\n", errormessage.FormatError(err)))
	}

	state.mizuServiceAccountExists, err = createRBACIfNecessary(ctx, kubernetesProvider)
	if err != nil {
		return err
	}

	var serviceAccountName string
	if state.mizuServiceAccountExists {
		serviceAccountName = kubernetes.ServiceAccountName
	} else {
		serviceAccountName = ""
	}

	opts := &kubernetes.ApiServerOptions{
		Namespace:             config.Config.MizuResourcesNamespace,
		PodName:               kubernetes.ApiServerPodName,
		PodImage:              config.Config.AgentImage,
		ServiceAccountName:    serviceAccountName,
		IsNamespaceRestricted: config.Config.IsNsRestrictedMode(),
		SyncEntriesConfig:     getSyncEntriesConfig(),
		MaxEntriesDBSizeBytes: config.Config.Tap.MaxEntriesDBSizeBytes(),
		Resources:             config.Config.Tap.ApiServerResources,
		ImagePullPolicy:       config.Config.ImagePullPolicy(),
	}

	if config.Config.Tap.DaemonMode {
		if state.mizuServiceAccountExists == false {
			defer cleanUpMizuResources(ctx, cancel, kubernetesProvider)
			logger.Log.Fatalf(uiUtils.Red, fmt.Sprintf("Failed to ensure the resources required for mizu to run in daemon mode. cannot proceed. error: %v", errormessage.FormatError(err)))
		}
		if err := createMizuApiServerDeployment(ctx, kubernetesProvider, opts); err != nil {
			return err
		}
	} else {
		if err := createMizuApiServerPod(ctx, kubernetesProvider, opts); err != nil {
			return err
		}
	}

	state.apiServerService, err = kubernetesProvider.CreateService(ctx, config.Config.MizuResourcesNamespace, kubernetes.ApiServerPodName, kubernetes.ApiServerPodName)
	if err != nil {
		return err
	}
	logger.Log.Debugf("Successfully created service: %s", kubernetes.ApiServerPodName)

	return nil
}

func createMizuConfigmap(ctx context.Context, kubernetesProvider *kubernetes.Provider, serializedValidationRules string, serializedContract string, serializedMizuConfig string) error {
	err := kubernetesProvider.CreateConfigMap(ctx, config.Config.MizuResourcesNamespace, kubernetes.ConfigMapName, serializedValidationRules, serializedContract, serializedMizuConfig)
	return err
}

func createMizuNamespace(ctx context.Context, kubernetesProvider *kubernetes.Provider) error {
	_, err := kubernetesProvider.CreateNamespace(ctx, config.Config.MizuResourcesNamespace)
	return err
}

func createMizuApiServerPod(ctx context.Context, kubernetesProvider *kubernetes.Provider, opts *kubernetes.ApiServerOptions) error {
	pod, err := kubernetesProvider.GetMizuApiServerPodObject(opts, false, "")
	if err != nil {
		return err
	}
	_, err = kubernetesProvider.CreatePod(ctx, config.Config.MizuResourcesNamespace, pod)
	if err != nil {
		return err
	}
	logger.Log.Debugf("Successfully created API server pod: %s", kubernetes.ApiServerPodName)
	return nil
}

func createMizuApiServerDeployment(ctx context.Context, kubernetesProvider *kubernetes.Provider, opts *kubernetes.ApiServerOptions) error {
	isDefaultStorageClassAvailable, err := kubernetesProvider.IsDefaultStorageProviderAvailable(ctx)
	volumeClaimCreated := false
	if err != nil {
		return err
	}
	if isDefaultStorageClassAvailable {
		_, err = kubernetesProvider.CreatePersistentVolumeClaim(ctx, config.Config.MizuResourcesNamespace, kubernetes.PersistentVolumeClaimName, config.Config.Tap.MaxEntriesDBSizeBytes() + mizu.DaemonModePersistentVolumeSizeBufferBytes)
		if err != nil {
			logger.Log.Warningf(uiUtils.Yellow, "An error has occured while creating a persistent volume claim for mizu, this will mean that mizu's data will be lost on pod restart")
			logger.Log.Debugf("error creating persistent volume claim: %v", err)
		} else {
			volumeClaimCreated = true
		}
	} else {
		logger.Log.Warningf(uiUtils.Yellow, "Could not find default volume provider in this cluster, this will mean that mizu's data will be lost on pod restart")
	}

	pod, err := kubernetesProvider.GetMizuApiServerPodObject(opts, volumeClaimCreated, kubernetes.PersistentVolumeClaimName)
	if err != nil {
		return err
	}

	_, err = kubernetesProvider.CreateDeployment(ctx, config.Config.MizuResourcesNamespace, opts.PodName, pod)
	if err != nil {
		return err
	}
	logger.Log.Debugf("Successfully created API server deployment: %s", kubernetes.ApiServerPodName)
	return nil
}


func createRBACIfNecessary(ctx context.Context, kubernetesProvider *kubernetes.Provider) (bool, error) {
	if !config.Config.IsNsRestrictedMode() {
		if err := kubernetesProvider.CreateMizuRBAC(ctx, config.Config.MizuResourcesNamespace, kubernetes.ServiceAccountName, kubernetes.ClusterRoleName, kubernetes.ClusterRoleBindingName, mizu.RBACVersion); err != nil {
			return false, err
		}
	} else {
		if err := kubernetesProvider.CreateMizuRBACNamespaceRestricted(ctx, config.Config.MizuResourcesNamespace, kubernetes.ServiceAccountName, kubernetes.RoleName, kubernetes.RoleBindingName, mizu.RBACVersion); err != nil {
			return false, err
		}
	}
	if config.Config.Tap.DaemonMode {
		if err := kubernetesProvider.CreateDaemonsetPatchRBAC(ctx, config.Config.MizuResourcesNamespace, kubernetes.ServiceAccountName, kubernetes.DaemonClusterRoleName, kubernetes.DaemonClusterRoleBindingName, mizu.RBACVersion); err != nil {
			return false, err
		}
	}
	return true, nil
}


func waitUntilNamespaceDeleted(ctx context.Context, cancel context.CancelFunc, kubernetesProvider *kubernetes.Provider) {
	// Call cancel if a terminating signal was received. Allows user to skip the wait.
	go func() {
		waitForFinish(ctx, cancel)
	}()

	if err := kubernetesProvider.WaitUtilNamespaceDeleted(ctx, config.Config.MizuResourcesNamespace); err != nil {
		switch {
		case ctx.Err() == context.Canceled:
			logger.Log.Debugf("Do nothing. User interrupted the wait")
		case err == wait.ErrWaitTimeout:
			logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Timeout while removing Namespace %s", config.Config.MizuResourcesNamespace))
		default:
			logger.Log.Errorf(uiUtils.Error, fmt.Sprintf("Error while waiting for Namespace %s to be deleted: %v", config.Config.MizuResourcesNamespace, errormessage.FormatError(err)))
		}
	}
}

func cleanUpMizuResources(ctx context.Context, cancel context.CancelFunc, kubernetesProvider *kubernetes.Provider) {
	logger.Log.Infof("\nRemoving mizu resources\n")

	var leftoverResources []string

	if config.Config.IsNsRestrictedMode() {
		leftoverResources = cleanUpRestrictedMode(ctx, kubernetesProvider)
	} else {
		leftoverResources = cleanUpNonRestrictedMode(ctx, cancel, kubernetesProvider)
	}

	if len(leftoverResources) > 0 {
		errMsg := fmt.Sprintf("Failed to remove the following resources, for more info check logs at %s:", fsUtils.GetLogFilePath())
		for _, resource := range leftoverResources {
			errMsg += "\n- " + resource
		}
		logger.Log.Errorf(uiUtils.Error, errMsg)
	}
}

func cleanUpRestrictedMode(ctx context.Context, kubernetesProvider *kubernetes.Provider) []string {
	leftoverResources := make([]string, 0)
	if err := kubernetesProvider.RemoveDeployment(ctx, config.Config.MizuResourcesNamespace, kubernetes.ApiServerPodName); err != nil {
		resourceDesc := fmt.Sprintf("Deployment %s in namespace %s", kubernetes.ApiServerPodName, config.Config.MizuResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	}

	if err := kubernetesProvider.RemovePod(ctx, config.Config.MizuResourcesNamespace, kubernetes.ApiServerPodName); err != nil {
		resourceDesc := fmt.Sprintf("Pod %s in namespace %s", kubernetes.ApiServerPodName, config.Config.MizuResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	}

	if err := kubernetesProvider.RemoveService(ctx, config.Config.MizuResourcesNamespace, kubernetes.ApiServerPodName); err != nil {
		resourceDesc := fmt.Sprintf("Service %s in namespace %s", kubernetes.ApiServerPodName, config.Config.MizuResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	}

	if err := kubernetesProvider.RemoveDaemonSet(ctx, config.Config.MizuResourcesNamespace, kubernetes.TapperDaemonSetName); err != nil {
		resourceDesc := fmt.Sprintf("DaemonSet %s in namespace %s", kubernetes.TapperDaemonSetName, config.Config.MizuResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	}

	if err := kubernetesProvider.RemoveConfigMap(ctx, config.Config.MizuResourcesNamespace, kubernetes.ConfigMapName); err != nil {
		resourceDesc := fmt.Sprintf("ConfigMap %s in namespace %s", kubernetes.ConfigMapName, config.Config.MizuResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	}

	if err := kubernetesProvider.RemoveServicAccount(ctx, config.Config.MizuResourcesNamespace, kubernetes.ServiceAccountName); err != nil {
		resourceDesc := fmt.Sprintf("Service Account %s in namespace %s", kubernetes.ServiceAccountName, config.Config.MizuResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	}

	if err := kubernetesProvider.RemoveRole(ctx, config.Config.MizuResourcesNamespace, kubernetes.RoleName); err != nil {
		resourceDesc := fmt.Sprintf("Role %s in namespace %s", kubernetes.RoleName, config.Config.MizuResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	}

	if err := kubernetesProvider.RemoveRoleBinding(ctx, config.Config.MizuResourcesNamespace, kubernetes.RoleBindingName); err != nil {
		resourceDesc := fmt.Sprintf("RoleBinding %s in namespace %s", kubernetes.RoleBindingName, config.Config.MizuResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	}

	return leftoverResources
}

func cleanUpNonRestrictedMode(ctx context.Context, cancel context.CancelFunc, kubernetesProvider *kubernetes.Provider) []string {
	leftoverResources := make([]string, 0)

	if err := kubernetesProvider.RemoveNamespace(ctx, config.Config.MizuResourcesNamespace); err != nil {
		resourceDesc := fmt.Sprintf("Namespace %s", config.Config.MizuResourcesNamespace)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	} else {
		defer waitUntilNamespaceDeleted(ctx, cancel, kubernetesProvider)
	}

	if err := kubernetesProvider.RemoveClusterRole(ctx, kubernetes.ClusterRoleName); err != nil {
		resourceDesc := fmt.Sprintf("ClusterRole %s", kubernetes.ClusterRoleName)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	}

	if err := kubernetesProvider.RemoveClusterRoleBinding(ctx, kubernetes.ClusterRoleBindingName); err != nil {
		resourceDesc := fmt.Sprintf("ClusterRoleBinding %s", kubernetes.ClusterRoleBindingName)
		handleDeletionError(err, resourceDesc, &leftoverResources)
	}

	return leftoverResources
}
