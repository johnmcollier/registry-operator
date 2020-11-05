package registry

import (
	routev1 "github.com/openshift/api/route/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"

	registryv1alpha1 "github.com/devfile/registry-operator/api/v1alpha1"
)

// GenerateDevfilesRoute returns a route exposing the devfile registry index
func GenerateDevfilesRoute(cr *registryv1alpha1.DevfileRegistry, host string, scheme *runtime.Scheme) *routev1.Route {

	ls := LabelsForDevfileRegistry(cr.Name)
	weight := int32(100)

	route := &routev1.Route{
		ObjectMeta: generateObjectMeta(DevfilesRouteName(cr.Name), cr.Namespace, ls),
		Spec: routev1.RouteSpec{
			To: routev1.RouteTargetReference{
				Kind:   "Service",
				Name:   ServiceName(cr.Name),
				Weight: &weight,
			},
			Port: &routev1.RoutePort{
				TargetPort: intstr.FromString(DevfileIndexPortName),
			},
			Path: "/",
		},
	}

	if host != "" {
		route.Spec.Host = host
	}
	// Set DevfileRegistry instance as the owner and controller
	ctrl.SetControllerReference(cr, route, scheme)
	return route
}

// GenerateOCIRoute returns a route object for the OCI registry server
func GenerateOCIRoute(cr *registryv1alpha1.DevfileRegistry, host string, scheme *runtime.Scheme) *routev1.Route {
	ls := LabelsForDevfileRegistry(cr.Name)
	weight := int32(100)

	route := &routev1.Route{
		ObjectMeta: generateObjectMeta(OCIRouteName(cr.Name), cr.Namespace, ls),
		Spec: routev1.RouteSpec{
			Host: host,
			To: routev1.RouteTargetReference{
				Kind:   "Service",
				Name:   ServiceName(cr.Name),
				Weight: &weight,
			},
			Port: &routev1.RoutePort{
				TargetPort: intstr.FromString(OCIRegistryPortName),
			},
			Path: "/v2",
		},
	}

	if host != "" {
		route.Spec.Host = host
	}

	// Set DevfileRegistry instance as the owner and controller
	ctrl.SetControllerReference(cr, route, scheme)
	return route
}
