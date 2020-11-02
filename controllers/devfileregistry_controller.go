/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/discovery"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	registryv1alpha1 "github.com/devfile/registry-operator/api/v1alpha1"
)

// DevfileRegistryReconciler reconciles a DevfileRegistry object
type DevfileRegistryReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=registry.devfile.io,resources=devfileregistries,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=registry.devfile.io,resources=devfileregistries/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=registry.devfile.io,resources=devfileregistries/finalizers,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;
// +kubebuilder:rbac:groups=route.openshift.io,resources=routes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=route.openshift.io,resources=routes/custom-host,verbs=get;list;watch;create;update;patch;delete

func (r *DevfileRegistryReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("devfileregistry", req.NamespacedName)

	// your logic here
	// Fetch the DevfileRegistry instance
	devfileRegistry := &registryv1alpha1.DevfileRegistry{}
	err := r.Get(ctx, req.NamespacedName, devfileRegistry)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("DevfileRegistry resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get DevfileRegistry")
		return ctrl.Result{}, err
	}

	// Check if the service already exists, if not create a new one
	svcFound := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: devfileRegistry.Name, Namespace: devfileRegistry.Namespace}, svcFound)
	if err != nil && errors.IsNotFound(err) {
		// Define a new service
		pvc := r.generateDevfileRegistryService(devfileRegistry)
		log.Info("Creating a new Service", "Service.Namespace", pvc.Namespace, "Service.Name", pvc.Name)
		err = r.Create(ctx, pvc)
		if err != nil {
			log.Error(err, "Failed to create new Service", "Service.Namespace", pvc.Namespace, "Service.Name", pvc.Name)
			return ctrl.Result{}, err
		}
	} else if err != nil {
		log.Error(err, "Failed to get Service")
		return ctrl.Result{}, err
	}

	// Check if the persistentvolumeclaim already exists, if not create a new one
	pvcFound := &corev1.PersistentVolumeClaim{}
	err = r.Get(ctx, types.NamespacedName{Name: devfileRegistry.Name, Namespace: devfileRegistry.Namespace}, pvcFound)
	if err != nil && errors.IsNotFound(err) {
		// Define a new PVC
		pvc := r.generateDevfileRegistryPVC(devfileRegistry)
		log.Info("Creating a new PersistentVolumeClaim", "PersistentVolumeClaim.Namespace", pvc.Namespace, "PersistentVolumeClaim.Name", pvc.Name)
		err = r.Create(ctx, pvc)
		if err != nil {
			log.Error(err, "Failed to create new PersistentVolumeClaim", "PersistentVolumeClaim.Namespace", pvc.Namespace, "PersistentVolumeClaim.Name", pvc.Name)
			return ctrl.Result{}, err
		}
	} else if err != nil {
		log.Error(err, "Failed to get PersistentVolumeClaim")
		return ctrl.Result{}, err
	}

	// Check if the deployment already exists, if not create a new one
	depFound := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: devfileRegistry.Name, Namespace: devfileRegistry.Namespace}, depFound)
	if err != nil && errors.IsNotFound(err) {
		// Define a new Deployment
		dep := r.generateDevfileRegistryDeployment(devfileRegistry)
		log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.Create(ctx, dep)
		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		//return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	// Check if we're running on OpenShift
	isOS, err := IsOpenShift()
	if err != nil {
		return ctrl.Result{}, err
	}

	if isOS {
		// Check if the two routes already exist, if not create new ones -- new func
		hostname := devfileRegistry.Spec.IngressDomain
		devfilesRouteFound := &routev1.Route{}
		err = r.Get(ctx, types.NamespacedName{Name: devfileRegistry.Name + "-devfiles", Namespace: devfileRegistry.Namespace}, devfilesRouteFound)
		if err != nil && errors.IsNotFound(err) {
			// Define a new route exposing the devfile registry index
			route := r.generateDevfilesRoute(devfileRegistry, hostname)
			log.Info("Creating a new Route", "Route.Namespace", route.Namespace, "Route.Name", route.Name)
			err = r.Create(ctx, route)
			if err != nil {
				log.Error(err, "Failed to create new Route", "Route.Namespace", route.Namespace, "Route.Name", route.Name)
				return ctrl.Result{}, err
			}
		} else if err != nil {
			log.Error(err, "Failed to get Route")
			return ctrl.Result{}, err
		}

		ociRouteFound := &routev1.Route{}
		err = r.Get(ctx, types.NamespacedName{Name: devfileRegistry.Name + "-oci", Namespace: devfileRegistry.Namespace}, ociRouteFound)
		if err != nil && errors.IsNotFound(err) {
			// Define a new route exposing the devfile registry index
			route := r.generateOCIRoute(devfileRegistry, hostname)
			log.Info("Creating a new Route", "Route.Namespace", route.Namespace, "Route.Name", route.Name)
			err = r.Create(ctx, route)
			if err != nil {
				log.Error(err, "Failed to create new Route", "Route.Namespace", route.Namespace, "Route.Name", devfileRegistry.Name+"-oci")
				return ctrl.Result{}, err
			}
		} else if err != nil {
			log.Error(err, "Failed to get Route")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// generateDevfileRegistryDeployment returns a devfileregistry Deployment object
func (r *DevfileRegistryReconciler) generateDevfileRegistryDeployment(d *registryv1alpha1.DevfileRegistry) *appsv1.Deployment {
	ls := labelsForDevfileRegistry(d.Name)
	replicas := int32(1)

	dep := &appsv1.Deployment{
		ObjectMeta: generateObjectMeta(d.Name, d.Namespace, ls),
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
							Image: d.Spec.BootstrapImage,
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
							Name:  "devfile-registry",
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
									ClaimName: d.Name,
								},
							},
						},
					},
				},
			},
		},
	}
	// Set Memcached instance as the owner and controller
	ctrl.SetControllerReference(d, dep, r.Scheme)
	return dep
}

// generateDevfileRegistryService returns a devfileregistry Service object
func (r *DevfileRegistryReconciler) generateDevfileRegistryService(d *registryv1alpha1.DevfileRegistry) *corev1.Service {
	ls := labelsForDevfileRegistry(d.Name)

	svc := &corev1.Service{
		ObjectMeta: generateObjectMeta(d.Name, d.Namespace, ls),
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name: "devfile-registry-metadata",
					Port: int32(8080),
				},
				{
					Name: "devfile-registry",
					Port: int32(5000),
				},
			},
			Selector: ls,
		},
	}

	// Set DevfileRegistry instance as the owner and controller
	ctrl.SetControllerReference(d, svc, r.Scheme)
	return svc
}

// generateDevfileRegistryPVC returns a devfileregistry Service object
func (r *DevfileRegistryReconciler) generateDevfileRegistryPVC(d *registryv1alpha1.DevfileRegistry) *corev1.PersistentVolumeClaim {
	ls := labelsForDevfileRegistry(d.Name)

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: generateObjectMeta(d.Name, d.Namespace, ls),
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("10Gi"),
				},
			},
		},
	}

	// Set DevfileRegistry instance as the owner and controller
	ctrl.SetControllerReference(d, pvc, r.Scheme)
	return pvc
}

// generateDevfileRegistryPVC returns a devfileregistry Service object
func (r *DevfileRegistryReconciler) generateDevfilesRoute(d *registryv1alpha1.DevfileRegistry, host string) *routev1.Route {

	ls := labelsForDevfileRegistry(d.Name)
	weight := int32(100)

	route := &routev1.Route{
		ObjectMeta: generateObjectMeta(d.Name+"-devfiles", d.Namespace, ls),
		Spec: routev1.RouteSpec{
			To: routev1.RouteTargetReference{
				Kind:   "Service",
				Name:   d.Name,
				Weight: &weight,
			},
			Port: &routev1.RoutePort{
				TargetPort: intstr.FromString("devfile-registry-metadata"),
			},
			Path: "/",
		},
	}

	if host != "" {
		route.Spec.Host = host
	}
	// Set DevfileRegistry instance as the owner and controller
	ctrl.SetControllerReference(d, route, r.Scheme)
	return route
}

// generateDevfileRegistryPVC returns a devfileregistry Service object
func (r *DevfileRegistryReconciler) generateOCIRoute(d *registryv1alpha1.DevfileRegistry, host string) *routev1.Route {
	ls := labelsForDevfileRegistry(d.Name)
	weight := int32(100)

	route := &routev1.Route{
		ObjectMeta: generateObjectMeta(d.Name+"-oci", d.Namespace, ls),
		Spec: routev1.RouteSpec{
			Host: host,
			To: routev1.RouteTargetReference{
				Kind:   "Service",
				Name:   d.Name,
				Weight: &weight,
			},
			Port: &routev1.RoutePort{
				TargetPort: intstr.FromString("devfile-registry"),
			},
			Path: "/",
		},
	}

	if host != "" {
		route.Spec.Host = host
	}

	// Set DevfileRegistry instance as the owner and controller
	ctrl.SetControllerReference(d, route, r.Scheme)
	return route
}

func generateObjectMeta(name string, namespace string, labels map[string]string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      name,
		Namespace: namespace,
		Labels:    labels,
	}
}

func IsOpenShift() (bool, error) {
	kubeCfg, err := config.GetConfig()
	if err != nil {
		return false, err
	}
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(kubeCfg)
	if err != nil {
		return false, err
	}
	apiList, err := discoveryClient.ServerGroups()
	if err != nil {
		return false, err
	}
	if findAPIGroup(apiList.Groups, "route.openshift.io") == nil {
		return false, nil
	} else {
		return true, nil
	}
}

func findAPIGroup(source []metav1.APIGroup, apiName string) *metav1.APIGroup {
	for i := 0; i < len(source); i++ {
		if source[i].Name == apiName {
			return &source[i]
		}
	}
	return nil
}

// labelsForDevfileRegistry returns the labels for selecting the resources
// belonging to the given memcached CR name.
func labelsForDevfileRegistry(name string) map[string]string {
	return map[string]string{"app": "devfileregistry", "devfileregistry_cr": name}
}

func (r *DevfileRegistryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&registryv1alpha1.DevfileRegistry{}).
		Complete(r)
}
