package registry

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	registryv1alpha1 "github.com/devfile/registry-operator/api/v1alpha1"
)

// GenerateDevfileRegistryPVC returns a PVC for providing storage on the OCI registry container
func GeneratePVC(cr *registryv1alpha1.DevfileRegistry, scheme *runtime.Scheme) *corev1.PersistentVolumeClaim {
	ls := LabelsForDevfileRegistry(cr.Name)

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: generateObjectMeta(cr.Name, cr.Namespace, ls),
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(getDevfileRegistryVolumeSize(cr)),
				},
			},
		},
	}

	// Set DevfileRegistry instance as the owner and controller
	ctrl.SetControllerReference(cr, pvc, scheme)
	return pvc
}
