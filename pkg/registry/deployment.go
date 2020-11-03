package registry

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"

	registryv1alpha1 "github.com/devfile/registry-operator/api/v1alpha1"
)

func GenerateDeployment(cr *registryv1alpha1.DevfileRegistry, scheme *runtime.Scheme) *appsv1.Deployment {
	ls := LabelsForDevfileRegistry(cr.Name)
	replicas := int32(1)

	dep := &appsv1.Deployment{
		ObjectMeta: generateObjectMeta(cr.Name, cr.Namespace, ls),
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: cr.Spec.BootstrapImage,
							Name:  "devfile-registry-bootstrap",
							Ports: []corev1.ContainerPort{{
								ContainerPort: 8080,
							}},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("64Mi"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("250m"),
									corev1.ResourceMemory: resource.MustParse("128Mi"),
								},
							},
							LivenessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/devfiles/index.json",
										Port: intstr.FromInt(8080),
									},
								},
								InitialDelaySeconds: int32(3),
								PeriodSeconds:       int32(3),
							},
							ReadinessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/devfiles/index.json",
										Port: intstr.FromInt(8080),
									},
								},
								InitialDelaySeconds: int32(3),
								PeriodSeconds:       int32(3),
							},
						},
						{
							Image: "registry:latest",
							Name:  "oci-registry",
							Ports: []corev1.ContainerPort{{
								ContainerPort: 5000,
							}},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("64Mi"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("500m"),
									corev1.ResourceMemory: resource.MustParse("256Mi"),
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "devfile-registry-storage",
									MountPath: "/var/lib/registry",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "devfile-registry-storage",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: PVCName(cr.Name),
								},
							},
						},
					},
				},
			},
		},
	}
	// Set Memcached instance as the owner and controller
	ctrl.SetControllerReference(cr, dep, scheme)
	return dep
}
