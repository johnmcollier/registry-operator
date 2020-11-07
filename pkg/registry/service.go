package registry

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	registryv1alpha1 "github.com/devfile/registry-operator/api/v1alpha1"
)

// GenerateDevfileRegistryService returns a devfileregistry Service object
func GenerateService(cr *registryv1alpha1.DevfileRegistry, scheme *runtime.Scheme) *corev1.Service {
	ls := LabelsForDevfileRegistry(cr.Name)

	svc := &corev1.Service{
		ObjectMeta: generateObjectMeta(ServiceName(cr.Name), cr.Namespace, ls),
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name: DevfileIndexPortName,
					Port: DevfileIndexPort,
				},
				{
					Name: OCIRegistryPortName,
					Port: OCIRegistryPort,
				},
			},
			Selector: ls,
		},
	}

	// Set DevfileRegistry instance as the owner and controller
	ctrl.SetControllerReference(cr, svc, scheme)
	return svc
}
