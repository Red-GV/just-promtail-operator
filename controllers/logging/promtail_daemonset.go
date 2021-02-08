package logging

import (
	corev1 "k8s.io/api/core/v1"
)

var (
	promtailImage = "docker.io/grafana/promtail:2.1.0"
)

func newDaemonSet() {
}

func newPodTemplateSpec() corev1.PodTemplateSpec {
	return corev1.PodTemplateSpec{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Image: promtailImage,
				},
			},
		},
	}
}
