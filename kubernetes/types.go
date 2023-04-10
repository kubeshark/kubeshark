package kubernetes

import (
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	applyconfapp "k8s.io/client-go/applyconfigurations/apps/v1"
	applyconfcore "k8s.io/client-go/applyconfigurations/core/v1"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
	applyconfmeta "k8s.io/client-go/applyconfigurations/meta/v1"
)

type DaemonSetSpec struct {
	Selector metav1.LabelSelector `json:"selector,omitempty" protobuf:"bytes,1,opt,name=selector"`
	Template core.Pod             `json:"template,omitempty" protobuf:"bytes,2,opt,name=template"`
}

type DaemonSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Spec              DaemonSetSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}

func (d *DaemonSet) GenerateApplyConfiguration(name string, namespace string, podName string, provider *Provider) *applyconfapp.DaemonSetApplyConfiguration {
	// Pod
	p := d.Spec.Template.Spec
	podSpec := applyconfcore.PodSpec()
	podSpec.WithHostNetwork(p.HostNetwork)
	podSpec.WithDNSPolicy(p.DNSPolicy)
	podSpec.WithTerminationGracePeriodSeconds(*p.TerminationGracePeriodSeconds)
	podSpec.WithServiceAccountName(p.ServiceAccountName)

	// Containers
	for _, c := range d.Spec.Template.Spec.Containers {
		// Common
		container := applyconfcore.Container()
		container.WithName(c.Name)
		container.WithImage(c.Image)
		container.WithImagePullPolicy(c.ImagePullPolicy)
		container.WithCommand(c.Command...)

		// Linux capabilities
		caps := applyconfcore.Capabilities().WithAdd(c.SecurityContext.Capabilities.Add...).WithDrop(c.SecurityContext.Capabilities.Drop...)
		container.WithSecurityContext(applyconfcore.SecurityContext().WithCapabilities(caps))

		// Environment variables
		var envvars []*v1.EnvVarApplyConfiguration
		for _, e := range c.Env {
			envvars = append(envvars, applyconfcore.EnvVar().WithName(e.Name).WithValue(e.Value))
		}
		container.WithEnv(envvars...)

		// Resource limits
		resources := applyconfcore.ResourceRequirements().WithRequests(c.Resources.Requests).WithLimits(c.Resources.Limits)
		container.WithResources(resources)

		// Volume mounts
		for _, m := range c.VolumeMounts {
			volumeMount := applyconfcore.VolumeMount().WithName(m.Name).WithMountPath(m.MountPath).WithReadOnly(m.ReadOnly)
			container.WithVolumeMounts(volumeMount)
		}

		podSpec.WithContainers(container)
	}

	// Node affinity (RequiredDuringSchedulingIgnoredDuringExecution only)
	if p.Affinity != nil {
		nodeSelector := applyconfcore.NodeSelector()
		for _, term := range p.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {
			nodeSelectorTerm := applyconfcore.NodeSelectorTerm()
			for _, selector := range term.MatchExpressions {
				nodeSelectorRequirement := applyconfcore.NodeSelectorRequirement()
				nodeSelectorRequirement.WithKey(selector.Key)
				nodeSelectorRequirement.WithOperator(selector.Operator)
				nodeSelectorRequirement.WithValues(selector.Values...)
				nodeSelectorTerm.WithMatchExpressions(nodeSelectorRequirement)
			}
			for _, selector := range term.MatchFields {
				nodeSelectorRequirement := applyconfcore.NodeSelectorRequirement()
				nodeSelectorRequirement.WithKey(selector.Key)
				nodeSelectorRequirement.WithOperator(selector.Operator)
				nodeSelectorRequirement.WithValues(selector.Values...)
				nodeSelectorTerm.WithMatchFields(nodeSelectorRequirement)
			}
			nodeSelector.WithNodeSelectorTerms(nodeSelectorTerm)
		}
		nodeAffinity := applyconfcore.NodeAffinity()
		nodeAffinity.WithRequiredDuringSchedulingIgnoredDuringExecution(nodeSelector)
		affinity := applyconfcore.Affinity()
		affinity.WithNodeAffinity(nodeAffinity)
		podSpec.WithAffinity(affinity)
	}

	// Tolerations
	for _, t := range p.Tolerations {
		toleration := applyconfcore.Toleration()
		toleration.WithKey(t.Key)
		toleration.WithOperator(t.Operator)
		toleration.WithValue(t.Value)
		toleration.WithEffect(t.Effect)
		if t.TolerationSeconds != nil {
			toleration.WithTolerationSeconds(*t.TolerationSeconds)
		}
		podSpec.WithTolerations(toleration)
	}

	// Volumes
	for _, v := range p.Volumes {
		volume := applyconfcore.Volume()
		volume.WithName(v.Name).WithHostPath(applyconfcore.HostPathVolumeSource().WithPath(v.HostPath.Path))
		podSpec.WithVolumes(volume)
	}

	// Image pull secrets
	if len(p.ImagePullSecrets) > 0 {
		localObjectReference := applyconfcore.LocalObjectReference()
		for _, o := range p.ImagePullSecrets {
			localObjectReference.WithName(o.Name)
		}
		podSpec.WithImagePullSecrets(localObjectReference)
	}

	podTemplate := applyconfcore.PodTemplateSpec()
	podTemplate.WithLabels(buildWithDefaultLabels(map[string]string{
		"app": podName,
	}, provider))
	podTemplate.WithSpec(podSpec)

	labelSelector := applyconfmeta.LabelSelector()
	labelSelector.WithMatchLabels(map[string]string{"app": podName})

	daemonSet := applyconfapp.DaemonSet(name, namespace)
	daemonSet.
		WithLabels(buildWithDefaultLabels(map[string]string{}, provider)).
		WithSpec(applyconfapp.DaemonSetSpec().WithSelector(labelSelector).WithTemplate(podTemplate))

	return daemonSet
}
