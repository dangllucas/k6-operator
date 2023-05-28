package containers

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"

	resource "k8s.io/apimachinery/pkg/api/resource"
)

// NewS3Container is used to download a script archive from S3.
func NewS3Container(uri, image string, volumeMount corev1.VolumeMount, command []string, env []corev1.EnvVar) corev1.Container {
	return corev1.Container{
		Name:  "archive-download",
		Image: image,
		Env:   env,
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    *resource.NewMilliQuantity(50, resource.DecimalSI),
				corev1.ResourceMemory: *resource.NewQuantity(2097152, resource.BinarySI),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    *resource.NewMilliQuantity(100, resource.DecimalSI),
				corev1.ResourceMemory: *resource.NewQuantity(209715200, resource.BinarySI),
			},
		},
		Command: append(
			command,
			fmt.Sprintf("curl -X GET -L '%s' > /test/archive.tar ; ls -l /test", uri),
		),
		VolumeMounts: []corev1.VolumeMount{volumeMount},
	}
}
