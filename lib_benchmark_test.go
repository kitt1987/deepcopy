package deepcopy_test

import (
	"github.com/kitt1987/deepcopy"
	"gotest.tools/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func BenchmarkDeepPartialCopy(b *testing.B) {
	justFalse := false
	justDigit := int64(10)

	src := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "podA",
			Namespace: "namespaceA",
			Labels: map[string]string{
				"A": "B",
			},
		},
		Spec: v1.PodSpec{
			InitContainers: []v1.Container{
				{
					Name: "containerA",
					LivenessProbe: &v1.Probe{
						InitialDelaySeconds: 121,
					},
					Ports: []v1.ContainerPort{
						{
							Name:     "port-81",
							HostIP:   "1.1.1.1",
							HostPort: 81,
						},
					},
				},
				{
					Name: "containerB",
					LivenessProbe: &v1.Probe{
						InitialDelaySeconds: 122,
					},
					Ports: []v1.ContainerPort{
						{
							Name:     "port-80",
							HostIP:   "2.2.2.2",
							HostPort: 80,
						},
						{
							Name:     "port-32767",
							HostIP:   "2.2.2.2",
							HostPort: 32767,
						},
					},
				},
				{
					Name: "containerC",
					LivenessProbe: &v1.Probe{
						InitialDelaySeconds: 125,
					},
					Ports: []v1.ContainerPort{
						{
							Name:     "port-82",
							HostIP:   "3.3.3.3",
							HostPort: 82,
						},
					},
				},
			},
			SecurityContext: &v1.PodSecurityContext{
				RunAsNonRoot: &justFalse,
				RunAsUser:    &justDigit,
			},
		},
	}

	for n := 0; n < b.N; n++ {
		var dst v1.Pod
		assert.Assert(b, deepcopy.Partial(&dst, &src,
			"ObjectMeta.Name",
			"ObjectMeta.Labels",
			"Spec.SecurityContext.RunAsNonRoot",
			"Spec.InitContainers.Ports.Name",
			"Spec.InitContainers.Ports.HostPort"))
	}
}

func BenchmarkDeepPartialCopyWithReplicator(b *testing.B) {
	justFalse := false
	justDigit := int64(10)

	src := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "podA",
			Namespace: "namespaceA",
			Labels: map[string]string{
				"A": "B",
			},
		},
		Spec: v1.PodSpec{
			InitContainers: []v1.Container{
				{
					Name: "containerA",
					LivenessProbe: &v1.Probe{
						InitialDelaySeconds: 121,
					},
					Ports: []v1.ContainerPort{
						{
							Name:     "port-81",
							HostIP:   "1.1.1.1",
							HostPort: 81,
						},
					},
				},
				{
					Name: "containerB",
					LivenessProbe: &v1.Probe{
						InitialDelaySeconds: 122,
					},
					Ports: []v1.ContainerPort{
						{
							Name:     "port-80",
							HostIP:   "2.2.2.2",
							HostPort: 80,
						},
						{
							Name:     "port-32767",
							HostIP:   "2.2.2.2",
							HostPort: 32767,
						},
					},
				},
				{
					Name: "containerC",
					LivenessProbe: &v1.Probe{
						InitialDelaySeconds: 125,
					},
					Ports: []v1.ContainerPort{
						{
							Name:     "port-82",
							HostIP:   "3.3.3.3",
							HostPort: 82,
						},
					},
				},
			},
			SecurityContext: &v1.PodSecurityContext{
				RunAsNonRoot: &justFalse,
				RunAsUser:    &justDigit,
			},
		},
	}

	replicator := deepcopy.NewPartialReplicator("ObjectMeta.Name",
		"ObjectMeta.Labels",
		"Spec.SecurityContext.RunAsNonRoot",
		"Spec.InitContainers.Ports.Name",
		"Spec.InitContainers.Ports.HostPort")
	for n := 0; n < b.N; n++ {
		var dst v1.Pod
		assert.Assert(b, replicator.Copy(&dst, &src))
	}
}
