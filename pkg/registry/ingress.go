package registry

import (
	registryv1alpha1 "github.com/devfile/registry-operator/api/v1alpha1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
)

func GenerateIngress(cr *registryv1alpha1.DevfileRegistry, host string, scheme *runtime.Scheme) *v1beta1.Ingress {
	ls := LabelsForDevfileRegistry(cr.Name)

	ingress := &v1beta1.Ingress{
		ObjectMeta: generateObjectMeta(IngressName(cr.Name), cr.Namespace, ls),
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{
				{
					Host: host,
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: []v1beta1.HTTPIngressPath{
								{
									Path: "/",
									Backend: v1beta1.IngressBackend{
										ServiceName: ServiceName(cr.Name),
										ServicePort: intstr.FromInt(int(DevfileIndexPort)),
									},
								},
								{
									Path: "/v2",
									Backend: v1beta1.IngressBackend{
										ServiceName: ServiceName(cr.Name),
										ServicePort: intstr.FromInt(int(OCIRegistryPort)),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Set DevfileRegistry instance as the owner and controller
	ctrl.SetControllerReference(cr, ingress, scheme)
	return ingress
}
